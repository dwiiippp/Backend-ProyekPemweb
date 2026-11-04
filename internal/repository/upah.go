package repository

import "database/sql"

// Pemilik: Anggota 1 | F-03 Riwayat Upah (BR-04)
// Tabel: riwayat_gaji (tulis)
// Semua query SQL modul ini ditulis di sini (pakai placeholder ?, jangan concat string).
type UpahRepository struct {
	db *sql.DB
}

func NewUpahRepository(db *sql.DB) *UpahRepository {
	return &UpahRepository{db: db}
}
