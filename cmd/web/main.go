// Command web menjalankan server Sistem Penggajian Konveksi Mitra Jaya.
//
// FILE BERSAMA (frozen) — jangan diubah tanpa PR yang direview seluruh anggota.
// Seluruh modul sudah terdaftar di internal/handler/routes.go.
package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"penggajian/internal/config"
	"penggajian/internal/handler"
	"penggajian/internal/middleware"
)

func main() {
	cfg := config.Load()

	db, err := config.OpenDB(cfg)
	if err != nil {
		log.Fatalf("gagal koneksi database: %v", err)
	}
	defer db.Close()

	render, err := handler.NewRenderer("templates", cfg.IsDev())
	if err != nil {
		log.Fatalf("gagal memuat template: %v", err)
	}

	deps := handler.Deps{
		DB:       db,
		Render:   render,
		Sessions: middleware.NewSessionManager(),
		Cfg:      cfg,
	}

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           handler.Routes(deps),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		log.Printf("server berjalan di http://localhost:%s (env=%s)", cfg.Port, cfg.AppEnv)
		if cfg.DevAutoLoginRole != "" {
			log.Printf("PERINGATAN: DEV_AUTO_LOGIN_ROLE=%s aktif — login dilewati (khusus development)", cfg.DevAutoLoginRole)
		}
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server berhenti: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("shutdown: %v", err)
	}
}
