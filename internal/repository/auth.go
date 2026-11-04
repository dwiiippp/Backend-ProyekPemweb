package repository

import "database/sql"

// Pemilik: Anggota 1 | F-01 · UC-01 Login
// Tabel: users (tulis/baca)
// Semua query SQL modul ini ditulis di sini (pakai placeholder ?, jangan concat string).
type AuthRepository struct {
	db *sql.DB
}

func NewAuthRepository(db *sql.DB) *AuthRepository {
	return &AuthRepository{db: db}
}
