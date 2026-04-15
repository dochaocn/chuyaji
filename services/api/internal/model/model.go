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

type Family struct {
	ID         uint64 `gorm:"primaryKey"`
	Name       string `gorm:"size:64;not null"`
	InviteCode string `gorm:"uniqueIndex;size:16;not null"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type FamilyMember struct {
	ID        uint64 `gorm:"primaryKey"`
	FamilyID  uint64 `gorm:"uniqueIndex:idx_family_user;not null"`
	UserID    uint64 `gorm:"uniqueIndex:idx_family_user;not null"`
	Role      string `gorm:"size:16;not null"` // owner | member
	CreatedAt time.Time
}

type Baby struct {
	ID        uint64 `gorm:"primaryKey"`
	FamilyID  uint64 `gorm:"index;not null"`
	Nickname  string `gorm:"size:64;not null"`
	LMPDate   *time.Time
	EDDDate   *time.Time
	BirthDate *time.Time
	Gender    string `gorm:"size:8"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Record struct {
	ID               uint64 `gorm:"primaryKey"`
	BabyID           uint64 `gorm:"index;not null"`
	Phase            string `gorm:"size:16;not null"` // prenatal | postnatal
	RecordType       string `gorm:"size:32;not null"`
	OccurredAt       time.Time
	GestationalWeeks *int
	GestationalDays  *int
	Summary          string `gorm:"size:512"`
	Payload          datatypes.JSON
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type Attachment struct {
	ID        uint64 `gorm:"primaryKey"`
	RecordID  uint64 `gorm:"index;not null"`
	URL       string `gorm:"size:1024;not null"`
	ThumbURL  string `gorm:"size:1024"`
	SortOrder int
	Size      int64
	// ShareToken 非空表示本附件由服务端托管，可通过 /api/v1/p/:token 匿名读取（随机不可猜测）。
	ShareToken *string `gorm:"uniqueIndex;size:64"`
	LocalPath  string  `gorm:"size:1024"`
	CreatedAt  time.Time
}
