package handler

import (
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/dochaocn/chuyaji/services/api/internal/config"
	"github.com/dochaocn/chuyaji/services/api/internal/model"
	"gorm.io/gorm"
)

type Handler struct {
	DB  *gorm.DB
	Cfg *config.Config
}

func New(db *gorm.DB, cfg *config.Config) *Handler {
	return &Handler{DB: db, Cfg: cfg}
}

func (h *Handler) babyUserID(babyID uint64) (uint64, error) {
	var b model.Baby
	if err := h.DB.First(&b, babyID).Error; err != nil {
		return 0, err
	}
	return b.UserID, nil
}

func (h *Handler) canAccessBaby(uid, babyID uint64) (bool, error) {
	ownerID, err := h.babyUserID(babyID)
	if err != nil {
		return false, err
	}
	return ownerID == uid, nil
}

func (h *Handler) canAccessRecord(uid, recordID uint64) (bool, error) {
	var r model.Record
	if err := h.DB.First(&r, recordID).Error; err != nil {
		return false, err
	}
	return h.canAccessBaby(uid, r.BabyID)
}

func (h *Handler) canAccessMother(uid, motherID uint64) (bool, error) {
	var m model.Mother
	if err := h.DB.First(&m, motherID).Error; err != nil {
		return false, err
	}
	return m.UserID == uid, nil
}

func (h *Handler) canAccessMotherRecord(uid, recordID uint64) (bool, error) {
	var r model.MotherRecord
	if err := h.DB.First(&r, recordID).Error; err != nil {
		return false, err
	}
	return h.canAccessMother(uid, r.MotherID)
}

func encodeCursor(t time.Time, id uint64) string {
	raw := fmt.Sprintf("%s|%d", t.UTC().Format(time.RFC3339Nano), id)
	return base64.URLEncoding.EncodeToString([]byte(raw))
}

func decodeCursor(s string) (time.Time, uint64, error) {
	b, err := base64.URLEncoding.DecodeString(s)
	if err != nil {
		return time.Time{}, 0, err
	}
	parts := strings.SplitN(string(b), "|", 2)
	if len(parts) != 2 {
		return time.Time{}, 0, fmt.Errorf("bad cursor")
	}
	t, err := time.Parse(time.RFC3339Nano, parts[0])
	if err != nil {
		return time.Time{}, 0, err
	}
	id, err := strconv.ParseUint(parts[1], 10, 64)
	if err != nil {
		return time.Time{}, 0, err
	}
	return t, id, nil
}
