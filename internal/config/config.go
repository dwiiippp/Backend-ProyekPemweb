// Package config membaca konfigurasi dari environment / file .env.
// FILE BERSAMA (frozen).
package config

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	AppEnv string
	Port   string

	DBHost string
	DBPort string
	DBUser string
	DBPass string
	DBName string

	// DevAutoLoginRole ("admin" / "pekerja") melewati login saat development
	// agar tiap anggota bisa menguji halamannya tanpa menunggu modul Auth selesai.
	// Kosongkan untuk memakai login sungguhan.
	DevAutoLoginRole string
	CookieSecure     bool
}

func (c Config) IsDev() bool { return c.AppEnv == "development" }

func (c Config) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&loc=Local&charset=utf8mb4",
		c.DBUser, c.DBPass, c.DBHost, c.DBPort, c.DBName)
}

func Load() Config {
	loadDotEnv(".env")
	cookieSecure, err := strconv.ParseBool(get("COOKIE_SECURE", "false"))
	if err != nil {
		log.Printf("WARNING: COOKIE_SECURE tidak valid, Secure cookie diaktifkan untuk keamanan")
		cookieSecure = true
	}
	cfg := Config{
		CookieSecure:     cookieSecure,
		AppEnv:           get("APP_ENV", "development"),
		Port:             get("APP_PORT", "8080"),
		DBHost:           get("DB_HOST", "127.0.0.1"),
		DBPort:           get("DB_PORT", "3306"),
		DBUser:           get("DB_USER", "root"),
		DBPass:           get("DB_PASS", ""),
		DBName:           get("DB_NAME", "penggajian_mitrajaya"),
		DevAutoLoginRole: get("DEV_AUTO_LOGIN_ROLE", ""),
	}
	if !cfg.IsDev() {
		cfg.DevAutoLoginRole = "" // tidak pernah aktif di luar development
	}
	return cfg
}

func get(key, def string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return def
}

// loadDotEnv memuat pasangan KEY=VALUE; variabel yang sudah ada di environment tidak ditimpa.
func loadDotEnv(path string) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		k, v, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		k = strings.TrimSpace(k)
		v = strings.Trim(strings.TrimSpace(v), `"'`)
		if _, exists := os.LookupEnv(k); !exists {
			os.Setenv(k, v)
		}
	}
}
