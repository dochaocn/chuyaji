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

type motherOut struct {
	ID                   uint64     `json:"id"`
	UserID               uint64     `json:"user_id"`
	Name                 string     `json:"name"`
	Birthday             *time.Time `json:"birthday,omitempty"`
	HeightCM             *float64   `json:"height_cm,omitempty"`
	PrePregnancyWeightKG *float64   `json:"pre_pregnancy_weight_kg,omitempty"`
	BloodType            string     `json:"blood_type,omitempty"`
	AllergyHistory       string     `json:"allergy_history,omitempty"`
	MedicalHistory       string     `json:"medical_history,omitempty"`
	Status               string     `json:"status,omitempty"`
	DeliveryDate         *time.Time `json:"delivery_date,omitempty"`
	Note                 string     `json:"note,omitempty"`
}

func motherToOut(m *model.Mother) motherOut {
	return motherOut{
		ID:                   m.ID,
		UserID:               m.UserID,
		Name:                 m.Name,
		Birthday:             m.Birthday,
		HeightCM:             m.HeightCM,
		PrePregnancyWeightKG: m.PrePregnancyWeightKG,
		BloodType:            m.BloodType,
		AllergyHistory:       m.AllergyHistory,
		MedicalHistory:       m.MedicalHistory,
		Status:               m.Status,
		DeliveryDate:         m.DeliveryDate,
		Note:                 m.Note,
	}
}

type createMotherReq struct {
	Name                 string     `json:"name" binding:"omitempty,max=64"`
	Birthday             *time.Time `json:"birthday"`
	HeightCM             *float64   `json:"height_cm"`
	PrePregnancyWeightKG *float64   `json:"pre_pregnancy_weight_kg"`
	BloodType            string     `json:"blood_type" binding:"omitempty,max=8"`
	AllergyHistory       string     `json:"allergy_history" binding:"omitempty,max=1024"`
	MedicalHistory       string     `json:"medical_history" binding:"omitempty,max=1024"`
	Status               string     `json:"status" binding:"omitempty,oneof=pregnant postpartum parenting"`
	DeliveryDate         *time.Time `json:"delivery_date"`
	Note                 string     `json:"note" binding:"omitempty,max=1024"`
}

type patchMotherReq = createMotherReq

func (h *Handler) ListMothers(c *gin.Context) {
	uid, ok := middleware.UserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	var mothers []model.Mother
	if err := h.DB.Where("user_id = ?", uid).Order("id ASC").Find(&mothers).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "query"})
		return
	}
	out := make([]motherOut, 0, len(mothers))
	for i := range mothers {
		out = append(out, motherToOut(&mothers[i]))
	}
	c.JSON(http.StatusOK, gin.H{"items": out})
}

func (h *Handler) CreateMother(c *gin.Context) {
	uid, ok := middleware.UserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	var req createMotherReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}
	var count int64
	if err := h.DB.Model(&model.Mother{}).Where("user_id = ?", uid).Count(&count).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "query"})
		return
	}
	if count > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "mother profile already exists"})
		return
	}
	m := model.Mother{
		UserID:               uid,
		Name:                 req.Name,
		Birthday:             req.Birthday,
		HeightCM:             req.HeightCM,
		PrePregnancyWeightKG: req.PrePregnancyWeightKG,
		BloodType:            req.BloodType,
		AllergyHistory:       req.AllergyHistory,
		MedicalHistory:       req.MedicalHistory,
		Status:               req.Status,
		DeliveryDate:         req.DeliveryDate,
		Note:                 req.Note,
	}
	if err := h.DB.Create(&m).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "create"})
		return
	}
	c.JSON(http.StatusCreated, motherToOut(&m))
}

func (h *Handler) GetMother(c *gin.Context) {
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
	ok2, err := h.canAccessMother(uid, id)
	if err != nil || !ok2 {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}
	var m model.Mother
	if err := h.DB.First(&m, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, motherToOut(&m))
}

func (h *Handler) PatchMother(c *gin.Context) {
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
	ok2, err := h.canAccessMother(uid, id)
	if err != nil || !ok2 {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}
	var req patchMotherReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}
	var m model.Mother
	if err := h.DB.First(&m, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	m.Name = req.Name
	m.Birthday = req.Birthday
	m.HeightCM = req.HeightCM
	m.PrePregnancyWeightKG = req.PrePregnancyWeightKG
	m.BloodType = req.BloodType
	m.AllergyHistory = req.AllergyHistory
	m.MedicalHistory = req.MedicalHistory
	m.Status = req.Status
	m.DeliveryDate = req.DeliveryDate
	m.Note = req.Note
	if err := h.DB.Save(&m).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "save"})
		return
	}
	c.JSON(http.StatusOK, motherToOut(&m))
}

