package repository

import "database/sql"

// Pemilik: Anggota 3 | F-12 Laporan status periode & rekap biaya gaji/lembur
// Tabel: periode_gaji, gaji (baca)
// Semua query SQL modul ini ditulis di sini (pakai placeholder ?, jangan concat string).
type LaporanGajiRepository struct {
	db *sql.DB
}

func NewLaporanGajiRepository(db *sql.DB) *LaporanGajiRepository {
	return &LaporanGajiRepository{db: db}
}
