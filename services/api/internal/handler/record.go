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

type recordAttachmentOut struct {
	ID       uint64 `json:"id"`
	URL      string `json:"url"`
	ThumbURL string `json:"thumb_url,omitempty"`
}

type recordOut struct {
	ID          uint64                `json:"id"`
	BabyID      uint64                `json:"baby_id"`
	Phase       string                `json:"phase"`
	RecordType  string                `json:"record_type"`
	OccurredAt  time.Time             `json:"occurred_at"`
	Summary     string                `json:"summary"`
	Payload     json.RawMessage       `json:"payload"`
	Attachments []recordAttachmentOut `json:"attachments,omitempty"`
	CreatedAt   time.Time             `json:"created_at"`
	UpdatedAt   time.Time             `json:"updated_at"`
}

// attachRecordAttachments 给一批记录批量挂上图片附件（一次 IN 查询，避免 N+1）。
// 只处理 owner_type = baby_record，且失败时保持原样返回，不影响主流程。
func (h *Handler) attachRecordAttachments(list []recordOut) []recordOut {
	if len(list) == 0 {
		return list
	}
	ids := make([]uint64, 0, len(list))
	for i := range list {
		ids = append(ids, list[i].ID)
	}
	var rows []model.Attachment
	if err := h.DB.Where("owner_type = ? AND owner_id IN ?", "baby_record", ids).
		Order("sort_order ASC, id ASC").Find(&rows).Error; err != nil {
		return list
	}
	if len(rows) == 0 {
		return list
	}
	grouped := make(map[uint64][]recordAttachmentOut, len(ids))
	for i := range rows {
		grouped[rows[i].OwnerID] = append(grouped[rows[i].OwnerID], recordAttachmentOut{
			ID:       rows[i].ID,
			URL:      rows[i].URL,
			ThumbURL: rows[i].ThumbURL,
		})
	}
	for i := range list {
		if atts, ok := grouped[list[i].ID]; ok {
			list[i].Attachments = atts
		}
	}
	return list
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
		q = whereInstantFrom(q, "occurred_at", t)
	}
	if to := c.Query("to"); to != "" {
		t, err := parseQueryDateTime(to, true)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "bad to"})
			return
		}
		q = whereInstantBefore(q, "occurred_at", t)
	}
	if cursor != "" {
		t, id, err := decodeCursor(cursor)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "bad cursor"})
			return
		}
		t = shanghaiWall(t)
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
	out = h.attachRecordAttachments(out)
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
	ok2, err := h.requireBabyWrite(uid, babyID)
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
		OccurredAt: shanghaiWall(req.OccurredAt),
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
	out := recordToOut(&r)
	out = h.attachRecordAttachments([]recordOut{out})[0]
	c.JSON(http.StatusOK, out)
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
	ok2, err := h.requireRecordWrite(uid, id)
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
			r.OccurredAt = shanghaiWall(*req.OccurredAt)
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
	out := recordToOut(&r)
	out = h.attachRecordAttachments([]recordOut{out})[0]
	c.JSON(http.StatusOK, out)
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
	ok2, err := h.requireRecordWrite(uid, id)
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
