package main

import (
	"log"
	"net/http"

	"github.com/dochaocn/chuyaji/services/api/internal/config"
	"github.com/dochaocn/chuyaji/services/api/internal/handler"
	"github.com/dochaocn/chuyaji/services/api/internal/router"
	"github.com/dochaocn/chuyaji/services/api/internal/store"
)

func main() {
	cfg := config.Load()
	db, err := store.Open(cfg.DBPath)
	if err != nil {
		log.Fatalf("db: %v", err)
	}

	h := handler.New(db, cfg)
	r := router.New(h, cfg)

	srv := &http.Server{
		Addr:    cfg.HTTPAddr,
		Handler: r,
	}
	log.Printf("chuyaji api listening on %s", cfg.HTTPAddr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
