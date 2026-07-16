package store

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/dochaocn/chuyaji/services/api/internal/model"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func Open(dbPath string) (*gorm.DB, error) {
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("mkdir data dir: %w", err)
	}
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return nil, err
	}
	if err := db.AutoMigrate(
		&model.User{},
		&model.Family{},
		&model.FamilyMember{},
		&model.FamilyInvite{},
		&model.Baby{},
		&model.Record{},
		&model.Mother{},
		&model.MotherRecord{},
		&model.Attachment{},
		&model.Reminder{},
	); err != nil {
		return nil, err
	}
	if err := migrateEnsureFamilies(db); err != nil {
		return nil, err
	}
	return db, nil
}

func migrateEnsureFamilies(db *gorm.DB) error {
	var users []model.User
	if err := db.Find(&users).Error; err != nil {
		return err
	}
	for _, u := range users {
		var member model.FamilyMember
		err := db.Where("user_id = ?", u.ID).First(&member).Error
		if err == nil {
			if err := db.Model(&model.Baby{}).
				Where("user_id = ? AND COALESCE(family_id, 0) = 0", u.ID).
				Update("family_id", member.FamilyID).Error; err != nil {
				return err
			}
			if err := db.Model(&model.Mother{}).
				Where("user_id = ? AND COALESCE(family_id, 0) = 0", u.ID).
				Update("family_id", member.FamilyID).Error; err != nil {
				return err
			}
			continue
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		fam := model.Family{
			Name:      "我的家庭",
			CreatedBy: u.ID,
		}
		if err := db.Create(&fam).Error; err != nil {
			return err
		}
		member = model.FamilyMember{
			FamilyID: fam.ID,
			UserID:   u.ID,
			Role:     model.FamilyRoleOwner,
		}
		if err := db.Create(&member).Error; err != nil {
			return err
		}
		if err := db.Model(&model.Baby{}).
			Where("user_id = ?", u.ID).
			Update("family_id", fam.ID).Error; err != nil {
			return err
		}
		if err := db.Model(&model.Mother{}).
			Where("user_id = ?", u.ID).
			Update("family_id", fam.ID).Error; err != nil {
			return err
		}
	}
	return nil
}
