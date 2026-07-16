package handler

import (
	"encoding/base64"
	"errors"
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

func roleCanWrite(role string) bool {
	return role == model.FamilyRoleOwner || role == model.FamilyRoleWrite
}

func (h *Handler) resolveFamilyRole(uid, familyID uint64) (string, bool, error) {
	if familyID == 0 {
		return "", false, nil
	}
	var m model.FamilyMember
	err := h.DB.Where("family_id = ? AND user_id = ?", familyID, uid).First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return m.Role, true, nil
}

func (h *Handler) userFamilyIDs(uid uint64) ([]uint64, error) {
	var ids []uint64
	err := h.DB.Model(&model.FamilyMember{}).Where("user_id = ?", uid).Pluck("family_id", &ids).Error
	return ids, err
}

func (h *Handler) getUserFamilyID(uid uint64) (uint64, error) {
	_, member, err := h.currentFamilyMembership(uid)
	if err != nil {
		return 0, err
	}
	return member.FamilyID, nil
}

func (h *Handler) createUserFamily(uid uint64) (uint64, error) {
	var familyID uint64
	err := h.DB.Transaction(func(tx *gorm.DB) error {
		fam := model.Family{Name: "我的家庭", CreatedBy: uid}
		if err := tx.Create(&fam).Error; err != nil {
			return err
		}
		m := model.FamilyMember{
			FamilyID: fam.ID,
			UserID:   uid,
			Role:     model.FamilyRoleOwner,
		}
		if err := tx.Create(&m).Error; err != nil {
			return err
		}
		familyID = fam.ID
		return nil
	})
	return familyID, err
}

func (h *Handler) getOrCreateUserFamilyID(uid uint64) (uint64, error) {
	familyID, err := h.getUserFamilyID(uid)
	if err == nil {
		return familyID, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, err
	}
	return h.createUserFamily(uid)
}

func (h *Handler) familyRoleForBaby(uid, babyID uint64) (string, bool, error) {
	var b model.Baby
	if err := h.DB.Select("id", "family_id").First(&b, babyID).Error; err != nil {
		return "", false, err
	}
	return h.resolveFamilyRole(uid, b.FamilyID)
}

func (h *Handler) familyRoleForMother(uid, motherID uint64) (string, bool, error) {
	var m model.Mother
	if err := h.DB.Select("id", "family_id").First(&m, motherID).Error; err != nil {
		return "", false, err
	}
	return h.resolveFamilyRole(uid, m.FamilyID)
}

func (h *Handler) canAccessBaby(uid, babyID uint64) (bool, error) {
	_, ok, err := h.familyRoleForBaby(uid, babyID)
	return ok, err
}

func (h *Handler) requireBabyWrite(uid, babyID uint64) (bool, error) {
	role, ok, err := h.familyRoleForBaby(uid, babyID)
	if err != nil || !ok {
		return false, err
	}
	return roleCanWrite(role), nil
}

func (h *Handler) requireBabyOwner(uid, babyID uint64) (bool, error) {
	role, ok, err := h.familyRoleForBaby(uid, babyID)
	if err != nil || !ok {
		return false, err
	}
	return role == model.FamilyRoleOwner, nil
}

func (h *Handler) canAccessRecord(uid, recordID uint64) (bool, error) {
	var r model.Record
	if err := h.DB.First(&r, recordID).Error; err != nil {
		return false, err
	}
	return h.canAccessBaby(uid, r.BabyID)
}

func (h *Handler) requireRecordWrite(uid, recordID uint64) (bool, error) {
	var r model.Record
	if err := h.DB.First(&r, recordID).Error; err != nil {
		return false, err
	}
	return h.requireBabyWrite(uid, r.BabyID)
}

func (h *Handler) canAccessMother(uid, motherID uint64) (bool, error) {
	_, ok, err := h.familyRoleForMother(uid, motherID)
	return ok, err
}

func (h *Handler) requireMotherWrite(uid, motherID uint64) (bool, error) {
	role, ok, err := h.familyRoleForMother(uid, motherID)
	if err != nil || !ok {
		return false, err
	}
	return roleCanWrite(role), nil
}

func (h *Handler) canAccessMotherRecord(uid, recordID uint64) (bool, error) {
	var r model.MotherRecord
	if err := h.DB.First(&r, recordID).Error; err != nil {
		return false, err
	}
	return h.canAccessMother(uid, r.MotherID)
}

func (h *Handler) requireMotherRecordWrite(uid, recordID uint64) (bool, error) {
	var r model.MotherRecord
	if err := h.DB.First(&r, recordID).Error; err != nil {
		return false, err
	}
	return h.requireMotherWrite(uid, r.MotherID)
}

func (h *Handler) accessibleBabyIDs(uid uint64) ([]uint64, error) {
	familyIDs, err := h.userFamilyIDs(uid)
	if err != nil {
		return nil, err
	}
	if len(familyIDs) == 0 {
		return nil, nil
	}
	var ids []uint64
	err = h.DB.Model(&model.Baby{}).Where("family_id IN ?", familyIDs).Pluck("id", &ids).Error
	return ids, err
}

func (h *Handler) accessibleMotherIDs(uid uint64) ([]uint64, error) {
	familyIDs, err := h.userFamilyIDs(uid)
	if err != nil {
		return nil, err
	}
	if len(familyIDs) == 0 {
		return nil, nil
	}
	var ids []uint64
	err = h.DB.Model(&model.Mother{}).Where("family_id IN ?", familyIDs).Pluck("id", &ids).Error
	return ids, err
}

func (h *Handler) canWriteReminder(uid uint64, r *model.Reminder) (bool, error) {
	switch r.OwnerType {
	case "baby":
		return h.requireBabyWrite(uid, r.OwnerID)
	case "mother":
		return h.requireMotherWrite(uid, r.OwnerID)
	default:
		return false, nil
	}
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
