package service

import "penggajian/internal/repository"

// Pemilik: Anggota 3 | Periode Gaji (BR-01, BR-02, BR-10)
// PeriodeService menampung logika bisnis modul ini. Buat periode baru; periode ditutup tidak boleh diubah.
type PeriodeService struct {
	repo *repository.PeriodeRepository
}

func NewPeriodeService(repo *repository.PeriodeRepository) *PeriodeService {
	return &PeriodeService{repo: repo}
}
