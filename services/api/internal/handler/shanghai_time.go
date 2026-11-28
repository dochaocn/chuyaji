package handler

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

// 今日统计一律按 Asia/Shanghai 日历日（00:00 含，次日 00:00 不含）。
//
// SQLite 把 time.Time 存成「墙钟 + 偏移」文本，例如 UTC 的
// 2026-11-27 17:30:00+00:00，或 +08 的 2026-11-28 01:30:00+08:00。
// 两者是同一瞬间，但 `occurred_at >= ?` 按文本比较：上海零点会被写成
// 2026-11-28 00:00:00+08:00，和 +00:00 的行对不上，今日窗口会滑到大约
// 08:00–次日 08:00。datetime() 会先按偏移折成同一时间轴，再比较。

func shanghaiLocation() *time.Location {
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		return time.FixedZone("CST", 8*3600)
	}
	return loc
}

// shanghaiDayBounds 返回包含 now 的上海日历日 [start, end)，两者都是 UTC 瞬间。
func shanghaiDayBounds(now time.Time) (start, end time.Time, date string) {
	loc := shanghaiLocation()
	local := now.In(loc)
	startLocal := time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, loc)
	endLocal := startLocal.AddDate(0, 0, 1)
	return startLocal.UTC(), endLocal.UTC(), startLocal.Format("2006-01-02")
}

// shanghaiStatsSince 是「近 days 个上海日历日」（含今天）的起始瞬间。
func shanghaiStatsSince(now time.Time, days int) time.Time {
	if days < 1 {
		days = 1
	}
	start, _, _ := shanghaiDayBounds(now)
	return start.AddDate(0, 0, -(days - 1))
}

// shanghaiCalendarDateExpr 把时间列折成上海日历日 YYYY-MM-DD。
// column 只能是本包写死的列名，不能来自请求参数。
func shanghaiCalendarDateExpr(column string) string {
	return fmt.Sprintf("date(%s, '+8 hours')", column)
}

func whereInstantFrom(q *gorm.DB, column string, t time.Time) *gorm.DB {
	return q.Where(fmt.Sprintf("datetime(%s) >= datetime(?)", column), t.UTC())
}

func whereInstantBefore(q *gorm.DB, column string, t time.Time) *gorm.DB {
	return q.Where(fmt.Sprintf("datetime(%s) < datetime(?)", column), t.UTC())
}

// parseQueryDateTime 解析筛选时间。
// 带偏移的 RFC3339 是绝对瞬间；纯日期 YYYY-MM-DD 是上海日历日。
// exclusiveEnd 为真时，纯日期表示该日的下一上海零点（右开区间）。
func parseQueryDateTime(raw string, exclusiveEnd bool) (time.Time, error) {
	if t, err := time.Parse(time.RFC3339Nano, raw); err == nil {
		return t.UTC(), nil
	}
	if t, err := time.Parse(time.RFC3339, raw); err == nil {
		return t.UTC(), nil
	}
	t, err := time.ParseInLocation("2006-01-02", raw, shanghaiLocation())
	if err != nil {
		return time.Time{}, err
	}
	if exclusiveEnd {
		t = t.AddDate(0, 0, 1)
	}
	return t.UTC(), nil
}
