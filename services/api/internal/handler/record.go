package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/dochaocn/chuyaji/services/api/internal/middleware"
	"github.com/dochaocn/chuyaji/services/api/internal/model"
	"github.com/gin-gonic/gin"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type recordOut struct {
	ID               uint64          `json:"id"`
	BabyID           uint64          `json:"baby_id"`
	Phase            string          `json:"phase"`
	RecordType       string          `json:"record_type"`
	OccurredAt       time.Time       `json:"occurred_at"`
	GestationalWeeks *int            `json:"gestational_weeks,omitempty"`
	GestationalDays  *int            `json:"gestational_days,omitempty"`
	Summary          string          `json:"summary"`
	Payload          json.RawMessage `json:"payload"`
	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at"`
}

func recordToOut(r *model.Record) recordOut {
	var raw json.RawMessage
	if len(r.Payload) > 0 {
		raw = json.RawMessage(r.Payload)
	} else {
		raw = json.RawMessage(`{}`)
	}
	return recordOut{
		ID:               r.ID,
		BabyID:           r.BabyID,
		Phase:            r.Phase,
		RecordType:       r.RecordType,
		OccurredAt:       r.OccurredAt,
		GestationalWeeks: r.GestationalWeeks,
		GestationalDays:  r.GestationalDays,
		Summary:          r.Summary,
		Payload:          raw,
		CreatedAt:        r.CreatedAt,
		UpdatedAt:        r.UpdatedAt,
	}
}

func (h *Handler) ListRecords(c *gin.Context) {
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

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	cursor := c.Query("cursor")

	q := h.DB.Where("baby_id = ?", babyID).Order("occurred_at DESC, id DESC")
	if cursor != "" {
		t, id, err := DecodeCursor(cursor)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "bad cursor"})
			return
		}
		q = q.Where("(occurred_at < ?) OR (occurred_at = ? AND id < ?)", t, t, id)
	}

	var rows []model.Record
	if err := q.Limit(limit + 1).Find(&rows).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "query"})
		return
	}

	next := ""
	if len(rows) > limit {
		last := rows[limit-1]
		next = EncodeCursor(last.OccurredAt, last.ID)
		rows = rows[:limit]
	}

	out := make([]recordOut, 0, len(rows))
	for i := range rows {
		out = append(out, recordToOut(&rows[i]))
	}
	c.JSON(http.StatusOK, gin.H{"items": out, "next_cursor": next})
}

type createRecordReq struct {
	Phase            string          `json:"phase" binding:"required,oneof=prenatal postnatal"`
	RecordType       string          `json:"record_type" binding:"required,max=32"`
	OccurredAt       time.Time       `json:"occurred_at" binding:"required"`
	GestationalWeeks *int            `json:"gestational_weeks"`
	GestationalDays  *int            `json:"gestational_days"`
	Summary          string          `json:"summary" binding:"max=512"`
	Payload          json.RawMessage `json:"payload"`
}

func (h *Handler) CreateRecord(c *gin.Context) {
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
	var req createRecordReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}
	payload := datatypes.JSON([]byte(`{}`))
	if len(req.Payload) > 0 {
		if !json.Valid(req.Payload) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload json"})
			return
		}
		payload = datatypes.JSON(req.Payload)
	}
	r := model.Record{
		BabyID:           babyID,
		Phase:            req.Phase,
		RecordType:       req.RecordType,
		OccurredAt:       req.OccurredAt,
		GestationalWeeks: req.GestationalWeeks,
		GestationalDays:  req.GestationalDays,
		Summary:          req.Summary,
		Payload:          payload,
	}
	if err := h.DB.Create(&r).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "create"})
		return
	}
	c.JSON(http.StatusCreated, recordToOut(&r))
}

func (h *Handler) GetRecord(c *gin.Context) {
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
	ok2, err := h.canAccessRecord(uid, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	if err != nil || !ok2 {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}
	var r model.Record
	if err := h.DB.First(&r, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, recordToOut(&r))
}

type patchRecordReq struct {
	Phase            *string         `json:"phase"`
	RecordType       *string         `json:"record_type"`
	OccurredAt       *time.Time      `json:"occurred_at"`
	GestationalWeeks *int            `json:"gestational_weeks"`
	GestationalDays  *int            `json:"gestational_days"`
	Summary          *string         `json:"summary"`
	Payload          json.RawMessage `json:"payload"`
}

func (h *Handler) PatchRecord(c *gin.Context) {
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
	ok2, err := h.canAccessRecord(uid, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	if err != nil || !ok2 {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}
	var r model.Record
	if err := h.DB.First(&r, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	var req patchRecordReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}
	if req.Phase != nil {
		r.Phase = *req.Phase
	}
	if req.RecordType != nil {
		r.RecordType = *req.RecordType
	}
	if req.OccurredAt != nil {
		r.OccurredAt = *req.OccurredAt
	}
	if req.GestationalWeeks != nil {
		r.GestationalWeeks = req.GestationalWeeks
	}
	if req.GestationalDays != nil {
		r.GestationalDays = req.GestationalDays
	}
	if req.Summary != nil {
		r.Summary = *req.Summary
	}
	if len(req.Payload) > 0 {
		if !json.Valid(req.Payload) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload json"})
			return
		}
		r.Payload = datatypes.JSON(req.Payload)
	}
	if err := h.DB.Save(&r).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "save"})
		return
	}
	c.JSON(http.StatusOK, recordToOut(&r))
}

func (h *Handler) DeleteRecord(c *gin.Context) {
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
	ok2, err := h.canAccessRecord(uid, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	if err != nil || !ok2 {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}
	if err := h.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("owner_type = ? AND owner_id = ?", "baby_record", id).Delete(&model.Attachment{}).Error; err != nil {
			return err
		}
		return tx.Delete(&model.Record{}, id).Error
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "delete"})
		return
	}
	c.Status(http.StatusNoContent)
}
