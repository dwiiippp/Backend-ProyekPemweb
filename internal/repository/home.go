package repository

import "database/sql"

// Pemilik: Anggota 1 | UC-02 · F-13 Dashboard & profil
// Tabel: pekerja (baca)
// Semua query SQL modul ini ditulis di sini (pakai placeholder ?, jangan concat string).
type HomeRepository struct {
	db *sql.DB
}

func NewHomeRepository(db *sql.DB) *HomeRepository {
	return &HomeRepository{db: db}
}
