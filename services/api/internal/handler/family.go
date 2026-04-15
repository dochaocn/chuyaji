package handler

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strconv"

	"github.com/dochaocn/chuyaji/services/api/internal/middleware"
	"github.com/dochaocn/chuyaji/services/api/internal/model"
	"github.com/gin-gonic/gin"
)

type familyOut struct {
	ID         uint64 `json:"id"`
	Name       string `json:"name"`
	InviteCode string `json:"invite_code"`
}

func randomInviteCode() (string, error) {
	b := make([]byte, 4)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func (h *Handler) ListFamilies(c *gin.Context) {
	uid, ok := middleware.UserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	var fams []model.Family
	err := h.DB.Model(&model.Family{}).
		Joins("JOIN family_members ON family_members.family_id = families.id").
		Where("family_members.user_id = ?", uid).
		Find(&fams).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "query"})
		return
	}
	out := make([]familyOut, 0, len(fams))
	for _, f := range fams {
		out = append(out, familyOut{ID: f.ID, Name: f.Name, InviteCode: f.InviteCode})
	}
	c.JSON(http.StatusOK, gin.H{"items": out})
}

type createFamilyReq struct {
	Name string `json:"name" binding:"required,max=64"`
}

func (h *Handler) CreateFamily(c *gin.Context) {
	uid, ok := middleware.UserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	var req createFamilyReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}
	code, err := randomInviteCode()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invite code"})
		return
	}
	f := model.Family{Name: req.Name, InviteCode: code}
	if err := h.DB.Create(&f).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "create family"})
		return
	}
	m := model.FamilyMember{FamilyID: f.ID, UserID: uid, Role: "owner"}
	if err := h.DB.Create(&m).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "create membership"})
		return
	}
	c.JSON(http.StatusCreated, familyOut{ID: f.ID, Name: f.Name, InviteCode: f.InviteCode})
}

type joinFamilyReq struct {
	InviteCode string `json:"invite_code" binding:"required"`
}

func (h *Handler) JoinFamily(c *gin.Context) {
	uid, ok := middleware.UserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	var req joinFamilyReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}
	var f model.Family
	if err := h.DB.Where("invite_code = ?", req.InviteCode).First(&f).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "family not found"})
		return
	}
	var n int64
	h.DB.Model(&model.FamilyMember{}).Where("family_id = ? AND user_id = ?", f.ID, uid).Count(&n)
	if n > 0 {
		c.JSON(http.StatusOK, familyOut{ID: f.ID, Name: f.Name, InviteCode: f.InviteCode})
		return
	}
	m := model.FamilyMember{FamilyID: f.ID, UserID: uid, Role: "member"}
	if err := h.DB.Create(&m).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "join"})
		return
	}
	c.JSON(http.StatusOK, familyOut{ID: f.ID, Name: f.Name, InviteCode: f.InviteCode})
}

func (h *Handler) GetFamily(c *gin.Context) {
	uid, ok := middleware.UserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad id"})
		return
	}
	ok2, err := h.isFamilyMember(uid, id)
	if err != nil || !ok2 {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}
	var f model.Family
	if err := h.DB.First(&f, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, familyOut{ID: f.ID, Name: f.Name, InviteCode: f.InviteCode})
}
