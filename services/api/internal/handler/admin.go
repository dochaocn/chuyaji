package handler

import (
	"crypto/subtle"
	"net/http"
	"strconv"
	"time"

	"github.com/dochaocn/chuyaji/services/api/internal/auth"
	"github.com/dochaocn/chuyaji/services/api/internal/model"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// --- Login ---

type adminLoginReq struct {
	Password string `json:"password" binding:"required"`
}

func (h *Handler) AdminLogin(c *gin.Context) {
	if h.Cfg.AdminPassword == "" {
		c.JSON(http.StatusForbidden, gin.H{"error": "admin not configured"})
		return
	}
	var req adminLoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}
	if subtle.ConstantTimeCompare([]byte(req.Password), []byte(h.Cfg.AdminPassword)) != 1 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid password"})
		return
	}
	ttl := 8 * time.Hour
	token, err := auth.SignAdminJWT(h.Cfg.JWTSecret, ttl)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "sign token"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": token, "expires_in": int64(ttl.Seconds())})
}

// --- Overview ---

func (h *Handler) AdminOverview(c *gin.Context) {
	var userCount, babyCount, motherCount, recordCount, motherRecordCount, reminderPendingCount int64

	h.DB.Model(&model.User{}).Count(&userCount)
	h.DB.Model(&model.Baby{}).Count(&babyCount)
	h.DB.Model(&model.Mother{}).Count(&motherCount)
	h.DB.Model(&model.Record{}).Count(&recordCount)
	h.DB.Model(&model.MotherRecord{}).Count(&motherRecordCount)
	h.DB.Model(&model.Reminder{}).Where("status = ?", "pending").Count(&reminderPendingCount)

	var recentUsers []model.User
	h.DB.Order("created_at DESC").Limit(5).Find(&recentUsers)
	userOuts := make([]gin.H, 0, len(recentUsers))
	for _, u := range recentUsers {
		userOuts = append(userOuts, gin.H{"id": u.ID, "open_id": u.OpenID, "nickname": u.Nickname, "avatar_url": u.AvatarURL})
	}

	var recentRecords []model.Record
	h.DB.Order("created_at DESC").Limit(5).Find(&recentRecords)
	recOuts := make([]recordOut, 0, len(recentRecords))
	for i := range recentRecords {
		recOuts = append(recOuts, recordToOut(&recentRecords[i]))
	}

	c.JSON(http.StatusOK, gin.H{
		"user_count":            userCount,
		"baby_count":            babyCount,
		"mother_count":          motherCount,
		"record_count":          recordCount,
		"mother_record_count":   motherRecordCount,
		"reminder_pending_count": reminderPendingCount,
		"recent_users":          userOuts,
		"recent_records":        recOuts,
	})
}

// --- Users ---

type adminUserItem struct {
	ID          uint64 `json:"id"`
	OpenID      string `json:"open_id"`
	Nickname    string `json:"nickname"`
	AvatarURL   string `json:"avatar_url"`
	BabyCount   int64  `json:"baby_count"`
	MotherCount int64  `json:"mother_count"`
	CreatedAt   time.Time `json:"created_at"`
}

