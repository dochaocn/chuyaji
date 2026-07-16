package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/dochaocn/chuyaji/services/api/internal/growthref"
	"github.com/dochaocn/chuyaji/services/api/internal/middleware"
	"github.com/dochaocn/chuyaji/services/api/internal/model"
	"github.com/gin-gonic/gin"
)

type growthPointOut struct {
	RecordID   uint64    `json:"record_id"`
	AgeDays    int       `json:"age_days"`
	OccurredAt time.Time `json:"occurred_at"`
	Value      float64   `json:"value"`
	Unit       string    `json:"unit"`
	Percentile *float64  `json:"percentile,omitempty"`
}

func (h *Handler) BabyGrowthSeries(c *gin.Context) {
	uid, ok := middleware.UserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	babyID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad baby id"})
		return
	}
	ok2, err := h.canAccessBaby(uid, babyID)
	if err != nil || !ok2 {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	var baby model.Baby
	if err := h.DB.First(&baby, babyID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, h.buildGrowthSeriesResponse(babyID, &baby))
}

func (h *Handler) buildGrowthSeriesResponse(babyID uint64, baby *model.Baby) gin.H {
	emptySeries := gin.H{
		"weight": []growthPointOut{},
		"height": []growthPointOut{},
		"head":   []growthPointOut{},
	}
	out := gin.H{
		"baby_id":    babyID,
		"birth_date": baby.BirthDate,
		"gender":     baby.Gender,
		"standard":   "who",
		"series":     emptySeries,
		"reference":  gin.H{"weight": []growthref.Point{}, "height": []growthref.Point{}, "head": []growthref.Point{}},
	}
	if baby.BirthDate == nil {
		out["hint"] = "请先在宝宝档案填写出生日期与性别，以显示 WHO 参考曲线与百分位"
		return out
	}

	sex := growthref.ParseSex(baby.Gender)
	var rows []model.Record
	_ = h.DB.Where("baby_id = ? AND record_type = ?", babyID, "growth").
		Order("occurred_at ASC, id ASC").
		Find(&rows).Error

	weight := []growthPointOut{}
	height := []growthPointOut{}
	head := []growthPointOut{}
	maxAge := 0
	for i := range rows {
		var payload map[string]any
		if err := json.Unmarshal(rows[i].Payload, &payload); err != nil {
			continue
		}
		ageDays := int(rows[i].OccurredAt.Sub(*baby.BirthDate).Hours() / 24)
		if ageDays > maxAge {
			maxAge = ageDays
		}
		if v, ok := numericPayloadValue(payload, "weight_g"); ok {
			val := v / 1000
			weight = append(weight, growthPointOut{
				RecordID: rows[i].ID, AgeDays: ageDays, OccurredAt: rows[i].OccurredAt, Value: val, Unit: "kg",
				Percentile: growthref.Percentile(sex, growthref.MetricWeight, ageDays, val),
			})
		}
		if v, ok := numericPayloadValue(payload, "height_cm"); ok {
			height = append(height, growthPointOut{
				RecordID: rows[i].ID, AgeDays: ageDays, OccurredAt: rows[i].OccurredAt, Value: v, Unit: "cm",
				Percentile: growthref.Percentile(sex, growthref.MetricHeight, ageDays, v),
			})
		}
		if v, ok := numericPayloadValue(payload, "head_circumference_cm"); ok {
			head = append(head, growthPointOut{
				RecordID: rows[i].ID, AgeDays: ageDays, OccurredAt: rows[i].OccurredAt, Value: v, Unit: "cm",
				Percentile: growthref.Percentile(sex, growthref.MetricHead, ageDays, v),
			})
		}
	}
	if maxAge < 180 {
		maxAge = 365
	}
	refMax := maxAge + 60
	if refMax < 365 {
		refMax = 365
	}
	out["series"] = gin.H{"weight": weight, "height": height, "head": head}
	out["reference"] = gin.H{
		"weight": growthref.Curves(sex, growthref.MetricWeight, refMax),
		"height": growthref.Curves(sex, growthref.MetricHeight, refMax),
		"head":   growthref.Curves(sex, growthref.MetricHead, refMax),
	}
	if baby.Gender == "" {
		out["hint"] = "未填写性别时默认按男宝 WHO 参考曲线；可在档案中补充性别以更准确"
	}
	return out
}

func numericPayloadValue(values map[string]any, key string) (float64, bool) {
	raw, ok := values[key]
	if !ok || raw == nil {
		return 0, false
	}
	switch v := raw.(type) {
	case float64:
		return v, true
	case int:
		return float64(v), true
	case json.Number:
		f, err := v.Float64()
		return f, err == nil
	default:
		return 0, false
	}
}
