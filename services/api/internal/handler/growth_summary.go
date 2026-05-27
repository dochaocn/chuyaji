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

type growthPointOut struct {
	RecordID   uint64    `json:"record_id"`
	AgeDays    int       `json:"age_days"`
	OccurredAt time.Time `json:"occurred_at"`
	Value      float64   `json:"value"`
	Unit       string    `json:"unit"`
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
	series := gin.H{
		"weight": []growthPointOut{},
		"height": []growthPointOut{},
		"head":   []growthPointOut{},
	}
	if baby.BirthDate == nil {
		c.JSON(http.StatusOK, gin.H{"baby_id": babyID, "birth_date": nil, "series": series})
		return
	}

	var rows []model.Record
	if err := h.DB.Where("baby_id = ? AND record_type = ?", babyID, "growth").
		Order("occurred_at ASC, id ASC").
		Find(&rows).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "query"})
		return
	}
	weight := []growthPointOut{}
	height := []growthPointOut{}
	head := []growthPointOut{}
	for i := range rows {
		var payload map[string]any
		if err := json.Unmarshal(rows[i].Payload, &payload); err != nil {
			continue
		}
		ageDays := int(rows[i].OccurredAt.Sub(*baby.BirthDate).Hours() / 24)
		if v, ok := numericPayloadValue(payload, "weight_g"); ok {
			weight = append(weight, growthPointOut{RecordID: rows[i].ID, AgeDays: ageDays, OccurredAt: rows[i].OccurredAt, Value: v / 1000, Unit: "kg"})
		}
		if v, ok := numericPayloadValue(payload, "height_cm"); ok {
			height = append(height, growthPointOut{RecordID: rows[i].ID, AgeDays: ageDays, OccurredAt: rows[i].OccurredAt, Value: v, Unit: "cm"})
		}
		if v, ok := numericPayloadValue(payload, "head_circumference_cm"); ok {
			head = append(head, growthPointOut{RecordID: rows[i].ID, AgeDays: ageDays, OccurredAt: rows[i].OccurredAt, Value: v, Unit: "cm"})
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"baby_id":    babyID,
		"birth_date": baby.BirthDate,
		"series":     gin.H{"weight": weight, "height": height, "head": head},
	})
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
	default:
		return 0, false
	}
}
