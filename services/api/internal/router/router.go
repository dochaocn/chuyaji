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

	authed.GET("/babies", h.ListBabies)
	authed.POST("/babies", h.CreateBaby)
	authed.GET("/babies/:id", h.GetBaby)
	authed.PATCH("/babies/:id", h.PatchBaby)
	authed.DELETE("/babies/:id", h.DeleteBaby)
	authed.GET("/dashboard/baby", h.BabyDashboard)

	authed.GET("/babies/:id/records", h.ListRecords)
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

	upload := v1.Group("")
	upload.Use(middleware.JWT(cfg.JWTSecret))
	upload.Use(maxBody(32 << 20))
	upload.POST("/records/:id/attachments/upload", h.UploadAttachment)
	upload.POST("/mother-records/:id/attachments/upload", h.UploadMotherAttachment)

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