func (h *Handler) AdminListUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	q := h.DB.Model(&model.User{})
	if kw := c.Query("q"); kw != "" {
		like := "%" + kw + "%"
		q = q.Where("nickname LIKE ?", like)
	}

	var total int64
	q.Count(&total)

	var users []model.User
	q.Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&users)

	items := make([]adminUserItem, 0, len(users))
	for _, u := range users {
		var bc, mc int64
		h.DB.Model(&model.Baby{}).Where("user_id = ?", u.ID).Count(&bc)
		h.DB.Model(&model.Mother{}).Where("user_id = ?", u.ID).Count(&mc)
		items = append(items, adminUserItem{
			ID: u.ID, OpenID: u.OpenID, Nickname: u.Nickname, AvatarURL: u.AvatarURL,
			BabyCount: bc, MotherCount: mc, CreatedAt: u.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{"items": items, "total": total, "page": page, "page_size": pageSize})
}

func (h *Handler) AdminGetUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad id"})
		return
	}
	var u model.User
	if err := h.DB.First(&u, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	var babies []model.Baby
	familyIDs := []uint64{}
	var memberships []model.FamilyMember
	h.DB.Where("user_id = ?", id).Find(&memberships)
	membersOut := make([]gin.H, 0)
	for _, m := range memberships {
		familyIDs = append(familyIDs, m.FamilyID)
		var fam model.Family
		_ = h.DB.First(&fam, m.FamilyID)
		var allMembers []model.FamilyMember
		h.DB.Where("family_id = ?", m.FamilyID).Find(&allMembers)
		memberItems := make([]gin.H, 0, len(allMembers))
		for _, am := range allMembers {
			var mu model.User
			_ = h.DB.First(&mu, am.UserID)
			memberItems = append(memberItems, gin.H{
				"user_id": am.UserID, "nickname": mu.Nickname, "role": am.Role,
			})
		}
		membersOut = append(membersOut, gin.H{
			"family_id": fam.ID, "family_name": fam.Name, "my_role": m.Role, "members": memberItems,
		})
	}
	if len(familyIDs) > 0 {
		h.DB.Where("family_id IN ?", familyIDs).Order("id ASC").Find(&babies)
	} else {
		h.DB.Where("user_id = ?", id).Order("id ASC").Find(&babies)
	}
	babyOuts := make([]babyOut, 0, len(babies))
	for i := range babies {
		babyOuts = append(babyOuts, babyToOut(&babies[i]))
	}

	var mothers []model.Mother
	if len(familyIDs) > 0 {
		h.DB.Where("family_id IN ?", familyIDs).Order("id ASC").Find(&mothers)
	} else {
		h.DB.Where("user_id = ?", id).Order("id ASC").Find(&mothers)
	}
	motherOuts := make([]motherOut, 0, len(mothers))
	for i := range mothers {
		motherOuts = append(motherOuts, motherToOut(&mothers[i]))
	}

	c.JSON(http.StatusOK, gin.H{
		"user":     gin.H{"id": u.ID, "open_id": u.OpenID, "nickname": u.Nickname, "avatar_url": u.AvatarURL, "created_at": u.CreatedAt},
		"babies":   babyOuts,
		"mothers":  motherOuts,
		"families": membersOut,
	})
}

func (h *Handler) AdminDeleteUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad id"})
		return
	}
	if err := h.DB.Transaction(func(tx *gorm.DB) error {
		// Delete reminders
		if err := tx.Where("user_id = ?", id).Delete(&model.Reminder{}).Error; err != nil {
			return err
		}
		// Delete baby records and their attachments
		var babyIDs []uint64
		tx.Model(&model.Baby{}).Where("user_id = ?", id).Pluck("id", &babyIDs)
		for _, bid := range babyIDs {
			tx.Where("owner_type = ? AND owner_id IN (SELECT id FROM records WHERE baby_id = ?)", "baby_record", bid).Delete(&model.Attachment{})
			tx.Where("baby_id = ?", bid).Delete(&model.Record{})
		}
		tx.Where("user_id = ?", id).Delete(&model.Baby{})
		// Delete mother records and their attachments
		var motherIDs []uint64
		tx.Model(&model.Mother{}).Where("user_id = ?", id).Pluck("id", &motherIDs)
		for _, mid := range motherIDs {
			tx.Where("owner_type = ? AND owner_id IN (SELECT id FROM mother_records WHERE mother_id = ?)", "mother_record", mid).Delete(&model.Attachment{})
			tx.Where("mother_id = ?", mid).Delete(&model.MotherRecord{})
		}
		tx.Where("user_id = ?", id).Delete(&model.Mother{})
		return tx.Delete(&model.User{}, id).Error
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "delete"})
		return
	}
	c.Status(http.StatusNoContent)
}

// --- Babies ---

