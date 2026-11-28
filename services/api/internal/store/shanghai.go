package store

import (
	"fmt"
	"strings"
	"time"

	"github.com/dochaocn/chuyaji/services/api/internal/tz"
	"gorm.io/gorm"
)

// 这些列保存的是时刻。历史数据多由 UTC 的 time.Time 写成 +00:00 文本，
// 和上海墙钟的查询对不齐。这里只改写文本，瞬间不变；已经是 +08:00 的行会跳过。
var shanghaiTimeColumns = []struct {
	table  string
	column string
}{
	{"users", "created_at"},
	{"users", "updated_at"},
	{"families", "created_at"},
	{"families", "updated_at"},
	{"family_members", "created_at"},
	{"family_members", "updated_at"},
	{"family_invites", "expires_at"},
	{"family_invites", "accepted_at"},
	{"family_invites", "revoked_at"},
	{"family_invites", "created_at"},
	{"family_invites", "updated_at"},
	{"babies", "lmp_date"},
	{"babies", "edd_date"},
	{"babies", "birth_date"},
	{"babies", "created_at"},
	{"babies", "updated_at"},
	{"records", "occurred_at"},
	{"records", "created_at"},
	{"records", "updated_at"},
	{"mothers", "birthday"},
	{"mothers", "delivery_date"},
	{"mothers", "created_at"},
	{"mothers", "updated_at"},
	{"mother_records", "occurred_at"},
	{"mother_records", "created_at"},
	{"mother_records", "updated_at"},
	{"attachments", "created_at"},
	{"reminders", "due_at"},
	{"reminders", "done_at"},
	{"reminders", "snoozed_until"},
	{"reminders", "created_at"},
	{"reminders", "updated_at"},
}

// NormalizeTimesToShanghai 把已有时间列改成上海墙钟文本。可重复执行。
func NormalizeTimesToShanghai(db *gorm.DB) error {
	for _, col := range shanghaiTimeColumns {
		if err := rewriteTimeColumn(db, col.table, col.column); err != nil {
			return fmt.Errorf("normalize %s.%s: %w", col.table, col.column, err)
		}
	}
	return nil
}

func rewriteTimeColumn(db *gorm.DB, table, column string) error {
	var tables int
	if err := db.Raw(`SELECT COUNT(*) FROM sqlite_master WHERE type = 'table' AND name = ?`, table).Scan(&tables).Error; err != nil {
		return err
	}
	if tables == 0 {
		return nil
	}
	var cols int
	if err := db.Raw(fmt.Sprintf(`SELECT COUNT(*) FROM pragma_table_info('%s') WHERE name = ?`, table), column).Scan(&cols).Error; err != nil {
		return err
	}
	if cols == 0 {
		return nil
	}

	type row struct {
		ID  uint64
		Raw string
	}
	var rows []row
	query := fmt.Sprintf(
		"SELECT id, %s AS raw FROM %s WHERE %s IS NOT NULL AND TRIM(%s) != '' AND %s NOT LIKE ?",
		column, table, column, column, column,
	)
	if err := db.Raw(query, "%+08:00").Scan(&rows).Error; err != nil {
		return err
	}
	for _, r := range rows {
		parsed, ok := parseStoredTime(r.Raw)
		if !ok {
			continue
		}
		stmt := fmt.Sprintf("UPDATE %s SET %s = ? WHERE id = ?", table, column)
		if err := db.Exec(stmt, tz.Wall(parsed), r.ID).Error; err != nil {
			return err
		}
	}
	return nil
}

func parseStoredTime(raw string) (time.Time, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return time.Time{}, false
	}
	layouts := []string{
		"2006-01-02 15:04:05.999999999-07:00",
		"2006-01-02T15:04:05.999999999-07:00",
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02 15:04:05.999999999",
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05.999999999",
		"2006-01-02T15:04:05",
		"2006-01-02",
	}
	for _, layout := range layouts {
		if strings.Contains(layout, "Z07") || strings.Contains(layout, "-07") {
			if t, err := time.Parse(layout, raw); err == nil {
				return t, true
			}
			continue
		}
		if t, err := time.ParseInLocation(layout, raw, time.UTC); err == nil {
			return t, true
		}
	}
	return time.Time{}, false
}
