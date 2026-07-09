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
	"gorm.io/gorm"
)

const (
	reminderStatusPending = "pending"
	reminderStatusDone    = "done"
	reminderStatusIgnored = "ignored"
)

type reminderOut struct {
	ID               uint64     `json:"id"`
	UserID           uint64     `json:"user_id"`
	OwnerType        string     `json:"owner_type"`
	OwnerID          uint64     `json:"owner_id"`
	OwnerName        string     `json:"owner_name,omitempty"`
	SourceType       string     `json:"source_type"`
	SourceID         uint64     `json:"source_id"`
	SourceRecordType string     `json:"source_record_type"`
	Category         string     `json:"category"`
	Title            string     `json:"title"`
	Note             string     `json:"note"`
	DueAt            time.Time  `json:"due_at"`
	DoneAt           *time.Time `json:"done_at,omitempty"`
	SnoozedUntil     *time.Time `json:"snoozed_until,omitempty"`
	Status           string     `json:"status"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

func reminderToOut(r *model.Reminder) reminderOut {
	return reminderOut{
		ID:               r.ID,
		UserID:           r.UserID,
		OwnerType:        r.OwnerType,
		OwnerID:          r.OwnerID,
		SourceType:       r.SourceType,
		SourceID:         r.SourceID,
		SourceRecordType: r.SourceRecordType,
		Category:         r.Category,
		Title:            r.Title,
		Note:             r.Note,
		DueAt:            r.DueAt,
		DoneAt:           r.DoneAt,
		SnoozedUntil:     r.SnoozedUntil,
		Status:           r.Status,
		CreatedAt:        r.CreatedAt,
		UpdatedAt:        r.UpdatedAt,
	}
}

func (h *Handler) ListReminders(c *gin.Context) {
	uid, ok := middleware.UserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	status := c.DefaultQuery("status", reminderStatusPending)
	if !validReminderStatus(status) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad status"})
		return
	}

	query := h.DB.Where("user_id = ? AND status = ?", uid, status)
	if ownerType := c.Query("owner_type"); ownerType != "" {
		if ownerType != "baby" && ownerType != "mother" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "bad owner_type"})
			return
		}
		query = query.Where("owner_type = ?", ownerType)
	}
	if ownerIDRaw := c.Query("owner_id"); ownerIDRaw != "" {
		ownerID, err := strconv.ParseUint(ownerIDRaw, 10, 64)
		if err != nil || ownerID == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "bad owner_id"})
			return
		}
		query = query.Where("owner_id = ?", ownerID)
	}
	if category := c.Query("category"); category != "" {
		query = query.Where("category = ?", category)
	}
	if from := c.Query("from"); from != "" {
		t, err := parseQueryDateTime(from, false)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "bad from"})
			return
		}
		query = query.Where("due_at >= ?", t)
	}
	if to := c.Query("to"); to != "" {
		t, err := parseQueryDateTime(to, true)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "bad to"})
			return
		}
		query = query.Where("due_at < ?", t)
	}

	var rows []model.Reminder
	if err := query.
		Order("due_at ASC, id ASC").
		Limit(limit).
		Find(&rows).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "query"})
		return
	}
	out := make([]reminderOut, 0, len(rows))
	for i := range rows {
		out = append(out, reminderToOut(&rows[i]))
	}
	c.JSON(http.StatusOK, gin.H{"items": out})
}

type patchReminderReq struct {
	Status       *string    `json:"status"`
	DueAt        *time.Time `json:"due_at"`
	SnoozedUntil *time.Time `json:"snoozed_until"`
	Note         *string    `json:"note"`
}

func (h *Handler) PatchReminder(c *gin.Context) {
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
	var req patchReminderReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}
	var r model.Reminder
	if err := h.DB.Where("id = ? AND user_id = ?", id, uid).First(&r).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "query"})
		return
	}
	if req.Status != nil {
		if !validReminderStatus(*req.Status) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "bad status"})
			return
		}
		r.Status = *req.Status
		if *req.Status == reminderStatusDone && r.DoneAt == nil {
			now := time.Now()
			r.DoneAt = &now
		}
		if *req.Status != reminderStatusDone {
			r.DoneAt = nil
		}
	}
	if req.DueAt != nil {
		r.DueAt = *req.DueAt
	}
	if req.SnoozedUntil != nil {
		r.SnoozedUntil = req.SnoozedUntil
		r.DueAt = *req.SnoozedUntil
	}
	if req.Note != nil {
		r.Note = *req.Note
	}
	if err := h.DB.Save(&r).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "save"})
		return
	}
	c.JSON(http.StatusOK, reminderToOut(&r))
}

func (h *Handler) DeleteReminder(c *gin.Context) {
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
	res := h.DB.Where("id = ? AND user_id = ?", id, uid).Delete(&model.Reminder{})
	if res.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "delete"})
		return
	}
	if res.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.Status(http.StatusNoContent)
}

func validReminderStatus(status string) bool {
	return status == reminderStatusPending || status == reminderStatusDone || status == reminderStatusIgnored
}

func (h *Handler) syncBabyRecordReminder(db *gorm.DB, uid uint64, r *model.Record) error {
	if r.Phase == "postnatal" && r.RecordType != "checkup" && r.RecordType != "vaccine" {
		return h.deleteSourceReminder(db, "baby_record", r.ID)
	}
	title, category, note := babyReminderMeta(r.Phase, r.RecordType, r.Payload)
	return h.upsertSourceReminder(db, uid, "baby", r.BabyID, "baby_record", r.ID, r.RecordType, category, title, note, r.Payload)
}

func (h *Handler) syncMotherRecordReminder(db *gorm.DB, uid uint64, r *model.MotherRecord) error {
	if r.RecordType != "checkup" && r.RecordType != "followup" {
		return h.deleteSourceReminder(db, "mother_record", r.ID)
	}
	title, category, note := motherReminderMeta(r.RecordType, r.Payload)
	return h.upsertSourceReminder(db, uid, "mother", r.MotherID, "mother_record", r.ID, r.RecordType, category, title, note, r.Payload)
}

func (h *Handler) upsertSourceReminder(db *gorm.DB, uid uint64, ownerType string, ownerID uint64, sourceType string, sourceID uint64, recordType string, category string, title string, note string, payload []byte) error {
	dueAt, ok := nextTimeFromPayload(payload)
	if !ok {
		return h.deleteSourceReminder(db, sourceType, sourceID)
	}
	var r model.Reminder
	err := db.Where("source_type = ? AND source_id = ?", sourceType, sourceID).First(&r).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		r = model.Reminder{
			UserID:           uid,
			OwnerType:        ownerType,
			OwnerID:          ownerID,
			SourceType:       sourceType,
			SourceID:         sourceID,
			SourceRecordType: recordType,
			Category:         category,
			Title:            title,
			Note:             note,
			DueAt:            dueAt,
			Status:           reminderStatusPending,
		}
		return db.Create(&r).Error
	}
	if err != nil {
		return err
	}
	r.UserID = uid
	r.OwnerType = ownerType
	r.OwnerID = ownerID
	r.SourceRecordType = recordType
	r.Category = category
	r.Title = title
	r.Note = note
	r.DueAt = dueAt
	r.Status = reminderStatusPending
	r.DoneAt = nil
	return db.Save(&r).Error
}

func (h *Handler) deleteSourceReminder(db *gorm.DB, sourceType string, sourceID uint64) error {
	return db.Where("source_type = ? AND source_id = ?", sourceType, sourceID).Delete(&model.Reminder{}).Error
}

func babyReminderMeta(phase string, recordType string, payload []byte) (string, string, string) {
	values := payloadMap(payload)
	switch recordType {
	case "vaccine":
		name := stringPayloadValue(values, "name")
		if name == "" {
			name = "疫苗"
		}
		return "下次疫苗：" + name, "vaccine", "来自宝宝疫苗记录"
	case "checkup":
		if phase == "prenatal" {
			return "下次产检", "checkup", "来自宝宝孕期记录"
		}
		return "下次体检", "checkup", "来自宝宝体检记录"
	default:
		if phase == "prenatal" {
			return "下次产检", "checkup", "来自宝宝孕期记录"
		}
		return "后续安排", "other", "来自宝宝记录"
	}
}

func motherReminderMeta(recordType string, payload []byte) (string, string, string) {
	switch recordType {
	case "checkup":
		return "下次产检", "checkup", "来自宝妈产检记录"
	case "followup":
		return "下次复诊", "followup", "来自宝妈复诊记录"
	default:
		return "后续安排", "other", "来自宝妈记录"
	}
}

func payloadMap(payload []byte) map[string]any {
	values := map[string]any{}
	if len(payload) == 0 {
		return values
	}
	_ = json.Unmarshal(payload, &values)
	return values
}

func stringPayloadValue(values map[string]any, key string) string {
	raw, ok := values[key]
	if !ok || raw == nil {
		return ""
	}
	s, ok := raw.(string)
	if !ok {
		return ""
	}
	return s
}

func nextTimeFromPayload(payload []byte) (time.Time, bool) {
	if len(payload) == 0 {
		return time.Time{}, false
	}
	var values map[string]any
	if err := json.Unmarshal(payload, &values); err != nil {
		return time.Time{}, false
	}
	raw, ok := values["next_time"].(string)
	if !ok || raw == "" {
		return time.Time{}, false
	}
	layouts := []string{time.RFC3339Nano, time.RFC3339, "2006-01-02"}
	for _, layout := range layouts {
		t, err := time.Parse(layout, raw)
		if err == nil {
			return t, true
		}
	}
	return time.Time{}, false
}
