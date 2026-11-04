package repository

import "database/sql"

// Pemilik: Anggota 3 | F-07 · F-09 · UC-12 Proses Penggajian · UC-13 Tutup Periode
// Tabel: gaji, potongan_kasbon_gaji (tulis); kasbon, periode_gaji (ubah); baca absensi, riwayat_gaji, pekerja, koreksi_gaji
// Semua query SQL modul ini ditulis di sini (pakai placeholder ?, jangan concat string).
type PenggajianRepository struct {
	db *sql.DB
}

func NewPenggajianRepository(db *sql.DB) *PenggajianRepository {
	return &PenggajianRepository{db: db}
}
