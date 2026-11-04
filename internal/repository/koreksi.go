package repository

import "database/sql"

// Pemilik: Anggota 3 | F-10 · UC-14 Koreksi Gaji (BR-11)
// Tabel: koreksi_gaji (tulis)
// Semua query SQL modul ini ditulis di sini (pakai placeholder ?, jangan concat string).
type KoreksiRepository struct {
	db *sql.DB
}

func NewKoreksiRepository(db *sql.DB) *KoreksiRepository {
	return &KoreksiRepository{db: db}
}
