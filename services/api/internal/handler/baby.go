package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/dochaocn/chuyaji/services/api/internal/middleware"
	"github.com/dochaocn/chuyaji/services/api/internal/model"
	"github.com/gin-gonic/gin"
)

type babyOut struct {
	ID            uint64     `json:"id"`
	UserID        uint64     `json:"user_id"`
	FamilyID      uint64     `json:"family_id"`
	Nickname      string     `json:"nickname"`
	Gender        string     `json:"gender,omitempty"`
	LMPDate       *time.Time `json:"lmp_date,omitempty"`
	EDDDate       *time.Time `json:"edd_date,omitempty"`
	BirthDate     *time.Time `json:"birth_date,omitempty"`
	BirthWeightG  *int       `json:"birth_weight_g,omitempty"`
	BirthHeightCM *float64   `json:"birth_height_cm,omitempty"`
	BirthHospital string     `json:"birth_hospital,omitempty"`
	FeedingType   string     `json:"feeding_type,omitempty"`
	Note          string     `json:"note,omitempty"`
}

func babyToOut(b *model.Baby) babyOut {
	return babyOut{
		ID:            b.ID,
		UserID:        b.UserID,
		FamilyID:      b.FamilyID,
		Nickname:      b.Nickname,
		Gender:        b.Gender,
		LMPDate:       b.LMPDate,
		EDDDate:       b.EDDDate,
		BirthDate:     b.BirthDate,
		BirthWeightG:  b.BirthWeightG,
		BirthHeightCM: b.BirthHeightCM,
		BirthHospital: b.BirthHospital,
		FeedingType:   b.FeedingType,
		Note:          b.Note,
	}
}

func (h *Handler) ListBabies(c *gin.Context) {
	uid, ok := middleware.UserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	familyIDs, err := h.userFamilyIDs(uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "query"})
		return
	}
	var babies []model.Baby
	if len(familyIDs) == 0 {
		c.JSON(http.StatusOK, gin.H{"items": []babyOut{}})
		return
	}
	if err := h.DB.Where("family_id IN ?", familyIDs).Order("id ASC").Find(&babies).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "query"})
		return
	}
	out := make([]babyOut, 0, len(babies))
	for i := range babies {
		out = append(out, babyToOut(&babies[i]))
	}
	c.JSON(http.StatusOK, gin.H{"items": out})
}

type createBabyReq struct {
	Nickname      string     `json:"nickname" binding:"required,max=64"`
	Gender        string     `json:"gender" binding:"omitempty,max=8"`
	LMPDate       *time.Time `json:"lmp_date"`
	EDDDate       *time.Time `json:"edd_date"`
	BirthDate     *time.Time `json:"birth_date"`
	BirthWeightG  *int       `json:"birth_weight_g"`
	BirthHeightCM *float64   `json:"birth_height_cm"`
	BirthHospital string     `json:"birth_hospital" binding:"omitempty,max=128"`
	FeedingType   string     `json:"feeding_type" binding:"omitempty,max=32"`
	Note          string     `json:"note" binding:"omitempty,max=1024"`
}

func (h *Handler) CreateBaby(c *gin.Context) {
	uid, ok := middleware.UserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	var req createBabyReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}
	familyID, err := h.getOrCreateUserFamilyID(uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "family"})
		return
	}
	role, okRole, err := h.resolveFamilyRole(uid, familyID)
	if err != nil || !okRole || !roleCanWrite(role) {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}
	b := model.Baby{
		UserID:        uid,
		FamilyID:      familyID,
		Nickname:      req.Nickname,
		Gender:        req.Gender,
		LMPDate:       req.LMPDate,
		EDDDate:       req.EDDDate,
		BirthDate:     req.BirthDate,
		BirthWeightG:  req.BirthWeightG,
		BirthHeightCM: req.BirthHeightCM,
		BirthHospital: req.BirthHospital,
		FeedingType:   req.FeedingType,
		Note:          req.Note,
	}
	if err := h.DB.Create(&b).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "create"})
		return
	}
	c.JSON(http.StatusCreated, babyToOut(&b))
}

func (h *Handler) GetBaby(c *gin.Context) {
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
	ok2, err := h.canAccessBaby(uid, id)
	if err != nil || !ok2 {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}
	var b model.Baby
	if err := h.DB.First(&b, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, babyToOut(&b))
}

type patchBabyReq struct {
	Nickname      *string    `json:"nickname"`
	Gender        *string    `json:"gender"`
	LMPDate       *time.Time `json:"lmp_date"`
	EDDDate       *time.Time `json:"edd_date"`
	BirthDate     *time.Time `json:"birth_date"`
	BirthWeightG  *int       `json:"birth_weight_g"`
	BirthHeightCM *float64   `json:"birth_height_cm"`
	BirthHospital *string    `json:"birth_hospital"`
	FeedingType   *string    `json:"feeding_type"`
	Note          *string    `json:"note"`
}

func (h *Handler) PatchBaby(c *gin.Context) {
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
	ok2, err := h.requireBabyWrite(uid, id)
	if err != nil || !ok2 {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}
	var req patchBabyReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}
	var b model.Baby
	if err := h.DB.First(&b, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	if req.Nickname != nil {
		b.Nickname = *req.Nickname
	}
	if req.Gender != nil {
		b.Gender = *req.Gender
	}
	if req.LMPDate != nil {
		b.LMPDate = req.LMPDate
	}
	if req.EDDDate != nil {
		b.EDDDate = req.EDDDate
	}
	if req.BirthDate != nil {
		b.BirthDate = req.BirthDate
	}
	if req.BirthWeightG != nil {
		b.BirthWeightG = req.BirthWeightG
	}
	if req.BirthHeightCM != nil {
		b.BirthHeightCM = req.BirthHeightCM
	}
	if req.BirthHospital != nil {
		b.BirthHospital = *req.BirthHospital
	}
	if req.FeedingType != nil {
		b.FeedingType = *req.FeedingType
	}
	if req.Note != nil {
		b.Note = *req.Note
	}
	if err := h.DB.Save(&b).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "save"})
		return
	}
	c.JSON(http.StatusOK, babyToOut(&b))
}

func (h *Handler) DeleteBaby(c *gin.Context) {
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
	ok2, err := h.requireBabyOwner(uid, id)
	if err != nil || !ok2 {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}
	if err := h.DB.Delete(&model.Baby{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "delete"})
		return
	}
	c.Status(http.StatusNoContent)
}
