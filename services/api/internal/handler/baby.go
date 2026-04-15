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
	ID        uint64     `json:"id"`
	FamilyID  uint64     `json:"family_id"`
	Nickname  string     `json:"nickname"`
	LMPDate   *time.Time `json:"lmp_date,omitempty"`
	EDDDate   *time.Time `json:"edd_date,omitempty"`
	BirthDate *time.Time `json:"birth_date,omitempty"`
	Gender    string     `json:"gender,omitempty"`
}

func babyToOut(b *model.Baby) babyOut {
	return babyOut{
		ID:        b.ID,
		FamilyID:  b.FamilyID,
		Nickname:  b.Nickname,
		LMPDate:   b.LMPDate,
		EDDDate:   b.EDDDate,
		BirthDate: b.BirthDate,
		Gender:    b.Gender,
	}
}

func (h *Handler) ListBabies(c *gin.Context) {
	uid, ok := middleware.UserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	fidStr := c.Query("family_id")
	if fidStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "family_id required"})
		return
	}
	fid, err := strconv.ParseUint(fidStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad family_id"})
		return
	}
	ok2, err := h.isFamilyMember(uid, fid)
	if err != nil || !ok2 {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}
	var babies []model.Baby
	if err := h.DB.Where("family_id = ?", fid).Order("id ASC").Find(&babies).Error; err != nil {
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
	FamilyID  uint64     `json:"family_id" binding:"required"`
	Nickname  string     `json:"nickname" binding:"required,max=64"`
	LMPDate   *time.Time `json:"lmp_date"`
	EDDDate   *time.Time `json:"edd_date"`
	BirthDate *time.Time `json:"birth_date"`
	Gender    string     `json:"gender" binding:"omitempty,max=8"`
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
	ok2, err := h.isFamilyMember(uid, req.FamilyID)
	if err != nil || !ok2 {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}
	b := model.Baby{
		FamilyID:  req.FamilyID,
		Nickname:  req.Nickname,
		LMPDate:   req.LMPDate,
		EDDDate:   req.EDDDate,
		BirthDate: req.BirthDate,
		Gender:    req.Gender,
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
	Nickname  *string    `json:"nickname"`
	LMPDate   *time.Time `json:"lmp_date"`
	EDDDate   *time.Time `json:"edd_date"`
	BirthDate *time.Time `json:"birth_date"`
	Gender    *string    `json:"gender"`
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
	ok2, err := h.canAccessBaby(uid, id)
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
	if req.LMPDate != nil {
		b.LMPDate = req.LMPDate
	}
	if req.EDDDate != nil {
		b.EDDDate = req.EDDDate
	}
	if req.BirthDate != nil {
		b.BirthDate = req.BirthDate
	}
	if req.Gender != nil {
		b.Gender = *req.Gender
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
	ok2, err := h.canAccessBaby(uid, id)
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
