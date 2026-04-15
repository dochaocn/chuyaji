package handler

import (
	"errors"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/dochaocn/chuyaji/services/api/internal/middleware"
	"github.com/dochaocn/chuyaji/services/api/internal/model"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UploadAttachment 接收 multipart 字段 file，写入 CHUYAJI_UPLOAD_DIR，并登记附件（外链带随机 token）。
func (h *Handler) UploadAttachment(c *gin.Context) {
	h.uploadAttachmentByOwner(c, "baby_record")
}

func (h *Handler) UploadMotherAttachment(c *gin.Context) {
	h.uploadAttachmentByOwner(c, "mother_record")
}

func (h *Handler) uploadAttachmentByOwner(c *gin.Context, ownerType string) {
	if strings.TrimSpace(h.Cfg.UploadDir) == "" {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "upload disabled: set CHUYAJI_UPLOAD_DIR"})
		return
	}
	uid, ok := middleware.UserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	ownerID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad owner id"})
		return
	}
	ok2, err := h.ensureAttachmentOwnerAccess(uid, ownerID, ownerType)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	if err != nil || !ok2 {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	fh, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing file"})
		return
	}
	src, err := fh.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "open file"})
		return
	}
	defer src.Close()

	ext := strings.ToLower(filepath.Ext(fh.Filename))
	if ext == "" {
		ext = ".bin"
	}
	token := uuid.NewString()
	dirName := "records"
	if ownerType == "mother_record" {
		dirName = "mother-records"
	}
	dir := filepath.Join(h.Cfg.UploadDir, dirName)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "mkdir"})
		return
	}
	filename := token + ext
	dstPath := filepath.Join(dir, filename)
	dst, err := os.Create(dstPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "create dst"})
		return
	}
	n, err := io.Copy(dst, src)
	_ = dst.Close()
	if err != nil {
		_ = os.Remove(dstPath)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "save"})
		return
	}

	base := strings.TrimRight(strings.TrimSpace(h.Cfg.PublicBaseURL), "/")
	if base == "" {
		base = "http://" + c.Request.Host
	}
	publicURL := base + "/api/v1/p/" + token

	tok := token
	a := model.Attachment{
		OwnerType:  ownerType,
		OwnerID:    ownerID,
		URL:        publicURL,
		ThumbURL:   "",
		SortOrder:  0,
		Size:       n,
		ShareToken: &tok,
		LocalPath:  dstPath,
	}
	if err := h.DB.Create(&a).Error; err != nil {
		_ = os.Remove(dstPath)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db"})
		return
	}

	c.JSON(http.StatusCreated, attachmentToOut(&a))
}

// PublicAttachment 通过不可猜测 token 读取本机托管附件（小程序 image 可直接使用该 URL）。
func (h *Handler) PublicAttachment(c *gin.Context) {
	token := c.Param("token")
	if token == "" {
		c.Status(http.StatusNotFound)
		return
	}
	var a model.Attachment
	if err := h.DB.Where("share_token = ?", token).First(&a).Error; err != nil {
		c.Status(http.StatusNotFound)
		return
	}
	if a.LocalPath == "" {
		c.Status(http.StatusNotFound)
		return
	}
	ct := mime.TypeByExtension(filepath.Ext(a.LocalPath))
	if ct == "" {
		ct = "application/octet-stream"
	}
	c.Header("Content-Type", ct)
	c.Header("Cache-Control", "public, max-age=86400")
	c.File(a.LocalPath)
}
