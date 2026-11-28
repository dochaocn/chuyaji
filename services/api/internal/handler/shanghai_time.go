package handler

import (
	"fmt"
	"time"

	"github.com/dochaocn/chuyaji/services/api/internal/tz"
	"gorm.io/gorm"
)

// 库里的时间和查询参数都用 Asia/Shanghai 墙钟（文本形如 2026-11-28 01:30:00+08:00）。
// SQLite 按文本比较时间。只要写入和查询都是 +08:00，今日就是上海日历日
// 00:00（含）到次日 00:00（不含）。混进 UTC 的 +00:00 文本会对不齐，
// 所以启动时把已有时间列改写成同一瞬间的上海墙钟。

func shanghaiLocation() *time.Location {
	return tz.Location()
}

func shanghaiWall(t time.Time) time.Time {
	return tz.Wall(t)
}

// shanghaiDayBounds 返回包含 now 的上海日历日 [start, end)。
func shanghaiDayBounds(now time.Time) (start, end time.Time, date string) {
	loc := shanghaiLocation()
	local := now.In(loc)
	start = time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, loc)
	end = start.AddDate(0, 0, 1)
	return start, end, start.Format("2006-01-02")
}

// shanghaiStatsSince 是「近 days 个上海日历日」（含今天）的起始时刻。
func shanghaiStatsSince(now time.Time, days int) time.Time {
	if days < 1 {
		days = 1
	}
	start, _, _ := shanghaiDayBounds(now)
	return start.AddDate(0, 0, -(days - 1))
}

// shanghaiCalendarDateExpr 取上海墙钟文本的日期。列里必须已经是 +08:00 墙钟。
// column 只能是本包写死的列名，不能来自请求参数。
func shanghaiCalendarDateExpr(column string) string {
	return fmt.Sprintf("substr(%s, 1, 10)", column)
}

func whereInstantFrom(q *gorm.DB, column string, t time.Time) *gorm.DB {
	return q.Where(fmt.Sprintf("%s >= ?", column), shanghaiWall(t))
}

func whereInstantBefore(q *gorm.DB, column string, t time.Time) *gorm.DB {
	return q.Where(fmt.Sprintf("%s < ?", column), shanghaiWall(t))
}

// parseQueryDateTime 解析筛选时间。
// 带偏移的 RFC3339 是绝对瞬间，结果换成上海墙钟；纯日期 YYYY-MM-DD 是上海日历日。
// exclusiveEnd 为真时，纯日期表示该日的下一上海零点（右开区间）。
func parseQueryDateTime(raw string, exclusiveEnd bool) (time.Time, error) {
	if t, err := time.Parse(time.RFC3339Nano, raw); err == nil {
		return shanghaiWall(t), nil
	}
	if t, err := time.Parse(time.RFC3339, raw); err == nil {
		return shanghaiWall(t), nil
	}
	t, err := time.ParseInLocation("2006-01-02", raw, shanghaiLocation())
	if err != nil {
		return time.Time{}, err
	}
	if exclusiveEnd {
		t = t.AddDate(0, 0, 1)
	}
	return t, nil
}
