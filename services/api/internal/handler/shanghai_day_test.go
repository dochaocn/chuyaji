package handler

import (
	"testing"
	"time"

	"github.com/dochaocn/chuyaji/services/api/internal/model"
	"github.com/dochaocn/chuyaji/services/api/internal/store"
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
	if start.Format(time.RFC3339) != "2026-11-28T00:00:00+08:00" {
		t.Fatalf("start = %s", start.Format(time.RFC3339))
	}
	if end.Format(time.RFC3339) != "2026-11-29T00:00:00+08:00" {
		t.Fatalf("end = %s", end.Format(time.RFC3339))
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
	if start.Format(time.RFC3339) != "2026-11-28T00:00:00+08:00" {
		t.Fatalf("from = %s", start.Format(time.RFC3339))
	}
	if end.Format(time.RFC3339) != "2026-11-29T00:00:00+08:00" {
		t.Fatalf("to = %s", end.Format(time.RFC3339))
	}

	abs, err := parseQueryDateTime("2026-11-27T17:30:00Z", false)
	if err != nil {
		t.Fatal(err)
	}
	if abs.Format(time.RFC3339) != "2026-11-28T01:30:00+08:00" {
		t.Fatalf("rfc3339 = %s", abs.Format(time.RFC3339))
	}
}

func TestBabyDailySummaryUsesShanghaiDay(t *testing.T) {
	db := openShanghaiTestDB(t)
	loc := shanghaiLocation()
	const babyID = 7
	rows := []model.Record{
		{BabyID: babyID, Phase: "postnatal", RecordType: "feeding", OccurredAt: time.Date(2026, 11, 28, 1, 30, 0, 0, loc), Payload: datatypes.JSON([]byte(`{"amount_ml":120}`))},
		{BabyID: babyID, Phase: "postnatal", RecordType: "feeding", OccurredAt: time.Date(2026, 11, 28, 23, 30, 0, 0, loc), Payload: datatypes.JSON([]byte(`{"amount_ml":80}`))},
		{BabyID: babyID, Phase: "postnatal", RecordType: "feeding", OccurredAt: time.Date(2026, 11, 29, 1, 30, 0, 0, loc), Payload: datatypes.JSON([]byte(`{"amount_ml":50}`))},
		{BabyID: babyID, Phase: "postnatal", RecordType: "sleep", OccurredAt: time.Date(2026, 11, 27, 23, 0, 0, 0, loc), Payload: datatypes.JSON([]byte(`{"duration_min":40}`))},
		{BabyID: babyID, Phase: "postnatal", RecordType: "sleep", OccurredAt: time.Date(2026, 11, 28, 8, 0, 0, 0, loc), Payload: datatypes.JSON([]byte(`{"duration_min":90}`))},
		{BabyID: babyID, Phase: "prenatal", RecordType: "feeding", OccurredAt: time.Date(2026, 11, 28, 12, 0, 0, 0, loc), Payload: datatypes.JSON([]byte(`{"amount_ml":999}`))},
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

	var buckets []dailyCount
	dayExpr := shanghaiCalendarDateExpr("occurred_at")
	since := shanghaiStatsSince(now, 2)
	if err := db.Model(&model.Record{}).
		Select(dayExpr+" as date, count(*) as baby_records").
		Where("occurred_at >= ? AND phase = ?", since, "postnatal").
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

func TestNormalizeUTCTextIntoShanghaiDay(t *testing.T) {
	db := openShanghaiTestDB(t)
	loc := shanghaiLocation()
	instant := time.Date(2026, 11, 28, 1, 30, 0, 0, loc)
	row := model.Record{
		BabyID: 1, Phase: "postnatal", RecordType: "feeding",
		OccurredAt: instant.UTC(),
		Payload:    datatypes.JSON([]byte(`{"amount_ml":30}`)),
	}
	if err := db.Create(&row).Error; err != nil {
		t.Fatal(err)
	}
	start, end, _ := shanghaiDayBounds(instant)
	var before int64
	if err := db.Model(&model.Record{}).Where("occurred_at >= ? AND occurred_at < ?", start, end).Count(&before).Error; err != nil {
		t.Fatal(err)
	}
	if before != 0 {
		t.Fatalf("utc text matched shanghai bounds: %d", before)
	}
	if err := store.NormalizeTimesToShanghai(db); err != nil {
		t.Fatal(err)
	}
	var raw string
	if err := db.Raw(`SELECT quote(occurred_at) FROM records WHERE id = ?`, row.ID).Scan(&raw).Error; err != nil {
		t.Fatal(err)
	}
	if raw != "'2026-11-28 01:30:00+08:00'" {
		t.Fatalf("stored = %s", raw)
	}
	var after int64
	if err := db.Model(&model.Record{}).Where("occurred_at >= ? AND occurred_at < ?", start, end).Count(&after).Error; err != nil {
		t.Fatal(err)
	}
	if after != 1 {
		t.Fatalf("after normalize count = %d", after)
	}
	if err := store.NormalizeTimesToShanghai(db); err != nil {
		t.Fatal(err)
	}
	var again string
	if err := db.Raw(`SELECT quote(occurred_at) FROM records WHERE id = ?`, row.ID).Scan(&again).Error; err != nil {
		t.Fatal(err)
	}
	if again != raw {
		t.Fatalf("second pass changed %s -> %s", raw, again)
	}
}

func openShanghaiTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
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
	return db
}
