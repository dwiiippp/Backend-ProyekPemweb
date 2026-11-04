package repository

import "database/sql"

// Pemilik: Anggota 2 | F-06 · UC-04 Input Presensi · UC-05 Riwayat (BR-13, BR-15)
// Tabel: absensi (tulis); baca jadwal_kerja, shift, validasi_presensi_harian, periode_gaji
// Semua query SQL modul ini ditulis di sini (pakai placeholder ?, jangan concat string).
type AbsensiRepository struct {
	db *sql.DB
}

func NewAbsensiRepository(db *sql.DB) *AbsensiRepository {
	return &AbsensiRepository{db: db}
}
