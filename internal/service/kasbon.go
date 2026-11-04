package service

import "penggajian/internal/repository"

// Pemilik: Anggota 2 | F-08 · UC-11 Kelola Kasbon · UC-07 Riwayat Kasbon
// KasbonService menampung logika bisnis modul ini. Saat dibuat: sisa_kasbon = nominal_kasbon, status 'aktif'.
type KasbonService struct {
	repo *repository.KasbonRepository
}

func NewKasbonService(repo *repository.KasbonRepository) *KasbonService {
	return &KasbonService{repo: repo}
}
