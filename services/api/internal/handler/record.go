package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/dochaocn/chuyaji/services/api/internal/middleware"
	"github.com/dochaocn/chuyaji/services/api/internal/model"
	"github.com/gin-gonic/gin"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

var errInvalidPayloadJSON = errors.New("invalid payload json")

type recordOut struct {
	ID         uint64          `json:"id"`
	BabyID     uint64          `json:"baby_id"`
	Phase      string          `json:"phase"`
	RecordType string          `json:"record_type"`
	OccurredAt time.Time       `json:"occurred_at"`
	Summary    string          `json:"summary"`
	Payload    json.RawMessage `json:"payload"`
	CreatedAt  time.Time       `json:"created_at"`
	UpdatedAt  time.Time       `json:"updated_at"`
}

func recordToOut(r *model.Record) recordOut {
	var raw json.RawMessage
	if len(r.Payload) > 0 {
		raw = json.RawMessage(r.Payload)
	} else {
		raw = json.RawMessage(`{}`)
	}
	return recordOut{
		ID:         r.ID,
		BabyID:     r.BabyID,
		Phase:      r.Phase,
		RecordType: r.RecordType,
		OccurredAt: r.OccurredAt,
		Summary:    r.Summary,
		Payload:    raw,
		CreatedAt:  r.CreatedAt,
		UpdatedAt:  r.UpdatedAt,
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
	if phase := c.Query("phase"); phase != "" {
		if phase != "prenatal" && phase != "postnatal" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "bad phase"})
			return
		}
		q = q.Where("phase = ?", phase)
	}
	if recordType := c.Query("record_type"); recordType != "" {
		q = q.Where("record_type = ?", recordType)
	}
	if keyword := strings.TrimSpace(c.Query("q")); keyword != "" {
		like := "%" + keyword + "%"
		q = q.Where("(summary LIKE ? OR payload LIKE ?)", like, like)
	}
	if from := c.Query("from"); from != "" {
		t, err := parseQueryDateTime(from, false)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "bad from"})
			return
		}
		q = q.Where("occurred_at >= ?", t)
	}
	if to := c.Query("to"); to != "" {
		t, err := parseQueryDateTime(to, true)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "bad to"})
			return
		}
		q = q.Where("occurred_at < ?", t)
	}
	if cursor != "" {
		t, id, err := decodeCursor(cursor)
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
		next = encodeCursor(last.OccurredAt, last.ID)
		rows = rows[:limit]
	}

	out := make([]recordOut, 0, len(rows))
	for i := range rows {
		out = append(out, recordToOut(&rows[i]))
	}
	c.JSON(http.StatusOK, gin.H{"items": out, "next_cursor": next})
}

func (h *Handler) LatestRecord(c *gin.Context) {
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
	recordType := strings.TrimSpace(c.Query("record_type"))
	if recordType == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad record_type"})
		return
	}
	q := h.DB.Where("baby_id = ? AND record_type = ?", babyID, recordType).Order("occurred_at DESC, id DESC")
	if phase := c.Query("phase"); phase != "" {
		if phase != "prenatal" && phase != "postnatal" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "bad phase"})
			return
		}
		q = q.Where("phase = ?", phase)
	}
	var r model.Record
	if err := q.First(&r).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusOK, gin.H{"item": nil})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "query"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"item": recordToOut(&r)})
}

type createRecordReq struct {
	Phase      string          `json:"phase" binding:"required,oneof=prenatal postnatal"`
	RecordType string          `json:"record_type" binding:"required,max=32"`
	OccurredAt time.Time       `json:"occurred_at" binding:"required"`
	Summary    string          `json:"summary" binding:"max=512"`
	Payload    json.RawMessage `json:"payload"`
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
		BabyID:     babyID,
		Phase:      req.Phase,
		RecordType: req.RecordType,
		OccurredAt: req.OccurredAt,
		Summary:    req.Summary,
		Payload:    payload,
	}
	if err := h.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&r).Error; err != nil {
			return err
		}
		return h.syncBabyRecordReminder(tx, uid, &r)
	}); err != nil {
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
	Phase      *string         `json:"phase"`
	RecordType *string         `json:"record_type"`
	OccurredAt *time.Time      `json:"occurred_at"`
	Summary    *string         `json:"summary"`
	Payload    json.RawMessage `json:"payload"`
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
	var req patchRecordReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}
	var r model.Record
	if err := h.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.First(&r, id).Error; err != nil {
			return err
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
		if req.Summary != nil {
			r.Summary = *req.Summary
		}
		if len(req.Payload) > 0 {
			if !json.Valid(req.Payload) {
				return errInvalidPayloadJSON
			}
			r.Payload = datatypes.JSON(req.Payload)
		}
		if err := tx.Save(&r).Error; err != nil {
			return err
		}
		return h.syncBabyRecordReminder(tx, uid, &r)
	}); err != nil {
		if errors.Is(err, errInvalidPayloadJSON) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload json"})
			return
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
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
		if err := tx.Where("source_type = ? AND source_id = ?", "baby_record", id).Delete(&model.Reminder{}).Error; err != nil {
			return err
		}
		return tx.Delete(&model.Record{}, id).Error
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "delete"})
		return
	}
	c.Status(http.StatusNoContent)
}

func parseQueryDateTime(raw string, exclusiveEnd bool) (time.Time, error) {
	layouts := []string{time.RFC3339Nano, time.RFC3339, "2006-01-02"}
	var lastErr error
	for _, layout := range layouts {
		t, err := time.Parse(layout, raw)
		if err == nil {
			if exclusiveEnd && layout == "2006-01-02" {
				return t.AddDate(0, 0, 1), nil
			}
			return t, nil
		}
		lastErr = err
	}
	return time.Time{}, lastErr
}
