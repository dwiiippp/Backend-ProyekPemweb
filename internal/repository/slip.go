package repository

import "database/sql"

// Pemilik: Anggota 3 | F-11 · UC-06 Lihat Slip · UC-15 Cetak Slip (BR-12)
// Tabel: gaji, potongan_kasbon_gaji (baca)
// Semua query SQL modul ini ditulis di sini (pakai placeholder ?, jangan concat string).
type SlipRepository struct {
	db *sql.DB
}

func NewSlipRepository(db *sql.DB) *SlipRepository {
	return &SlipRepository{db: db}
}
