package repository

import "database/sql"

// Pemilik: Anggota 1 | F-02 · UC-08 Kelola Pekerja
// Tabel: pekerja, users (tulis)
// Semua query SQL modul ini ditulis di sini (pakai placeholder ?, jangan concat string).
type PekerjaRepository struct {
	db *sql.DB
}

func NewPekerjaRepository(db *sql.DB) *PekerjaRepository {
	return &PekerjaRepository{db: db}
}
