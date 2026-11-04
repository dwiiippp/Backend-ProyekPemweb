package repository

import "database/sql"

// Pemilik: Anggota 3 | Periode Gaji (BR-01, BR-02, BR-10)
// Tabel: periode_gaji (tulis)
// Semua query SQL modul ini ditulis di sini (pakai placeholder ?, jangan concat string).
type PeriodeRepository struct {
	db *sql.DB
}

func NewPeriodeRepository(db *sql.DB) *PeriodeRepository {
	return &PeriodeRepository{db: db}
}
