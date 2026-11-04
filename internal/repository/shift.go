package repository

import "database/sql"

// Pemilik: Anggota 1 | F-04 Kelola Shift (BR-05)
// Tabel: shift (tulis)
// Semua query SQL modul ini ditulis di sini (pakai placeholder ?, jangan concat string).
type ShiftRepository struct {
	db *sql.DB
}

func NewShiftRepository(db *sql.DB) *ShiftRepository {
	return &ShiftRepository{db: db}
}
