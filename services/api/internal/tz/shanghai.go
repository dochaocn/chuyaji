package tz

import "time"

// Location 是业务时间统一使用的时区。中国没有夏令时，固定 UTC+8。
func Location() *time.Location {
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		return time.FixedZone("CST", 8*3600)
	}
	return loc
}

// Wall 把绝对瞬间换成上海墙钟。写入 SQLite 和作为查询参数时都用它，
// 驱动会写成带 +08:00 的同一套文本，字符串比较才和日历日一致。
func Wall(t time.Time) time.Time {
	if t.IsZero() {
		return t
	}
	return t.In(Location())
}

func Now() time.Time {
	return Wall(time.Now())
}
