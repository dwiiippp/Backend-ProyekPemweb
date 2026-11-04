package repository

import "database/sql"

// Pemilik: Anggota 1 | F-05 Jadwal Kerja · UC-03
// Tabel: jadwal_kerja (tulis)
// Semua query SQL modul ini ditulis di sini (pakai placeholder ?, jangan concat string).
type JadwalRepository struct {
	db *sql.DB
}

func NewJadwalRepository(db *sql.DB) *JadwalRepository {
	return &JadwalRepository{db: db}
}
