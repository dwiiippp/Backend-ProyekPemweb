package repository

import "database/sql"

// Pemilik: Anggota 2 | F-12 Laporan presensi & rekap kasbon
// Tabel: absensi, kasbon (baca)
// Semua query SQL modul ini ditulis di sini (pakai placeholder ?, jangan concat string).
type LaporanPresensiKasbonRepository struct {
	db *sql.DB
}

func NewLaporanPresensiKasbonRepository(db *sql.DB) *LaporanPresensiKasbonRepository {
	return &LaporanPresensiKasbonRepository{db: db}
}
