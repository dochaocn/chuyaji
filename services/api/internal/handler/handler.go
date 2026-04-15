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

func (h *Handler) familyIDsForUser(uid uint64) ([]uint64, error) {
	var ids []uint64
	err := h.DB.Model(&model.FamilyMember{}).Where("user_id = ?", uid).Pluck("family_id", &ids).Error
	return ids, err
}

func (h *Handler) isFamilyMember(uid, familyID uint64) (bool, error) {
	var n int64
	err := h.DB.Model(&model.FamilyMember{}).Where("user_id = ? AND family_id = ?", uid, familyID).Count(&n).Error
	return n > 0, err
}

func (h *Handler) babyFamilyID(babyID uint64) (uint64, error) {
	var b model.Baby
	if err := h.DB.First(&b, babyID).Error; err != nil {
		return 0, err
	}
	return b.FamilyID, nil
}

func (h *Handler) canAccessBaby(uid, babyID uint64) (bool, error) {
	fid, err := h.babyFamilyID(babyID)
	if err != nil {
		return false, err
	}
	return h.isFamilyMember(uid, fid)
}

func (h *Handler) canAccessRecord(uid, recordID uint64) (bool, error) {
	var r model.Record
	if err := h.DB.First(&r, recordID).Error; err != nil {
		return false, err
	}
	return h.canAccessBaby(uid, r.BabyID)
}

// EncodeCursor packs occurred_at + id for keyset pagination (records sorted by occurred_at DESC, id DESC).
func EncodeCursor(t time.Time, id uint64) string {
	raw := fmt.Sprintf("%s|%d", t.UTC().Format(time.RFC3339Nano), id)
	return base64.URLEncoding.EncodeToString([]byte(raw))
}

func DecodeCursor(s string) (time.Time, uint64, error) {
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
