package service

import "penggajian/internal/repository"

// Pemilik: Anggota 3 | F-11 · UC-06 Lihat Slip · UC-15 Cetak Slip (BR-12)
// SlipService menampung logika bisnis modul ini. Pekerja hanya boleh melihat slip dengan pekerja_id miliknya.
type SlipService struct {
	repo *repository.SlipRepository
}

func NewSlipService(repo *repository.SlipRepository) *SlipService {
	return &SlipService{repo: repo}
}
