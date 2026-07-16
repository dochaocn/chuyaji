package handler

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/dochaocn/chuyaji/services/api/internal/middleware"
	"github.com/dochaocn/chuyaji/services/api/internal/model"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type familyMemberOut struct {
	UserID    uint64    `json:"user_id"`
	Nickname  string    `json:"nickname"`
	AvatarURL string    `json:"avatar_url,omitempty"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

type familyOut struct {
	ID        uint64            `json:"id"`
	Name      string            `json:"name"`
	CreatedBy uint64            `json:"created_by"`
	MyUserID  uint64            `json:"my_user_id"`
	MyRole    string            `json:"my_role"`
	Members   []familyMemberOut `json:"members"`
}

func (h *Handler) currentFamilyMembership(uid uint64) (*model.Family, *model.FamilyMember, error) {
	var member model.FamilyMember
	if err := h.DB.Where("user_id = ?", uid).First(&member).Error; err != nil {
		return nil, nil, err
	}
	var fam model.Family
	if err := h.DB.First(&fam, member.FamilyID).Error; err != nil {
		return nil, nil, err
	}
	return &fam, &member, nil
}

func (h *Handler) loadFamilyOut(fam *model.Family, myUserID uint64, myRole string) (familyOut, error) {
	var members []model.FamilyMember
	if err := h.DB.Where("family_id = ?", fam.ID).Order("id ASC").Find(&members).Error; err != nil {
		return familyOut{}, err
	}
	userIDs := make([]uint64, 0, len(members))
	for _, m := range members {
		userIDs = append(userIDs, m.UserID)
	}
	var users []model.User
	if len(userIDs) > 0 {
		h.DB.Where("id IN ?", userIDs).Find(&users)
	}
	userMap := make(map[uint64]model.User, len(users))
	for _, u := range users {
		userMap[u.ID] = u
	}
	outMembers := make([]familyMemberOut, 0, len(members))
	for _, m := range members {
		u := userMap[m.UserID]
		outMembers = append(outMembers, familyMemberOut{
			UserID:    m.UserID,
			Nickname:  u.Nickname,
			AvatarURL: u.AvatarURL,
			Role:      m.Role,
			CreatedAt: m.CreatedAt,
		})
	}
	return familyOut{
		ID:        fam.ID,
		Name:      fam.Name,
		CreatedBy: fam.CreatedBy,
		MyUserID:  myUserID,
		MyRole:    myRole,
		Members:   outMembers,
	}, nil
}

func (h *Handler) GetCurrentFamily(c *gin.Context) {
	uid, ok := middleware.UserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	fam, member, err := h.currentFamilyMembership(uid)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "no family"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "query"})
		return
	}
	out, err := h.loadFamilyOut(fam, uid, member.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "query"})
		return
	}
	c.JSON(http.StatusOK, out)
}

func (h *Handler) CreateCurrentFamily(c *gin.Context) {
	uid, ok := middleware.UserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	if _, _, err := h.currentFamilyMembership(uid); err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "already has family"})
		return
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "query"})
		return
	}
	familyID, err := h.createUserFamily(uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "family"})
		return
	}
	var fam model.Family
	if err := h.DB.First(&fam, familyID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "query"})
		return
	}
	out, err := h.loadFamilyOut(&fam, uid, model.FamilyRoleOwner)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "query"})
		return
	}
	c.JSON(http.StatusCreated, out)
}

type patchFamilyReq struct {
	Name *string `json:"name" binding:"omitempty,max=64"`
}

func (h *Handler) PatchCurrentFamily(c *gin.Context) {
	uid, ok := middleware.UserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	fam, member, err := h.currentFamilyMembership(uid)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "no family"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "query"})
		return
	}
	if member.Role != model.FamilyRoleOwner {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}
	var req patchFamilyReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}
	if req.Name != nil {
		fam.Name = *req.Name
	}
	if err := h.DB.Save(fam).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "save"})
		return
	}
	out, err := h.loadFamilyOut(fam, uid, member.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "query"})
		return
	}
	c.JSON(http.StatusOK, out)
}

type createInviteReq struct {
	Role string `json:"role" binding:"required,oneof=write read"`
}

func randomInviteToken() (string, error) {
	b := make([]byte, 24)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func (h *Handler) CreateFamilyInvite(c *gin.Context) {
	uid, ok := middleware.UserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	fam, member, err := h.currentFamilyMembership(uid)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "no family"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "query"})
		return
	}
	if member.Role != model.FamilyRoleOwner {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}
	var req createInviteReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}
	token, err := randomInviteToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "token"})
		return
	}
	inv := model.FamilyInvite{
		Token:     token,
		FamilyID:  fam.ID,
		Role:      req.Role,
		InviterID: uid,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}
	if err := h.DB.Create(&inv).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "create"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"token":      inv.Token,
		"path":       "/pages/family/join?token=" + inv.Token,
		"role":       inv.Role,
		"expires_at": inv.ExpiresAt,
	})
}

type familyInvitePreviewOut struct {
	FamilyName      string             `json:"family_name"`
	InviterNickname string             `json:"inviter_nickname"`
	Role            string             `json:"role"`
	ExpiresAt       time.Time          `json:"expires_at"`
	MemberCount     int64              `json:"member_count"`
	AlreadyMember   bool               `json:"already_member"`
	Conflict        *familyConflictOut `json:"conflict,omitempty"`
}

func (h *Handler) loadInviteByToken(token string) (*model.FamilyInvite, error) {
	var inv model.FamilyInvite
	if err := h.DB.Where("token = ?", token).First(&inv).Error; err != nil {
		return nil, err
	}
	return &inv, nil
}

func inviteStatus(inv *model.FamilyInvite, now time.Time) string {
	if inv.RevokedAt != nil {
		return "revoked"
	}
	if inv.AcceptedAt != nil {
		return "accepted"
	}
	if now.After(inv.ExpiresAt) {
		return "expired"
	}
	return "pending"
}

func (h *Handler) PreviewFamilyInvite(c *gin.Context) {
	uid, ok := middleware.UserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	token := c.Param("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad token"})
		return
	}
	inv, err := h.loadInviteByToken(token)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "invite not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "query"})
		return
	}

	var fam model.Family
	if err := h.DB.First(&fam, inv.FamilyID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "query"})
		return
	}
	var inviter model.User
	_ = h.DB.First(&inviter, inv.InviterID).Error
	var memberCount int64
	_ = h.DB.Model(&model.FamilyMember{}).Where("family_id = ?", fam.ID).Count(&memberCount).Error

	_, alreadyMember, _ := h.resolveFamilyRole(uid, inv.FamilyID)
	if alreadyMember {
		c.JSON(http.StatusOK, familyInvitePreviewOut{
			FamilyName:      fam.Name,
			InviterNickname: inviter.Nickname,
			Role:            inv.Role,
			ExpiresAt:       inv.ExpiresAt,
			MemberCount:     memberCount,
			AlreadyMember:   true,
		})
		return
	}

	status := inviteStatus(inv, time.Now())
	if status == "revoked" {
		c.JSON(http.StatusGone, gin.H{"error": "invite revoked"})
		return
	}
	if status == "expired" {
		c.JSON(http.StatusGone, gin.H{"error": "invite expired"})
		return
	}
	if status == "accepted" {
		c.JSON(http.StatusConflict, gin.H{"error": "invite already used"})
		return
	}

	conflict, err := h.buildFamilyConflict(uid, inv.FamilyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "query"})
		return
	}

	c.JSON(http.StatusOK, familyInvitePreviewOut{
		FamilyName:      fam.Name,
		InviterNickname: inviter.Nickname,
		Role:            inv.Role,
		ExpiresAt:       inv.ExpiresAt,
		MemberCount:     memberCount,
		AlreadyMember:   false,
		Conflict:        conflict,
	})
}

func (h *Handler) AcceptFamilyInvite(c *gin.Context) {
	uid, ok := middleware.UserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	token := c.Param("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad token"})
		return
	}
	var inv model.FamilyInvite
	if err := h.DB.Where("token = ?", token).First(&inv).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "invite not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "query"})
		return
	}

	// Already member of this family → idempotent success (even if invite was consumed).
	if role, okRole, err := h.resolveFamilyRole(uid, inv.FamilyID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "query"})
		return
	} else if okRole {
		var fam model.Family
		_ = h.DB.First(&fam, inv.FamilyID)
		out, _ := h.loadFamilyOut(&fam, uid, role)
		c.JSON(http.StatusOK, out)
		return
	}

	if inv.RevokedAt != nil {
		c.JSON(http.StatusGone, gin.H{"error": "invite revoked"})
		return
	}
	if inv.AcceptedAt != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "invite already used"})
		return
	}
	if time.Now().After(inv.ExpiresAt) {
		c.JSON(http.StatusGone, gin.H{"error": "invite expired"})
		return
	}

	var existing model.FamilyMember
	hasExisting := false
	err := h.DB.Where("user_id = ?", uid).First(&existing).Error
	if err == nil {
		hasExisting = existing.FamilyID != inv.FamilyID
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "query"})
		return
	}

	if hasExisting {
		conflict, cErr := h.buildFamilyConflict(uid, inv.FamilyID)
		if cErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "query"})
			return
		}
		if conflict != nil && conflict.HasOtherMembers {
			c.JSON(http.StatusConflict, gin.H{"error": "family has other members", "conflict": conflict})
			return
		}
	}

	now := time.Now()
	if err := h.DB.Transaction(func(tx *gorm.DB) error {
		if hasExisting {
			if err := detachUserFromFamilyForJoin(tx, uid, &existing); err != nil {
				return err
			}
		}
		m := model.FamilyMember{
			FamilyID: inv.FamilyID,
			UserID:   uid,
			Role:     inv.Role,
		}
		if err := tx.Create(&m).Error; err != nil {
			return err
		}
		inv.AcceptedBy = &uid
		inv.AcceptedAt = &now
		return tx.Save(&inv).Error
	}); err != nil {
		if errors.Is(err, errFamilyHasOtherMembers) {
			c.JSON(http.StatusConflict, gin.H{"error": "family has other members"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "accept"})
		return
	}

	var fam model.Family
	_ = h.DB.First(&fam, inv.FamilyID)
	out, err := h.loadFamilyOut(&fam, uid, inv.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "query"})
		return
	}
	c.JSON(http.StatusOK, out)
}

func (h *Handler) LeaveCurrentFamily(c *gin.Context) {
	uid, ok := middleware.UserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	fam, member, err := h.currentFamilyMembership(uid)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "no family"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "query"})
		return
	}
	var others int64
	if err := h.DB.Model(&model.FamilyMember{}).Where("family_id = ? AND user_id != ?", fam.ID, uid).Count(&others).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "query"})
		return
	}
	if others > 0 && member.Role == model.FamilyRoleOwner {
		c.JSON(http.StatusConflict, gin.H{"error": "family has other members"})
		return
	}
	if err := h.DB.Transaction(func(tx *gorm.DB) error {
		if others == 0 {
			return purgeFamilyData(tx, fam.ID)
		}
		return tx.Delete(member).Error
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "leave"})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) TransferFamilyOwner(c *gin.Context) {
	uid, ok := middleware.UserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	targetID, err := strconv.ParseUint(c.Param("userId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad user id"})
		return
	}
	if targetID == uid {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot transfer to self"})
		return
	}
	fam, member, err := h.currentFamilyMembership(uid)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "no family"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "query"})
		return
	}
	if member.Role != model.FamilyRoleOwner {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}
	var target model.FamilyMember
	if err := h.DB.Where("family_id = ? AND user_id = ?", fam.ID, targetID).First(&target).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "member not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "query"})
		return
	}
	if err := h.DB.Transaction(func(tx *gorm.DB) error {
		target.Role = model.FamilyRoleOwner
		if err := tx.Save(&target).Error; err != nil {
			return err
		}
		member.Role = model.FamilyRoleWrite
		return tx.Save(member).Error
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "transfer"})
		return
	}
	out, err := h.loadFamilyOut(fam, uid, model.FamilyRoleWrite)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "query"})
		return
	}
	c.JSON(http.StatusOK, out)
}

type patchMemberReq struct {
	Role string `json:"role" binding:"required,oneof=owner write read"`
}

func (h *Handler) PatchFamilyMember(c *gin.Context) {
	uid, ok := middleware.UserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	targetID, err := strconv.ParseUint(c.Param("userId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad user id"})
		return
	}
	fam, member, err := h.currentFamilyMembership(uid)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "no family"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "query"})
		return
	}
	if member.Role != model.FamilyRoleOwner {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}
	if targetID == uid {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot change own role"})
		return
	}
	var req patchMemberReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}
	var target model.FamilyMember
	if err := h.DB.Where("family_id = ? AND user_id = ?", fam.ID, targetID).First(&target).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "member not found"})
		return
	}
	if req.Role == model.FamilyRoleOwner {
		c.JSON(http.StatusBadRequest, gin.H{"error": "use transfer-owner to assign owner"})
		return
	}
	target.Role = req.Role
	if err := h.DB.Save(&target).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "save"})
		return
	}
	out, err := h.loadFamilyOut(fam, uid, member.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "query"})
		return
	}
	c.JSON(http.StatusOK, out)
}

func (h *Handler) DeleteFamilyMember(c *gin.Context) {
	uid, ok := middleware.UserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	targetID, err := strconv.ParseUint(c.Param("userId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad user id"})
		return
	}
	fam, member, err := h.currentFamilyMembership(uid)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "no family"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "query"})
		return
	}
	if member.Role != model.FamilyRoleOwner {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}
	if targetID == uid {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot remove yourself"})
		return
	}
	res := h.DB.Where("family_id = ? AND user_id = ?", fam.ID, targetID).Delete(&model.FamilyMember{})
	if res.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "delete"})
		return
	}
	if res.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "member not found"})
		return
	}
	c.Status(http.StatusNoContent)
}

type patchMemberNicknameReq struct {
	Nickname string `json:"nickname" binding:"required,max=64"`
}

func (h *Handler) PatchFamilyMemberNickname(c *gin.Context) {
	uid, ok := middleware.UserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	targetID, err := strconv.ParseUint(c.Param("userId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad user id"})
		return
	}
	fam, member, err := h.currentFamilyMembership(uid)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "no family"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "query"})
		return
	}
	var targetMember model.FamilyMember
	if err := h.DB.Where("family_id = ? AND user_id = ?", fam.ID, targetID).First(&targetMember).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "member not found"})
		return
	}
	var req patchMemberNicknameReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}
	nickname := strings.TrimSpace(req.Nickname)
	if nickname == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "nickname required"})
		return
	}
	var u model.User
	if err := h.DB.First(&u, targetID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	u.Nickname = nickname
	if err := h.DB.Save(&u).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "save"})
		return
	}
	out, err := h.loadFamilyOut(fam, uid, member.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "query"})
		return
	}
	c.JSON(http.StatusOK, out)
}
