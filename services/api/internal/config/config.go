package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	HTTPAddr        string
	JWTSecret       string
	DBPath          string
	WechatAppID     string
	WechatAppSecret string
	DevMode         bool
	UploadDir       string
	PublicBaseURL   string
	AdminPassword   string
}

func Load() *Config {
	_ = godotenv.Load()
	_ = godotenv.Load("services/api/.env")
	_ = godotenv.Load(".env")

	cfg := &Config{
		HTTPAddr:        getEnv("CHUYAJI_HTTP_ADDR", ":8282"),
		JWTSecret:       getEnv("CHUYAJI_JWT_SECRET", ""),
		DBPath:          getEnv("CHUYAJI_DB_PATH", "./data/chuyaji.db"),
		WechatAppID:     getEnv("CHUYAJI_WECHAT_APP_ID", ""),
		WechatAppSecret: getEnv("CHUYAJI_WECHAT_APP_SECRET", ""),
		DevMode:         getEnvBool("CHUYAJI_DEV_MODE", false),
		UploadDir:       getEnv("CHUYAJI_UPLOAD_DIR", "./data/uploads"),
		PublicBaseURL:   getEnv("CHUYAJI_PUBLIC_BASE_URL", ""),
		AdminPassword:   getEnv("CHUYAJI_ADMIN_PASSWORD", ""),
	}
	if cfg.JWTSecret == "" {
		log.Fatal("CHUYAJI_JWT_SECRET is required")
	}
	return cfg
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getEnvBool(key string, def bool) bool {
	s := os.Getenv(key)
	if s == "" {
		return def
	}
	b, err := strconv.ParseBool(s)
	if err != nil {
		return def
	}
	return b
}
