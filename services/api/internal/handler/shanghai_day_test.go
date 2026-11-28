package handler

import (
	"testing"
	"time"

	"github.com/dochaocn/chuyaji/services/api/internal/model"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

func TestShanghaiDayBounds(t *testing.T) {
	loc := shanghaiLocation()
	now := time.Date(2026, 11, 28, 1, 30, 0, 0, loc)
	start, end, date := shanghaiDayBounds(now)
	if date != "2026-11-28" {
		t.Fatalf("date = %s", date)
	}
	wantStart := time.Date(2026, 11, 27, 16, 0, 0, 0, time.UTC)
	wantEnd := time.Date(2026, 11, 28, 16, 0, 0, 0, time.UTC)
	if !start.Equal(wantStart) || !end.Equal(wantEnd) {
		t.Fatalf("bounds %s .. %s", start.Format(time.RFC3339), end.Format(time.RFC3339))
	}
}

func TestParseQueryDateTimeShanghaiDate(t *testing.T) {
	start, err := parseQueryDateTime("2026-11-28", false)
	if err != nil {
		t.Fatal(err)
	}
	end, err := parseQueryDateTime("2026-11-28", true)
	if err != nil {
		t.Fatal(err)
	}
	if !start.Equal(time.Date(2026, 11, 27, 16, 0, 0, 0, time.UTC)) {
		t.Fatalf("from = %s", start.Format(time.RFC3339))
	}
	if !end.Equal(time.Date(2026, 11, 28, 16, 0, 0, 0, time.UTC)) {
		t.Fatalf("to = %s", end.Format(time.RFC3339))
	}

	abs, err := parseQueryDateTime("2026-11-28T01:30:00+08:00", false)
	if err != nil {
		t.Fatal(err)
	}
	if !abs.Equal(time.Date(2026, 11, 27, 17, 30, 0, 0, time.UTC)) {
		t.Fatalf("rfc3339 = %s", abs.Format(time.RFC3339))
	}
}

func TestBabyDailySummaryUsesShanghaiDay(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:shanghai-day-test?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	sqlDB.SetMaxOpenConns(1)
	if err := db.AutoMigrate(&model.Record{}); err != nil {
		t.Fatal(err)
	}
	loc := shanghaiLocation()
	const babyID = 7
	rows := []model.Record{
		// 上海 11-28 01:30，按 UTC 文本存储（客户端 toISOString 的常见形态）
		{BabyID: babyID, Phase: "postnatal", RecordType: "feeding", OccurredAt: time.Date(2026, 11, 28, 1, 30, 0, 0, loc).UTC(), Payload: datatypes.JSON([]byte(`{"amount_ml":120}`))},
		// 同一天 23:30，以 +08 墙钟写入，文本偏移不同但瞬间正确
		{BabyID: babyID, Phase: "postnatal", RecordType: "feeding", OccurredAt: time.Date(2026, 11, 28, 23, 30, 0, 0, loc), Payload: datatypes.JSON([]byte(`{"amount_ml":80}`))},
		// 次日凌晨，不应计入 11-28
		{BabyID: babyID, Phase: "postnatal", RecordType: "feeding", OccurredAt: time.Date(2026, 11, 29, 1, 30, 0, 0, loc).UTC(), Payload: datatypes.JSON([]byte(`{"amount_ml":50}`))},
		// 前一天晚上
		{BabyID: babyID, Phase: "postnatal", RecordType: "sleep", OccurredAt: time.Date(2026, 11, 27, 23, 0, 0, 0, loc).UTC(), Payload: datatypes.JSON([]byte(`{"duration_min":40}`))},
		// 当天睡眠
		{BabyID: babyID, Phase: "postnatal", RecordType: "sleep", OccurredAt: time.Date(2026, 11, 28, 8, 0, 0, 0, loc).UTC(), Payload: datatypes.JSON([]byte(`{"duration_min":90}`))},
		// 怀孕期同刻，不进今日汇总
		{BabyID: babyID, Phase: "prenatal", RecordType: "feeding", OccurredAt: time.Date(2026, 11, 28, 12, 0, 0, 0, loc).UTC(), Payload: datatypes.JSON([]byte(`{"amount_ml":999}`))},
	}
	for i := range rows {
		if err := db.Create(&rows[i]).Error; err != nil {
			t.Fatal(err)
		}
	}

	h := &Handler{DB: db}
	now := time.Date(2026, 11, 28, 15, 0, 0, 0, loc)
	summary := h.buildBabyDailySummaryAt(babyID, now)
	if summary["date"] != "2026-11-28" {
		t.Fatalf("date = %#v", summary["date"])
	}
	feeding := summary["feeding"].(gin.H)
	if feeding["count"] != 2 || feeding["total_ml"] != 200 {
		t.Fatalf("feeding = %#v", feeding)
	}
	sleep := summary["sleep"].(gin.H)
	if sleep["count"] != 1 || sleep["total_duration_min"] != 90 {
		t.Fatalf("sleep = %#v", sleep)
	}

	// 管理端按日分桶也必须落在上海日历日，而不是 UTC 日期。
	var buckets []dailyCount
	dayExpr := shanghaiCalendarDateExpr("occurred_at")
	since := shanghaiStatsSince(now, 2)
	if err := db.Model(&model.Record{}).
		Select(dayExpr+" as date, count(*) as baby_records").
		Where("datetime(occurred_at) >= datetime(?) AND phase = ?", since, "postnatal").
		Group(dayExpr).
		Order("date ASC").
		Scan(&buckets).Error; err != nil {
		t.Fatal(err)
	}
	got := map[string]int64{}
	for _, b := range buckets {
		got[b.Date] = b.BabyRecords
	}
	if got["2026-11-27"] != 1 || got["2026-11-28"] != 3 || got["2026-11-29"] != 1 {
		t.Fatalf("buckets = %#v", got)
	}
}