func (h *Handler) AdminListBabies(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	q := h.DB.Model(&model.Baby{})
	if kw := c.Query("q"); kw != "" {
		q = q.Where("nickname LIKE ?", "%"+kw+"%")
	}
	if uid := c.Query("user_id"); uid != "" {
		q = q.Where("user_id = ?", uid)
	}

	var total int64
	q.Count(&total)

	var babies []model.Baby
	q.Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&babies)

	items := make([]babyOut, 0, len(babies))
	for i := range babies {
		items = append(items, babyToOut(&babies[i]))
	}

	c.JSON(http.StatusOK, gin.H{"items": items, "total": total, "page": page, "page_size": pageSize})
}

// --- Mothers ---

func (h *Handler) AdminListMothers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	q := h.DB.Model(&model.Mother{})
	if kw := c.Query("q"); kw != "" {
		q = q.Where("name LIKE ?", "%"+kw+"%")
	}
	if uid := c.Query("user_id"); uid != "" {
		q = q.Where("user_id = ?", uid)
	}

	var total int64
	q.Count(&total)

	var mothers []model.Mother
	q.Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&mothers)

	items := make([]motherOut, 0, len(mothers))
	for i := range mothers {
		items = append(items, motherToOut(&mothers[i]))
	}

	c.JSON(http.StatusOK, gin.H{"items": items, "total": total, "page": page, "page_size": pageSize})
}

// --- Records ---

func (h *Handler) AdminListRecords(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	q := h.DB.Model(&model.Record{})
	if bid := c.Query("baby_id"); bid != "" {
		q = q.Where("baby_id = ?", bid)
	}
	if rt := c.Query("record_type"); rt != "" {
		q = q.Where("record_type = ?", rt)
	}
	if phase := c.Query("phase"); phase != "" {
		q = q.Where("phase = ?", phase)
	}
	if uid := c.Query("user_id"); uid != "" {
		var babyIDs []uint64
		h.DB.Model(&model.Baby{}).Where("user_id = ?", uid).Pluck("id", &babyIDs)
		if len(babyIDs) == 0 {
			c.JSON(http.StatusOK, gin.H{"items": []recordOut{}, "total": 0, "page": page, "page_size": pageSize})
			return
		}
		q = q.Where("baby_id IN ?", babyIDs)
	}
	if from := c.Query("from"); from != "" {
		t, err := parseQueryDateTime(from, false)
		if err == nil {
			q = q.Where("occurred_at >= ?", t)
		}
	}
	if to := c.Query("to"); to != "" {
		t, err := parseQueryDateTime(to, true)
		if err == nil {
			q = q.Where("occurred_at < ?", t)
		}
	}

	var total int64
	q.Count(&total)

	var records []model.Record
	q.Order("occurred_at DESC, id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&records)

	items := make([]recordOut, 0, len(records))
	for i := range records {
		items = append(items, recordToOut(&records[i]))
	}

	c.JSON(http.StatusOK, gin.H{"items": items, "total": total, "page": page, "page_size": pageSize})
}

func (h *Handler) AdminGetRecord(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad id"})
		return
	}
	var r model.Record
	if err := h.DB.First(&r, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, recordToOut(&r))
}

func (h *Handler) AdminDeleteRecord(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad id"})
		return
	}
	if err := h.DB.Transaction(func(tx *gorm.DB) error {
		tx.Where("owner_type = ? AND owner_id = ?", "baby_record", id).Delete(&model.Attachment{})
		tx.Where("source_type = ? AND source_id = ?", "baby_record", id).Delete(&model.Reminder{})
		return tx.Delete(&model.Record{}, id).Error
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "delete"})
		return
	}
	c.Status(http.StatusNoContent)
}

// --- Mother Records ---

