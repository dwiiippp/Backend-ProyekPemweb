package repository

import "database/sql"

// Pemilik: Anggota 2 | F-08 · UC-11 Kelola Kasbon · UC-07 Riwayat Kasbon
// Tabel: kasbon (tulis; sisa_kasbon juga diubah modul penggajian saat tutup periode)
// Semua query SQL modul ini ditulis di sini (pakai placeholder ?, jangan concat string).
type KasbonRepository struct {
	db *sql.DB
}

func NewKasbonRepository(db *sql.DB) *KasbonRepository {
	return &KasbonRepository{db: db}
}
