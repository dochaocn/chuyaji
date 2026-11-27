package router

import (
	"net/http"
	"time"

	"github.com/dochaocn/chuyaji/services/api/internal/apidocs"
	"github.com/dochaocn/chuyaji/services/api/internal/config"
	"github.com/dochaocn/chuyaji/services/api/internal/handler"
	"github.com/dochaocn/chuyaji/services/api/internal/middleware"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
)

func New(h *handler.Handler, cfg *config.Config) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{
			"http://localhost:5173",
			"http://localhost:5174",
			"http://127.0.0.1:5173",
			"http://127.0.0.1:5174",
		},
		AllowMethods: []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"Origin", "Content-Type", "Accept", "Authorization"},
		MaxAge:       12 * time.Hour,
	}))
	r.Use(gzip.Gzip(gzip.DefaultCompression))
	apidocs.Register(r)

	r.GET("/healthz", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})
	r.GET("/api/v1/p/:token", h.PublicAttachment)

	v1 := r.Group("/api/v1")
	v1.POST("/auth/wechat", maxBody(1<<20), h.AuthWechat)

	authed := v1.Group("")
	authed.Use(middleware.JWT(cfg.JWTSecret))
	authed.Use(maxBody(1 << 20))
	authed.GET("/me", h.Me)
	authed.PATCH("/me", h.PatchMe)

	authed.GET("/babies", h.ListBabies)
	authed.POST("/babies", h.CreateBaby)
	authed.GET("/babies/:id", h.GetBaby)
	authed.PATCH("/babies/:id", h.PatchBaby)
	authed.DELETE("/babies/:id", h.DeleteBaby)
	authed.GET("/dashboard/baby", h.BabyDashboard)

	authed.GET("/babies/:id/records", h.ListRecords)
	authed.GET("/babies/:id/records/latest", h.LatestRecord)
	authed.GET("/babies/:id/growth-series", h.BabyGrowthSeries)
	authed.GET("/babies/:id/attachments", h.ListBabyAttachments)
	authed.POST("/babies/:id/records", h.CreateRecord)
	authed.GET("/records/:id", h.GetRecord)
	authed.PATCH("/records/:id", h.PatchRecord)
	authed.DELETE("/records/:id", h.DeleteRecord)

	authed.GET("/records/:id/attachments", h.ListAttachments)
	authed.POST("/records/:id/attachments", h.CreateAttachment)
	authed.GET("/mothers", h.ListMothers)
	authed.POST("/mothers", h.CreateMother)
	authed.GET("/mothers/:id", h.GetMother)
	authed.PATCH("/mothers/:id", h.PatchMother)
	authed.GET("/dashboard/mother", h.MotherDashboard)
	authed.GET("/mothers/:id/records", h.ListMotherRecords)
	authed.GET("/mothers/:id/records/latest", h.LatestMotherRecord)
	authed.POST("/mothers/:id/records", h.CreateMotherRecord)
	authed.GET("/mother-records/:id", h.GetMotherRecord)
	authed.PATCH("/mother-records/:id", h.PatchMotherRecord)
	authed.DELETE("/mother-records/:id", h.DeleteMotherRecord)
	authed.GET("/mother-records/:id/attachments", h.ListMotherAttachments)
	authed.POST("/mother-records/:id/attachments", h.CreateMotherAttachment)
	authed.DELETE("/attachments/:id", h.DeleteAttachment)
	authed.GET("/reminders", h.ListReminders)
	authed.PATCH("/reminders/:id", h.PatchReminder)
	authed.DELETE("/reminders/:id", h.DeleteReminder)

	authed.GET("/families/current", h.GetCurrentFamily)
	authed.POST("/families/current", h.CreateCurrentFamily)
	authed.PATCH("/families/current", h.PatchCurrentFamily)
	authed.POST("/families/current/leave", h.LeaveCurrentFamily)
	authed.POST("/families/current/invites", h.CreateFamilyInvite)
	authed.GET("/invites/:token/preview", h.PreviewFamilyInvite)
	authed.POST("/invites/:token/accept", h.AcceptFamilyInvite)
	authed.POST("/families/current/members/:userId/transfer-owner", h.TransferFamilyOwner)
	authed.PATCH("/families/current/members/:userId", h.PatchFamilyMember)
	authed.PATCH("/families/current/members/:userId/nickname", h.PatchFamilyMemberNickname)
	authed.DELETE("/families/current/members/:userId", h.DeleteFamilyMember)

	upload := v1.Group("")
	upload.Use(middleware.JWT(cfg.JWTSecret))
	upload.Use(maxBody(32 << 20))
	upload.POST("/records/:id/attachments/upload", h.UploadAttachment)
	upload.POST("/mother-records/:id/attachments/upload", h.UploadMotherAttachment)

	// Admin routes
	admin := v1.Group("/admin")
	admin.POST("/login", h.AdminLogin)

	adminAuth := admin.Group("")
	adminAuth.Use(middleware.Admin(cfg.JWTSecret))
	adminAuth.GET("/overview", h.AdminOverview)
	adminAuth.GET("/users", h.AdminListUsers)
	adminAuth.GET("/users/:id", h.AdminGetUser)
	adminAuth.DELETE("/users/:id", h.AdminDeleteUser)
	adminAuth.GET("/babies", h.AdminListBabies)
	adminAuth.GET("/mothers", h.AdminListMothers)
	adminAuth.GET("/records", h.AdminListRecords)
	adminAuth.GET("/records/:id", h.AdminGetRecord)
	adminAuth.GET("/records/:id/attachments", h.AdminListRecordAttachments)
	adminAuth.DELETE("/records/:id", h.AdminDeleteRecord)
	adminAuth.GET("/mother-records", h.AdminListMotherRecords)
	adminAuth.GET("/mother-records/:id", h.AdminGetMotherRecord)
	adminAuth.GET("/mother-records/:id/attachments", h.AdminListMotherRecordAttachments)
	adminAuth.DELETE("/mother-records/:id", h.AdminDeleteMotherRecord)
	adminAuth.GET("/reminders", h.AdminListReminders)
	adminAuth.GET("/reminders/stats", h.AdminReminderStats)
	adminAuth.GET("/analytics/records", h.AdminAnalyticsRecords)
	adminAuth.GET("/analytics/activity", h.AdminAnalyticsActivity)
	adminAuth.GET("/analytics/growth", h.AdminAnalyticsGrowth)

	return r
}

func maxBody(n int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Body != nil {
			c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, n)
		}
		c.Next()
	}
}
