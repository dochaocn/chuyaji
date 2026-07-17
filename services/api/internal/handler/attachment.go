package handler

import (
	"errors"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/dochaocn/chuyaji/services/api/internal/middleware"
	"github.com/dochaocn/chuyaji/services/api/internal/model"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type attachmentOut struct {
	ID        uint64 `json:"id"`
	OwnerType string `json:"owner_type"`
	OwnerID   uint64 `json:"owner_id"`
	URL       string `json:"url"`
	ThumbURL  string `json:"thumb_url,omitempty"`
	SortOrder int    `json:"sort_order"`
	Size      int64  `json:"size"`
}

type createAttachmentReq struct {
	URL       string `json:"url" binding:"required,max=1024"`
	ThumbURL  string `json:"thumb_url" binding:"max=1024"`
	SortOrder int    `json:"sort_order"`
	Size      int64  `json:"size"`
}

func attachmentToOut(a *model.Attachment) attachmentOut {
	return attachmentOut{
		ID:        a.ID,
		OwnerType: a.OwnerType,
		OwnerID:   a.OwnerID,
		URL:       a.URL,
		ThumbURL:  a.ThumbURL,
		SortOrder: a.SortOrder,
		Size:      a.Size,
	}
}

func (h *Handler) ensureAttachmentOwnerAccess(uid, ownerID uint64, ownerType string) (bool, error) {
	switch ownerType {
	case "baby_record":
		return h.canAccessRecord(uid, ownerID)
	case "mother_record":
		return h.canAccessMotherRecord(uid, ownerID)
	default:
		return false, gorm.ErrRecordNotFound
	}
}

func (h *Handler) ensureAttachmentOwnerWrite(uid, ownerID uint64, ownerType string) (bool, error) {
	switch ownerType {
	case "baby_record":
		return h.requireRecordWrite(uid, ownerID)
	case "mother_record":
		return h.requireMotherRecordWrite(uid, ownerID)
	default:
		return false, gorm.ErrRecordNotFound
	}
}

func (h *Handler) listAttachmentsByOwner(c *gin.Context, ownerType string) {
	uid, ok := middleware.UserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	ownerID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad owner id"})
		return
	}
	ok2, err := h.ensureAttachmentOwnerAccess(uid, ownerID, ownerType)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	if err != nil || !ok2 {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}
	var atts []model.Attachment
	if err := h.DB.Where("owner_type = ? AND owner_id = ?", ownerType, ownerID).Order("sort_order ASC, id ASC").Find(&atts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "query"})
		return
	}
	out := make([]attachmentOut, 0, len(atts))
	for i := range atts {
		out = append(out, attachmentToOut(&atts[i]))
	}
	c.JSON(http.StatusOK, gin.H{"items": out})
}

func (h *Handler) createAttachmentByOwner(c *gin.Context, ownerType string) {
	uid, ok := middleware.UserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	ownerID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad owner id"})
		return
	}
	ok2, err := h.ensureAttachmentOwnerWrite(uid, ownerID, ownerType)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	if err != nil || !ok2 {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}
	var req createAttachmentReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}
	a := model.Attachment{
		OwnerType: ownerType,
		OwnerID:   ownerID,
		URL:       req.URL,
		ThumbURL:  req.ThumbURL,
		SortOrder: req.SortOrder,
		Size:      req.Size,
	}
	if err := h.DB.Create(&a).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "create"})
		return
	}
	c.JSON(http.StatusCreated, attachmentToOut(&a))
}

func (h *Handler) ListAttachments(c *gin.Context) {
	h.listAttachmentsByOwner(c, "baby_record")
}

type babyAttachmentOut struct {
	ID         uint64    `json:"id"`
	URL        string    `json:"url"`
	ThumbURL   string    `json:"thumb_url,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
	RecordID   uint64    `json:"record_id"`
	RecordType string    `json:"record_type"`
	OccurredAt time.Time `json:"occurred_at"`
	Summary    string    `json:"summary"`
}

func (h *Handler) ListBabyAttachments(c *gin.Context) {
	uid, ok := middleware.UserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	babyID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad baby id"})
		return
	}
	ok2, err := h.canAccessBaby(uid, babyID)
	if err != nil || !ok2 {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "40"))
	if limit <= 0 || limit > 100 {
		limit = 40
	}
	cursor := c.Query("cursor")

	type babyAttRow struct {
		ID         uint64
		URL        string
		ThumbURL   string
		CreatedAt  time.Time
		RecordID   uint64
		RecordType string
		OccurredAt time.Time
		Summary    string
	}

	q := h.DB.Table("attachments AS a").
		Select("a.id, a.url, a.thumb_url, a.created_at, r.id AS record_id, r.record_type, r.occurred_at, r.summary").
		Joins("INNER JOIN records AS r ON r.id = a.owner_id AND a.owner_type = ?", "baby_record").
		Where("r.baby_id = ?", babyID).
		Order("a.created_at DESC, a.id DESC")

	if cursor != "" {
		t, id, err := decodeCursor(cursor)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "bad cursor"})
			return
		}
		q = q.Where("(a.created_at < ?) OR (a.created_at = ? AND a.id < ?)", t, t, id)
	}

	var rows []babyAttRow
	if err := q.Limit(limit + 1).Scan(&rows).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "query"})
		return
	}

	next := ""
	if len(rows) > limit {
		last := rows[limit-1]
		next = encodeCursor(last.CreatedAt, last.ID)
		rows = rows[:limit]
	}

	out := make([]babyAttachmentOut, 0, len(rows))
	for i := range rows {
		out = append(out, babyAttachmentOut{
			ID:         rows[i].ID,
			URL:        rows[i].URL,
			ThumbURL:   rows[i].ThumbURL,
			CreatedAt:  rows[i].CreatedAt,
			RecordID:   rows[i].RecordID,
			RecordType: rows[i].RecordType,
			OccurredAt: rows[i].OccurredAt,
			Summary:    rows[i].Summary,
		})
	}
	c.JSON(http.StatusOK, gin.H{"items": out, "next_cursor": next})
}

func (h *Handler) CreateAttachment(c *gin.Context) {
	h.createAttachmentByOwner(c, "baby_record")
}

func (h *Handler) ListMotherAttachments(c *gin.Context) {
	h.listAttachmentsByOwner(c, "mother_record")
}

func (h *Handler) CreateMotherAttachment(c *gin.Context) {
	h.createAttachmentByOwner(c, "mother_record")
}

func (h *Handler) DeleteAttachment(c *gin.Context) {
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
	var a model.Attachment
	if err := h.DB.First(&a, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	ok2, err := h.ensureAttachmentOwnerWrite(uid, a.OwnerID, a.OwnerType)
	if err != nil || !ok2 {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}
	if a.LocalPath != "" {
		_ = os.Remove(a.LocalPath)
	}
	if err := h.DB.Delete(&model.Attachment{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "delete"})
		return
	}
	c.Status(http.StatusNoContent)
}