type motherRecordOut struct {
	ID         uint64          `json:"id"`
	MotherID   uint64          `json:"mother_id"`
	RecordType string          `json:"record_type"`
	OccurredAt time.Time       `json:"occurred_at"`
	Summary    string          `json:"summary"`
	Payload    json.RawMessage `json:"payload"`
	CreatedAt  time.Time       `json:"created_at"`
	UpdatedAt  time.Time       `json:"updated_at"`
}

func motherRecordToOut(r *model.MotherRecord) motherRecordOut {
	var raw json.RawMessage
	if len(r.Payload) > 0 {
		raw = json.RawMessage(r.Payload)
	} else {
		raw = json.RawMessage(`{}`)
	}
	return motherRecordOut{
		ID:         r.ID,
		MotherID:   r.MotherID,
		RecordType: r.RecordType,
		OccurredAt: r.OccurredAt,
		Summary:    r.Summary,
		Payload:    raw,
		CreatedAt:  r.CreatedAt,
		UpdatedAt:  r.UpdatedAt,
	}
}

type createMotherRecordReq struct {
	RecordType string          `json:"record_type" binding:"required,max=32"`
	OccurredAt time.Time       `json:"occurred_at" binding:"required"`
	Summary    string          `json:"summary" binding:"max=512"`
	Payload    json.RawMessage `json:"payload"`
}

type patchMotherRecordReq struct {
	RecordType *string         `json:"record_type"`
	OccurredAt *time.Time      `json:"occurred_at"`
	Summary    *string         `json:"summary"`
	Payload    json.RawMessage `json:"payload"`
}

func (h *Handler) ListMotherRecords(c *gin.Context) {
	uid, ok := middleware.UserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	motherID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad mother id"})
		return
	}
	ok2, err := h.canAccessMother(uid, motherID)
	if err != nil || !ok2 {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	cursor := c.Query("cursor")

	q := h.DB.Where("mother_id = ?", motherID).Order("occurred_at DESC, id DESC")
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

	var rows []model.MotherRecord
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

	out := make([]motherRecordOut, 0, len(rows))
	for i := range rows {
		out = append(out, motherRecordToOut(&rows[i]))
	}
	c.JSON(http.StatusOK, gin.H{"items": out, "next_cursor": next})
}

func (h *Handler) LatestMotherRecord(c *gin.Context) {
	uid, ok := middleware.UserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	motherID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad mother id"})
		return
	}
	ok2, err := h.canAccessMother(uid, motherID)
	if err != nil || !ok2 {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}
	recordType := strings.TrimSpace(c.Query("record_type"))
	if recordType == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad record_type"})
		return
	}
	var r model.MotherRecord
	if err := h.DB.Where("mother_id = ? AND record_type = ?", motherID, recordType).
		Order("occurred_at DESC, id DESC").
		First(&r).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusOK, gin.H{"item": nil})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "query"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"item": motherRecordToOut(&r)})
}

func (h *Handler) CreateMotherRecord(c *gin.Context) {
	uid, ok := middleware.UserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	motherID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad mother id"})
		return
	}
	ok2, err := h.canAccessMother(uid, motherID)
	if err != nil || !ok2 {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}
	var req createMotherRecordReq
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
	r := model.MotherRecord{
		MotherID:   motherID,
		RecordType: req.RecordType,
		OccurredAt: req.OccurredAt,
		Summary:    req.Summary,
		Payload:    payload,
	}
	if err := h.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&r).Error; err != nil {
			return err
		}
		return h.syncMotherRecordReminder(tx, uid, &r)
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "create"})
		return
	}
	c.JSON(http.StatusCreated, motherRecordToOut(&r))
}

func (h *Handler) GetMotherRecord(c *gin.Context) {
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
	ok2, err := h.canAccessMotherRecord(uid, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	if err != nil || !ok2 {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}
	var r model.MotherRecord
	if err := h.DB.First(&r, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, motherRecordToOut(&r))
}

func (h *Handler) PatchMotherRecord(c *gin.Context) {
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
	ok2, err := h.canAccessMotherRecord(uid, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	if err != nil || !ok2 {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}
	var req patchMotherRecordReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}
	var r model.MotherRecord
	if err := h.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.First(&r, id).Error; err != nil {
			return err
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
		return h.syncMotherRecordReminder(tx, uid, &r)
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
	c.JSON(http.StatusOK, motherRecordToOut(&r))
}

func (h *Handler) DeleteMotherRecord(c *gin.Context) {
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
	ok2, err := h.canAccessMotherRecord(uid, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	if err != nil || !ok2 {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}
	if err := h.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("owner_type = ? AND owner_id = ?", "mother_record", id).Delete(&model.Attachment{}).Error; err != nil {
			return err
		}
		if err := tx.Where("source_type = ? AND source_id = ?", "mother_record", id).Delete(&model.Reminder{}).Error; err != nil {
			return err
		}
		return tx.Delete(&model.MotherRecord{}, id).Error
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "delete"})
		return
	}
	c.Status(http.StatusNoContent)
}
