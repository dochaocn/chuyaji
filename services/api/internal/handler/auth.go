package handler

import (
	"net/http"
	"strings"
	"time"

	"github.com/dochaocn/chuyaji/services/api/internal/auth"
	"github.com/dochaocn/chuyaji/services/api/internal/middleware"
	"github.com/dochaocn/chuyaji/services/api/internal/model"
	"github.com/dochaocn/chuyaji/services/api/internal/wechat"
	"github.com/gin-gonic/gin"
)

type wechatLoginReq struct {
	Code string `json:"code" binding:"required"`
}

type wechatLoginResp struct {
	Token     string  `json:"token"`
	ExpiresIn int64   `json:"expires_in"`
	User      userOut `json:"user"`
}

type userOut struct {
	ID        uint64 `json:"id"`
	Nickname  string `json:"nickname"`
	AvatarURL string `json:"avatar_url"`
}

func (h *Handler) AuthWechat(c *gin.Context) {
	var req wechatLoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}

	var openID string
	switch {
	case h.Cfg.DevMode && req.Code == "dev":
		openID = "dev-openid"
	default:
		if h.Cfg.WechatAppID == "" || h.Cfg.WechatAppSecret == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "wechat not configured"})
			return
		}
		sess, err := wechat.Code2Session(h.Cfg.WechatAppID, h.Cfg.WechatAppSecret, req.Code)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		openID = sess.OpenID
	}

	var u model.User
	err := h.DB.Where("open_id = ?", openID).First(&u).Error
	if err != nil {
		u = model.User{OpenID: openID}
		if err := h.DB.Create(&u).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "create user"})
			return
		}
	}

	ttl := 30 * 24 * time.Hour
	token, err := auth.SignJWT(h.Cfg.JWTSecret, u.ID, ttl)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "sign token"})
		return
	}

	c.JSON(http.StatusOK, wechatLoginResp{
		Token:     token,
		ExpiresIn: int64(ttl.Seconds()),
		User: userOut{
			ID:        u.ID,
			Nickname:  u.Nickname,
			AvatarURL: u.AvatarURL,
		},
	})
}

func (h *Handler) Me(c *gin.Context) {
	uid, ok := middleware.UserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	var u model.User
	if err := h.DB.First(&u, uid).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	c.JSON(http.StatusOK, userOut{
		ID:        u.ID,
		Nickname:  u.Nickname,
		AvatarURL: u.AvatarURL,
	})
}

type patchMeReq struct {
	Nickname *string `json:"nickname" binding:"omitempty,max=64"`
}

func (h *Handler) PatchMe(c *gin.Context) {
	uid, ok := middleware.UserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	var req patchMeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}
	var u model.User
	if err := h.DB.First(&u, uid).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	if req.Nickname != nil {
		u.Nickname = strings.TrimSpace(*req.Nickname)
	}
	if err := h.DB.Save(&u).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "save"})
		return
	}
	c.JSON(http.StatusOK, userOut{
		ID:        u.ID,
		Nickname:  u.Nickname,
		AvatarURL: u.AvatarURL,
	})
}
