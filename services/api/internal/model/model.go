package model

import (
	"time"

	"gorm.io/datatypes"
)

type User struct {
	ID        uint64 `gorm:"primaryKey"`
	OpenID    string `gorm:"uniqueIndex;size:64;not null"`
	Nickname  string `gorm:"size:64"`
	AvatarURL string `gorm:"size:512"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Baby struct {
	ID            uint64 `gorm:"primaryKey"`
	UserID        uint64 `gorm:"index;not null"`
	Nickname      string `gorm:"size:64;not null"`
	Gender        string `gorm:"size:8"`
	LMPDate       *time.Time
	EDDDate       *time.Time
	BirthDate     *time.Time
	BirthWeightG  *int
	BirthHeightCM *float64
	BirthHospital string `gorm:"size:128"`
	FeedingType   string `gorm:"size:32"`
	Note          string `gorm:"size:1024"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type Mother struct {
	ID                   uint64 `gorm:"primaryKey"`
	UserID               uint64 `gorm:"uniqueIndex;not null"`
	Name                 string `gorm:"size:64"`
	Birthday             *time.Time
	HeightCM             *float64
	PrePregnancyWeightKG *float64
	BloodType            string `gorm:"size:8"`
	AllergyHistory       string `gorm:"size:1024"`
	MedicalHistory       string `gorm:"size:1024"`
	Status               string `gorm:"size:16"`
	DeliveryDate         *time.Time
	Note                 string `gorm:"size:1024"`
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

type Record struct {
	ID         uint64 `gorm:"primaryKey"`
	BabyID     uint64 `gorm:"index;not null"`
	Phase      string `gorm:"size:16;not null"`
	RecordType string `gorm:"size:32;not null"`
	OccurredAt time.Time
	Summary    string `gorm:"size:512"`
	Payload    datatypes.JSON
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type MotherRecord struct {
	ID         uint64 `gorm:"primaryKey"`
	MotherID   uint64 `gorm:"index;not null"`
	RecordType string `gorm:"size:32;not null"`
	OccurredAt time.Time
	Summary    string `gorm:"size:512"`
	Payload    datatypes.JSON
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type Attachment struct {
	ID         uint64 `gorm:"primaryKey"`
	OwnerType  string `gorm:"index;size:32;not null"`
	OwnerID    uint64 `gorm:"index;not null"`
	URL        string `gorm:"size:1024;not null"`
	ThumbURL   string `gorm:"size:1024"`
	SortOrder  int
	Size       int64
	ShareToken *string `gorm:"uniqueIndex;size:64"`
	LocalPath  string  `gorm:"size:1024"`
	CreatedAt  time.Time
}

type Reminder struct {
	ID               uint64 `gorm:"primaryKey"`
	UserID           uint64 `gorm:"index;not null"`
	OwnerType        string `gorm:"index;size:32;not null"`
	OwnerID          uint64 `gorm:"index;not null"`
	SourceType       string `gorm:"uniqueIndex:idx_reminder_source;size:32;not null"`
	SourceID         uint64 `gorm:"uniqueIndex:idx_reminder_source;not null"`
	SourceRecordType string `gorm:"size:32;not null"`
	Title            string `gorm:"size:128;not null"`
	DueAt            time.Time
	Status           string `gorm:"index;size:16;not null;default:pending"`
	CreatedAt        time.Time
	UpdatedAt        time.Time
}
