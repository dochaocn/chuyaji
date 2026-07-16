package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

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

	c.JSON(http.StatusOK, gin.H{
		"profile":        babyToOut(&baby),
		"phase_summary":  gin.H{"stage": stage, "record_count": len(out)},
		"latest_records": out,
		"growth_summary": growthSummary,
	})
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
