package handler

import (
	"errors"
	"net/http"
	"os"
	"strconv"

	"github.com/dochaocn/chuyaji/services/api/internal/middleware"
	"github.com/dochaocn/chuyaji/services/api/internal/model"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type attachmentOut struct {
	ID        uint64 `json:"id"`
	RecordID  uint64 `json:"record_id"`
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

func (h *Handler) ListAttachments(c *gin.Context) {
	uid, ok := middleware.UserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	rid, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad record id"})
		return
	}
	ok2, err := h.canAccessRecord(uid, rid)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	if err != nil || !ok2 {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}
	var atts []model.Attachment
	if err := h.DB.Where("record_id = ?", rid).Order("sort_order ASC, id ASC").Find(&atts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "query"})
		return
	}
	out := make([]attachmentOut, 0, len(atts))
	for _, a := range atts {
		out = append(out, attachmentOut{
			ID: a.ID, RecordID: a.RecordID, URL: a.URL, ThumbURL: a.ThumbURL, SortOrder: a.SortOrder, Size: a.Size,
		})
	}
	c.JSON(http.StatusOK, gin.H{"items": out})
}

func (h *Handler) CreateAttachment(c *gin.Context) {
	uid, ok := middleware.UserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	rid, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad record id"})
		return
	}
	ok2, err := h.canAccessRecord(uid, rid)
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
		RecordID:  rid,
		URL:       req.URL,
		ThumbURL:  req.ThumbURL,
		SortOrder: req.SortOrder,
		Size:      req.Size,
	}
	if err := h.DB.Create(&a).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "create"})
		return
	}
	c.JSON(http.StatusCreated, attachmentOut{
		ID: a.ID, RecordID: a.RecordID, URL: a.URL, ThumbURL: a.ThumbURL, SortOrder: a.SortOrder, Size: a.Size,
	})
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
	ok2, err := h.canAccessRecord(uid, a.RecordID)
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