func (h *Handler) AdminListMotherRecords(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	q := h.DB.Model(&model.MotherRecord{})
	if mid := c.Query("mother_id"); mid != "" {
		q = q.Where("mother_id = ?", mid)
	}
	if rt := c.Query("record_type"); rt != "" {
		q = q.Where("record_type = ?", rt)
	}
	if uid := c.Query("user_id"); uid != "" {
		var motherIDs []uint64
		h.DB.Model(&model.Mother{}).Where("user_id = ?", uid).Pluck("id", &motherIDs)
		if len(motherIDs) == 0 {
			c.JSON(http.StatusOK, gin.H{"items": []motherRecordOut{}, "total": 0, "page": page, "page_size": pageSize})
			return
		}
		q = q.Where("mother_id IN ?", motherIDs)
	}
	if from := c.Query("from"); from != "" {
		t, err := parseQueryDateTime(from, false)
		if err == nil {
			q = q.Where("occurred_at >= ?", t)
		}
	}
	if to := c.Query("to"); to != "" {
		t, err := parseQueryDateTime(to, true)
		if err == nil {
			q = q.Where("occurred_at < ?", t)
		}
	}

	var total int64
	q.Count(&total)

	var records []model.MotherRecord
	q.Order("occurred_at DESC, id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&records)

	items := make([]motherRecordOut, 0, len(records))
	for i := range records {
		items = append(items, motherRecordToOut(&records[i]))
	}

	c.JSON(http.StatusOK, gin.H{"items": items, "total": total, "page": page, "page_size": pageSize})
}

func (h *Handler) AdminGetMotherRecord(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad id"})
		return
	}
	var r model.MotherRecord
	if err := h.DB.First(&r, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, motherRecordToOut(&r))
}

func (h *Handler) AdminDeleteMotherRecord(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad id"})
		return
	}
	if err := h.DB.Transaction(func(tx *gorm.DB) error {
		tx.Where("owner_type = ? AND owner_id = ?", "mother_record", id).Delete(&model.Attachment{})
		tx.Where("source_type = ? AND source_id = ?", "mother_record", id).Delete(&model.Reminder{})
		return tx.Delete(&model.MotherRecord{}, id).Error
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "delete"})
		return
	}
	c.Status(http.StatusNoContent)
}

// --- Reminders ---

