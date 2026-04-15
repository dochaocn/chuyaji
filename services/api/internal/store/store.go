package store

import (
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
		&model.Baby{},
		&model.Record{},
		&model.Mother{},
		&model.MotherRecord{},
		&model.Attachment{},
	); err != nil {
		return nil, err
	}
	if err := migrateLegacyBabyOwnership(db); err != nil {
		return nil, err
	}
	if err := migrateLegacyAttachmentOwners(db); err != nil {
		return nil, err
	}
	return db, nil
}

type pragmaColumn struct {
	Name string `gorm:"column:name"`
}

func hasColumn(db *gorm.DB, table, column string) (bool, error) {
	var cols []pragmaColumn
	if err := db.Raw(fmt.Sprintf("PRAGMA table_info(%s)", table)).Scan(&cols).Error; err != nil {
		return false, err
	}
	for _, col := range cols {
		if col.Name == column {
			return true, nil
		}
	}
	return false, nil
}

func migrateLegacyBabyOwnership(db *gorm.DB) error {
	hasFamilyID, err := hasColumn(db, "babies", "family_id")
	if err != nil {
		return err
	}
	if !hasFamilyID {
		return nil
	}
	return db.Exec(`
		UPDATE babies
		SET user_id = COALESCE(
			NULLIF(user_id, 0),
			(
				SELECT fm.user_id
				FROM family_members fm
				WHERE fm.family_id = babies.family_id
				ORDER BY CASE WHEN fm.role = 'owner' THEN 0 ELSE 1 END, fm.id ASC
				LIMIT 1
			)
		)
		WHERE COALESCE(user_id, 0) = 0
	`).Error
}

func migrateLegacyAttachmentOwners(db *gorm.DB) error {
	hasRecordID, err := hasColumn(db, "attachments", "record_id")
	if err != nil {
		return err
	}
	if !hasRecordID {
		return nil
	}
	return db.Exec(`
		UPDATE attachments
		SET owner_type = COALESCE(NULLIF(owner_type, ''), 'baby_record'),
		    owner_id = COALESCE(NULLIF(owner_id, 0), record_id)
		WHERE COALESCE(owner_id, 0) = 0 OR COALESCE(owner_type, '') = ''
	`).Error
}
