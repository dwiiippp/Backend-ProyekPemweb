package repository

import "database/sql"

// Pemilik: Anggota 2 | F-14 · UC-10 Validasi Presensi Harian (BR-14)
// Tabel: absensi (ubah), validasi_presensi_harian (tulis)
// Semua query SQL modul ini ditulis di sini (pakai placeholder ?, jangan concat string).
type ValidasiRepository struct {
	db *sql.DB
}

func NewValidasiRepository(db *sql.DB) *ValidasiRepository {
	return &ValidasiRepository{db: db}
}