func (h *Handler) AdminListReminders(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	q := h.DB.Model(&model.Reminder{})
	if status := c.Query("status"); status != "" {
		q = q.Where("status = ?", status)
	}
	if category := c.Query("category"); category != "" {
		q = q.Where("category = ?", category)
	}
	if ownerType := c.Query("owner_type"); ownerType != "" {
		q = q.Where("owner_type = ?", ownerType)
	}

	var total int64
	q.Count(&total)

	var reminders []model.Reminder
	q.Order("due_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&reminders)

	// Collect unique owner IDs by type for batch lookup
	babyIDs := map[uint64]bool{}
	motherIDs := map[uint64]bool{}
	for i := range reminders {
		switch reminders[i].OwnerType {
		case "baby":
			babyIDs[reminders[i].OwnerID] = true
		case "mother":
			motherIDs[reminders[i].OwnerID] = true
		}
	}

	babyNames := map[uint64]string{}
	if len(babyIDs) > 0 {
		ids := make([]uint64, 0, len(babyIDs))
		for id := range babyIDs {
			ids = append(ids, id)
		}
		var babies []model.Baby
		h.DB.Where("id IN ?", ids).Find(&babies)
		for _, b := range babies {
			babyNames[b.ID] = b.Nickname
		}
	}

	motherNames := map[uint64]string{}
	if len(motherIDs) > 0 {
		ids := make([]uint64, 0, len(motherIDs))
		for id := range motherIDs {
			ids = append(ids, id)
		}
		var mothers []model.Mother
		h.DB.Where("id IN ?", ids).Find(&mothers)
		for _, m := range mothers {
			motherNames[m.ID] = m.Name
		}
	}

	items := make([]reminderOut, 0, len(reminders))
	for i := range reminders {
		out := reminderToOut(&reminders[i])
		switch reminders[i].OwnerType {
		case "baby":
			out.OwnerName = babyNames[reminders[i].OwnerID]
		case "mother":
			out.OwnerName = motherNames[reminders[i].OwnerID]
		}
		items = append(items, out)
	}

	c.JSON(http.StatusOK, gin.H{"items": items, "total": total, "page": page, "page_size": pageSize})
}

type reminderStatsOut struct {
	Total     int64            `json:"total"`
	ByStatus  map[string]int64 `json:"by_status"`
	ByCategory map[string]int64 `json:"by_category"`
	ByOwnerType map[string]int64 `json:"by_owner_type"`
	OverdueCount int64          `json:"overdue_count"`
}

func (h *Handler) AdminReminderStats(c *gin.Context) {
	var total int64
	h.DB.Model(&model.Reminder{}).Count(&total)

	byStatus := map[string]int64{}
	var statusRows []struct {
		Status string
		Count  int64
	}
	h.DB.Model(&model.Reminder{}).Select("status, count(*) as count").Group("status").Scan(&statusRows)
	for _, r := range statusRows {
		byStatus[r.Status] = r.Count
	}

	byCategory := map[string]int64{}
	var catRows []struct {
		Category string
		Count    int64
	}
	h.DB.Model(&model.Reminder{}).Select("category, count(*) as count").Group("category").Scan(&catRows)
	for _, r := range catRows {
		byCategory[r.Category] = r.Count
	}

	byOwnerType := map[string]int64{}
	var ownerRows []struct {
		OwnerType string
		Count     int64
	}
	h.DB.Model(&model.Reminder{}).Select("owner_type, count(*) as count").Group("owner_type").Scan(&ownerRows)
	for _, r := range ownerRows {
		byOwnerType[r.OwnerType] = r.Count
	}

	var overdueCount int64
	h.DB.Model(&model.Reminder{}).Where("status = ? AND due_at < ?", "pending", time.Now()).Count(&overdueCount)

	c.JSON(http.StatusOK, reminderStatsOut{
		Total:        total,
		ByStatus:     byStatus,
		ByCategory:   byCategory,
		ByOwnerType:  byOwnerType,
		OverdueCount: overdueCount,
	})
}

// --- Analytics ---

type dailyCount struct {
	Date          string `json:"date"`
	BabyRecords   int64  `json:"baby_records"`
	MotherRecords int64  `json:"mother_records"`
}

type typeCount struct {
	RecordType string `json:"record_type"`
	Count      int64  `json:"count"`
}

func (h *Handler) AdminAnalyticsRecords(c *gin.Context) {
	days, _ := strconv.Atoi(c.DefaultQuery("days", "30"))
	if days <= 0 || days > 365 {
		days = 30
	}
	since := time.Now().AddDate(0, 0, -days)

	var babyRows []dailyCount
	h.DB.Model(&model.Record{}).
		Select("date(occurred_at) as date, count(*) as baby_records").
		Where("occurred_at >= ?", since).
		Group("date(occurred_at)").
		Order("date ASC").
		Scan(&babyRows)

	var motherRows []struct {
		Date          string
		MotherRecords int64
	}
	h.DB.Model(&model.MotherRecord{}).
		Select("date(occurred_at) as date, count(*) as mother_records").
		Where("occurred_at >= ?", since).
		Group("date(occurred_at)").
		Order("date ASC").
		Scan(&motherRows)

	motherMap := map[string]int64{}
	for _, r := range motherRows {
		motherMap[r.Date] = r.MotherRecords
	}

	merged := make([]dailyCount, 0, len(babyRows))
	for _, r := range babyRows {
		merged = append(merged, dailyCount{
			Date: r.Date, BabyRecords: r.BabyRecords, MotherRecords: motherMap[r.Date],
		})
	}

	var typeDist []typeCount
	h.DB.Model(&model.Record{}).
		Select("record_type, count(*) as count").
		Where("occurred_at >= ?", since).
		Group("record_type").
		Order("count DESC").
		Scan(&typeDist)

	c.JSON(http.StatusOK, gin.H{"daily_counts": merged, "type_distribution": typeDist})
}

type activeUserCount struct {
	Date  string `json:"date"`
	Count int64  `json:"count"`
}

func (h *Handler) AdminAnalyticsActivity(c *gin.Context) {
	days, _ := strconv.Atoi(c.DefaultQuery("days", "30"))
	if days <= 0 || days > 365 {
		days = 30
	}
	since := time.Now().AddDate(0, 0, -days)

	// Daily active users from baby records (via baby -> user_id)
	var babyActive []activeUserCount
	h.DB.Raw(`
		SELECT date(r.occurred_at) as date, COUNT(DISTINCT b.user_id) as count
		FROM records r
		JOIN babies b ON b.id = r.baby_id
		WHERE r.occurred_at >= ?
		GROUP BY date(r.occurred_at)
		ORDER BY date ASC
	`, since).Scan(&babyActive)

	// Daily active users from mother records (via mother -> user_id)
	var motherActive []activeUserCount
	h.DB.Raw(`
		SELECT date(r.occurred_at) as date, COUNT(DISTINCT m.user_id) as count
		FROM mother_records r
		JOIN mothers m ON m.id = r.mother_id
		WHERE r.occurred_at >= ?
		GROUP BY date(r.occurred_at)
		ORDER BY date ASC
	`, since).Scan(&motherActive)

	// Merge: union distinct user counts per day
	dayMap := map[string]map[uint64]bool{}
	for _, row := range babyActive {
		if dayMap[row.Date] == nil {
			dayMap[row.Date] = map[uint64]bool{}
		}
	}
	// Re-query with user IDs for proper merge
	var babyUserDays []struct {
		Date   string
		UserID uint64
	}
	h.DB.Raw(`
		SELECT date(r.occurred_at) as date, b.user_id
		FROM records r
		JOIN babies b ON b.id = r.baby_id
		WHERE r.occurred_at >= ?
		GROUP BY date(r.occurred_at), b.user_id
	`, since).Scan(&babyUserDays)

	var motherUserDays []struct {
		Date   string
		UserID uint64
	}
	h.DB.Raw(`
		SELECT date(r.occurred_at) as date, m.user_id
		FROM mother_records r
		JOIN mothers m ON m.id = r.mother_id
		WHERE r.occurred_at >= ?
		GROUP BY date(r.occurred_at), m.user_id
	`, since).Scan(&motherUserDays)

	userDayMap := map[string]map[uint64]bool{}
	for _, row := range babyUserDays {
		if userDayMap[row.Date] == nil {
			userDayMap[row.Date] = map[uint64]bool{}
		}
		userDayMap[row.Date][row.UserID] = true
	}
	for _, row := range motherUserDays {
		if userDayMap[row.Date] == nil {
			userDayMap[row.Date] = map[uint64]bool{}
		}
		userDayMap[row.Date][row.UserID] = true
	}

	dailyActive := make([]activeUserCount, 0, len(userDayMap))
	for date, users := range userDayMap {
		dailyActive = append(dailyActive, activeUserCount{Date: date, Count: int64(len(users))})
	}
	// Sort by date
	for i := 0; i < len(dailyActive); i++ {
		for j := i + 1; j < len(dailyActive); j++ {
			if dailyActive[i].Date > dailyActive[j].Date {
				dailyActive[i], dailyActive[j] = dailyActive[j], dailyActive[i]
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{"daily_active_users": dailyActive})
}

func (h *Handler) AdminAnalyticsGrowth(c *gin.Context) {
	babyID, err := strconv.ParseUint(c.Query("baby_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad baby_id"})
		return
	}

	var baby model.Baby
	if err := h.DB.First(&baby, babyID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "baby not found"})
		return
	}
	c.JSON(http.StatusOK, h.buildGrowthSeriesResponse(babyID, &baby))
}

// --- Attachments (admin) ---

func (h *Handler) AdminListRecordAttachments(c *gin.Context) {
	recordID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad id"})
		return
	}
	var attachments []model.Attachment
	h.DB.Where("owner_type = ? AND owner_id = ?", "baby_record", recordID).
		Order("sort_order ASC, id ASC").Find(&attachments)
	items := make([]attachmentOut, 0, len(attachments))
	for i := range attachments {
		items = append(items, attachmentToOut(&attachments[i]))
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

func (h *Handler) AdminListMotherRecordAttachments(c *gin.Context) {
	recordID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad id"})
		return
	}
	var attachments []model.Attachment
	h.DB.Where("owner_type = ? AND owner_id = ?", "mother_record", recordID).
		Order("sort_order ASC, id ASC").Find(&attachments)
	items := make([]attachmentOut, 0, len(attachments))
	for i := range attachments {
		items = append(items, attachmentToOut(&attachments[i]))
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}
