package handler

import (
	"errors"

	"github.com/dochaocn/chuyaji/services/api/internal/model"
	"gorm.io/gorm"
)

var errFamilyHasOtherMembers = errors.New("family has other members")

func purgeFamilyData(tx *gorm.DB, familyID uint64) error {
	var babyIDs []uint64
	if err := tx.Model(&model.Baby{}).Where("family_id = ?", familyID).Pluck("id", &babyIDs).Error; err != nil {
		return err
	}
	for _, bid := range babyIDs {
		var recordIDs []uint64
		if err := tx.Model(&model.Record{}).Where("baby_id = ?", bid).Pluck("id", &recordIDs).Error; err != nil {
			return err
		}
		if len(recordIDs) > 0 {
			if err := tx.Where("owner_type = ? AND owner_id IN ?", "baby_record", recordIDs).Delete(&model.Attachment{}).Error; err != nil {
				return err
			}
			if err := tx.Where("source_type = ? AND source_id IN ?", "baby_record", recordIDs).Delete(&model.Reminder{}).Error; err != nil {
				return err
			}
		}
		if err := tx.Where("baby_id = ?", bid).Delete(&model.Record{}).Error; err != nil {
			return err
		}
		if err := tx.Where("owner_type = ? AND owner_id = ?", "baby", bid).Delete(&model.Reminder{}).Error; err != nil {
			return err
		}
	}
	if err := tx.Where("family_id = ?", familyID).Delete(&model.Baby{}).Error; err != nil {
		return err
	}

	var motherIDs []uint64
	if err := tx.Model(&model.Mother{}).Where("family_id = ?", familyID).Pluck("id", &motherIDs).Error; err != nil {
		return err
	}
	for _, mid := range motherIDs {
		var recordIDs []uint64
		if err := tx.Model(&model.MotherRecord{}).Where("mother_id = ?", mid).Pluck("id", &recordIDs).Error; err != nil {
			return err
		}
		if len(recordIDs) > 0 {
			if err := tx.Where("owner_type = ? AND owner_id IN ?", "mother_record", recordIDs).Delete(&model.Attachment{}).Error; err != nil {
				return err
			}
			if err := tx.Where("source_type = ? AND source_id IN ?", "mother_record", recordIDs).Delete(&model.Reminder{}).Error; err != nil {
				return err
			}
		}
		if err := tx.Where("mother_id = ?", mid).Delete(&model.MotherRecord{}).Error; err != nil {
			return err
		}
		if err := tx.Where("owner_type = ? AND owner_id = ?", "mother", mid).Delete(&model.Reminder{}).Error; err != nil {
			return err
		}
	}
	if err := tx.Where("family_id = ?", familyID).Delete(&model.Mother{}).Error; err != nil {
		return err
	}

	if err := tx.Where("family_id = ?", familyID).Delete(&model.FamilyMember{}).Error; err != nil {
		return err
	}
	if err := tx.Where("family_id = ?", familyID).Delete(&model.FamilyInvite{}).Error; err != nil {
		return err
	}
	return tx.Delete(&model.Family{}, familyID).Error
}

// detachUserFromFamilyForJoin removes the user from their current family so they can join another.
// Sole-member families are fully purged; non-owners in multi-member families only leave.
func detachUserFromFamilyForJoin(tx *gorm.DB, uid uint64, existing *model.FamilyMember) error {
	var others int64
	if err := tx.Model(&model.FamilyMember{}).Where("family_id = ? AND user_id != ?", existing.FamilyID, uid).Count(&others).Error; err != nil {
		return err
	}
	if others > 0 {
		if existing.Role == model.FamilyRoleOwner {
			return errFamilyHasOtherMembers
		}
		return tx.Delete(existing).Error
	}
	return purgeFamilyData(tx, existing.FamilyID)
}

type familyConflictOut struct {
	HasExistingFamily bool   `json:"has_existing_family"`
	FamilyName        string `json:"family_name,omitempty"`
	BabyCount         int64  `json:"baby_count"`
	MotherCount       int64  `json:"mother_count"`
	HasOtherMembers   bool   `json:"has_other_members"`
	MyRole            string `json:"my_role,omitempty"`
}

func (h *Handler) buildFamilyConflict(uid, targetFamilyID uint64) (*familyConflictOut, error) {
	fam, member, err := h.currentFamilyMembership(uid)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if fam.ID == targetFamilyID {
		return nil, nil
	}
	var babyCount, motherCount int64
	if err := h.DB.Model(&model.Baby{}).Where("family_id = ?", fam.ID).Count(&babyCount).Error; err != nil {
		return nil, err
	}
	if err := h.DB.Model(&model.Mother{}).Where("family_id = ?", fam.ID).Count(&motherCount).Error; err != nil {
		return nil, err
	}
	var others int64
	if err := h.DB.Model(&model.FamilyMember{}).Where("family_id = ? AND user_id != ?", fam.ID, uid).Count(&others).Error; err != nil {
		return nil, err
	}
	return &familyConflictOut{
		HasExistingFamily: true,
		FamilyName:        fam.Name,
		BabyCount:         babyCount,
		MotherCount:       motherCount,
		HasOtherMembers:   others > 0 && member.Role == model.FamilyRoleOwner,
		MyRole:            member.Role,
	}, nil
}
