package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/dochaocn/chuyaji/services/api/internal/middleware"
	"github.com/dochaocn/chuyaji/services/api/internal/model"
	"github.com/gin-gonic/gin"
)

func (h *Handler) BabyDashboard(c *gin.Context) {
	uid, ok := middleware.UserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	babyID, err := strconv.ParseUint(c.Query("baby_id"), 10, 64)
	if err != nil && c.Query("baby_id") != "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad baby_id"})
		return
	}

	familyIDs, err := h.userFamilyIDs(uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "query"})
		return
	}
	if len(familyIDs) == 0 {
		c.JSON(http.StatusOK, gin.H{"profile": nil, "latest_records": []recordOut{}})
		return
	}

	var baby model.Baby
	query := h.DB.Where("family_id IN ?", familyIDs).Order("id ASC")
	if babyID > 0 {
		query = query.Where("id = ?", babyID)
	}
	if err := query.First(&baby).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"profile": nil, "latest_records": []recordOut{}})
		return
	}

	var records []model.Record
	if err := h.DB.Where("baby_id = ?", baby.ID).Order("occurred_at DESC, id DESC").Limit(5).Find(&records).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "query"})
		return
	}

	out := make([]recordOut, 0, len(records))
	for i := range records {
		out = append(out, recordToOut(&records[i]))
	}

	stage := "prenatal"
	if baby.BirthDate != nil {
		stage = "postnatal"
	}

	growthSummary := gin.H{}
	var latestGrowth model.Record
	if err := h.DB.Where("baby_id = ? AND record_type = ?", baby.ID, "growth").Order("occurred_at DESC, id DESC").First(&latestGrowth).Error; err == nil {
		var payload map[string]any
		if json.Unmarshal(latestGrowth.Payload, &payload) == nil {
			growthSummary = gin.H{
				"latest_weight_g":  payload["weight_g"],
				"latest_weight_at": latestGrowth.OccurredAt,
			}
		}
	}

	dailySummary := h.buildBabyDailySummary(baby.ID)

	c.JSON(http.StatusOK, gin.H{
		"profile":        babyToOut(&baby),
		"phase_summary":  gin.H{"stage": stage, "record_count": len(out)},
		"latest_records": out,
		"growth_summary": growthSummary,
		"daily_summary":  dailySummary,
	})
}

func shanghaiLocation() *time.Location {
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		return time.FixedZone("CST", 8*3600)
	}
	return loc
}

func payloadFloat(payload map[string]any, key string) float64 {
	v, ok := payload[key]
	if !ok || v == nil {
		return 0
	}
	switch n := v.(type) {
	case float64:
		return n
	case float32:
		return float64(n)
	case int:
		return float64(n)
	case int64:
		return float64(n)
	case json.Number:
		f, err := n.Float64()
		if err != nil {
			return 0
		}
		return f
	case string:
		f, err := strconv.ParseFloat(n, 64)
		if err != nil {
			return 0
		}
		return f
	default:
		return 0
	}
}

func (h *Handler) buildBabyDailySummary(babyID uint64) gin.H {
	loc := shanghaiLocation()
	now := time.Now().In(loc)
	dayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
	dayEnd := dayStart.AddDate(0, 0, 1)
	dateStr := dayStart.Format("2006-01-02")

	var feedCount, sleepCount, diaperCount int
	var feedML, feedDur, sleepDur, diaperTimes float64

	var rows []model.Record
	if err := h.DB.Where(
		"baby_id = ? AND phase = ? AND record_type IN ? AND occurred_at >= ? AND occurred_at < ?",
		babyID, "postnatal", []string{"feeding", "sleep", "diaper"}, dayStart, dayEnd,
	).Find(&rows).Error; err == nil {
		for i := range rows {
			var payload map[string]any
			_ = json.Unmarshal(rows[i].Payload, &payload)
			if payload == nil {
				payload = map[string]any{}
			}
			switch rows[i].RecordType {
			case "feeding":
				feedCount++
				feedML += payloadFloat(payload, "amount_ml")
				feedDur += payloadFloat(payload, "duration_min")
			case "sleep":
				sleepCount++
				sleepDur += payloadFloat(payload, "duration_min")
			case "diaper":
				diaperCount++
				times := payloadFloat(payload, "times")
				if times <= 0 {
					times = 1
				}
				diaperTimes += times
			}
		}
	}

	return gin.H{
		"date": dateStr,
		"feeding": gin.H{
			"count":               feedCount,
			"total_ml":            int(feedML + 0.5),
			"total_duration_min":  int(feedDur + 0.5),
		},
		"sleep": gin.H{
			"count":              sleepCount,
			"total_duration_min": int(sleepDur + 0.5),
		},
		"diaper": gin.H{
			"count":       diaperCount,
			"total_times": int(diaperTimes + 0.5),
		},
	}
}

func (h *Handler) MotherDashboard(c *gin.Context) {
	uid, ok := middleware.UserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	motherID, err := strconv.ParseUint(c.Query("mother_id"), 10, 64)
	if err != nil && c.Query("mother_id") != "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad mother_id"})
		return
	}

	familyIDs, err := h.userFamilyIDs(uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "query"})
		return
	}
	if len(familyIDs) == 0 {
		c.JSON(http.StatusOK, gin.H{"profile": nil, "latest_records": []motherRecordOut{}})
		return
	}

	var mother model.Mother
	query := h.DB.Where("family_id IN ?", familyIDs).Order("id ASC")
	if motherID > 0 {
		query = query.Where("id = ?", motherID)
	}
	if err := query.First(&mother).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"profile": nil, "latest_records": []motherRecordOut{}})
		return
	}

	var records []model.MotherRecord
	if err := h.DB.Where("mother_id = ?", mother.ID).Order("occurred_at DESC, id DESC").Limit(5).Find(&records).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "query"})
		return
	}
	out := make([]motherRecordOut, 0, len(records))
	for i := range records {
		out = append(out, motherRecordToOut(&records[i]))
	}

	c.JSON(http.StatusOK, gin.H{
		"profile":        motherToOut(&mother),
		"health_summary": gin.H{"status": mother.Status, "record_count": len(out)},
		"latest_records": out,
	})
}
