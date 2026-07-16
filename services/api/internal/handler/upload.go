package handler

import (
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/dochaocn/chuyaji/services/api/internal/middleware"
	"github.com/dochaocn/chuyaji/services/api/internal/model"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

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
	ok2, err := h.ensureAttachmentOwnerWrite(uid, ownerID, ownerType)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	if err != nil || !ok2 {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	var babyOrMotherID uint64
	var recordType string
	switch ownerType {
	case "baby_record":
		var rec model.Record
		if err := h.DB.First(&rec, ownerID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "db"})
			}
			return
		}
		babyOrMotherID = rec.BabyID
		recordType = rec.RecordType
	case "mother_record":
		var rec model.MotherRecord
		if err := h.DB.First(&rec, ownerID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "db"})
			}
			return
		}
		babyOrMotherID = rec.MotherID
		recordType = rec.RecordType
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad owner"})
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
	var token string
	for attempt := 0; attempt < 8; attempt++ {
		tk, errGen := attachmentShareToken(babyOrMotherID, ownerID, recordType)
		if errGen != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "name"})
			return
		}
		var cnt int64
		if err := h.DB.Model(&model.Attachment{}).Where("share_token = ?", tk).Count(&cnt).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "db"})
			return
		}
		if cnt == 0 {
			token = tk
			break
		}
	}
	if token == "" {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "name collision"})
		return
	}
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

const maxShareTokenLen = 64

func sanitizeRecordTypeSegment(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == '_', r == '-':
			b.WriteRune(r)
		default:
			b.WriteByte('_')
		}
	}
	out := strings.Trim(b.String(), "_-")
	if out == "" {
		return "type"
	}
	return out
}

func randomAlnum5() (string, error) {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	buf := make([]byte, 5)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	out := make([]byte, 5)
	for i := range out {
		out[i] = letters[int(buf[i])%len(letters)]
	}
	return string(out), nil
}

func truncateUTF8ByMaxBytes(s string, maxBytes int) string {
	if maxBytes <= 0 {
		return ""
	}
	if len(s) <= maxBytes {
		return s
	}
	s = s[:maxBytes]
	for !utf8.ValidString(s) {
		s = s[:len(s)-1]
		if len(s) == 0 {
			return ""
		}
	}
	return s
}

func attachmentShareToken(profileID, recordID uint64, recordType string) (string, error) {
	safe := sanitizeRecordTypeSegment(recordType)
	a := strconv.FormatUint(profileID, 10)
	b := strconv.FormatUint(recordID, 10)
	suf, err := randomAlnum5()
	if err != nil {
		return "", err
	}
	overhead := len(a) + len(b) + len(suf) + 3
	maxSafe := maxShareTokenLen - overhead
	if maxSafe < 1 {
		maxSafe = 1
	}
	safe = truncateUTF8ByMaxBytes(safe, maxSafe)
	return fmt.Sprintf("%s_%s_%s_%s", a, b, safe, suf), nil
}
