package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"penggajian/internal/config"
	"penggajian/internal/service"
)

type SeedUser struct {
	Nama     string
	Email    string
	Role     string
	Password string
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	cfg := config.Load()

	// Seeder hanya boleh dijalankan di development.
	if !cfg.IsDev() {
		return fmt.Errorf("seeder hanya boleh dijalankan pada APP_ENV=development")
	}

	if os.Getenv("SEED_USERS_ENABLED") != "true" {
		return fmt.Errorf("seeder dinonaktifkan: atur SEED_USERS_ENABLED=true di .env")
	}

	users := []SeedUser{
		{"Admin Mitra Jaya", "admin@mitrajaya.test", "admin", os.Getenv("SEED_ADMIN_PASSWORD")},
		{"Budi", "budi@mitrajaya.test", "pekerja", os.Getenv("SEED_BUDI_PASSWORD")},
		{"Siti", "siti@mitrajaya.test", "pekerja", os.Getenv("SEED_SITI_PASSWORD")},
		{"Andi", "andi@mitrajaya.test", "pekerja", os.Getenv("SEED_ANDI_PASSWORD")},
		{"Dewi", "dewi@mitrajaya.test", "pekerja", os.Getenv("SEED_DEWI_PASSWORD")},
	}

	// Validasi password sebelum diproses.
	for _, user := range users {
		if strings.TrimSpace(user.Password) == "" {
			return fmt.Errorf("password untuk akun %s belum diisi", user.Email)
		}

		if len(user.Password) < 8 {
			return fmt.Errorf("password akun %s minimal 8 karakter", user.Email)
		}
	}

	// Ubah plaintext menjadi bcrypt hash.
	for i := range users {
		hash, err := service.HashPassword(users[i].Password)
		if err != nil {
			return fmt.Errorf("gagal melakukan hashing untuk %s: %w", users[i].Email, err)
		}

		users[i].Password = hash
	}

	// Koneksi database.
	db, err := config.OpenDB(cfg)
	if err != nil {
		return fmt.Errorf("gagal koneksi database: %w", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Gunakan transaksi agar perubahan akun konsisten.
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("gagal memulai transaksi: %w", err)
	}
	defer tx.Rollback()

	query := `
		INSERT INTO users (nama, email, password_hash, role, status)
		VALUES (?, ?, ?, ?, 'aktif')
		ON DUPLICATE KEY UPDATE
			nama = VALUES(nama),
			password_hash = VALUES(password_hash),
			role = VALUES(role),
			status = 'aktif'
	`

	for _, user := range users {
		_, err := tx.ExecContext(
			ctx,
			query,
			user.Nama,
			user.Email,
			user.Password,
			user.Role,
		)
		if err != nil {
			return fmt.Errorf("gagal menyimpan akun %s: %w", user.Email, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("gagal menyimpan transaksi: %w", err)
	}

	fmt.Println("Seeder akun berhasil dijalankan.")
	fmt.Println("Total akun diproses:", len(users))

	return nil
}
