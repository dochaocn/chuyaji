package router

import (
	"net/http"

	"github.com/dochaocn/chuyaji/services/api/internal/apidocs"
	"github.com/dochaocn/chuyaji/services/api/internal/config"
	"github.com/dochaocn/chuyaji/services/api/internal/handler"
	"github.com/dochaocn/chuyaji/services/api/internal/middleware"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
)

func New(h *handler.Handler, cfg *config.Config) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
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

	authed.GET("/families", h.ListFamilies)
	authed.POST("/families", h.CreateFamily)
	authed.POST("/families/join", h.JoinFamily)
	authed.GET("/families/:id", h.GetFamily)

	authed.GET("/babies", h.ListBabies)
	authed.POST("/babies", h.CreateBaby)
	authed.GET("/babies/:id", h.GetBaby)
	authed.PATCH("/babies/:id", h.PatchBaby)
	authed.DELETE("/babies/:id", h.DeleteBaby)

	authed.GET("/babies/:id/records", h.ListRecords)
	authed.POST("/babies/:id/records", h.CreateRecord)
	authed.GET("/records/:id", h.GetRecord)
	authed.PATCH("/records/:id", h.PatchRecord)
	authed.DELETE("/records/:id", h.DeleteRecord)

	authed.GET("/records/:id/attachments", h.ListAttachments)
	authed.POST("/records/:id/attachments", h.CreateAttachment)
	authed.DELETE("/attachments/:id", h.DeleteAttachment)

	upload := v1.Group("")
	upload.Use(middleware.JWT(cfg.JWTSecret))
	upload.Use(maxBody(32 << 20))
	upload.POST("/records/:id/attachments/upload", h.UploadAttachment)

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
