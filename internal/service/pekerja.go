package service

import "penggajian/internal/repository"

// Pemilik: Anggota 1 | F-02 · UC-08 Kelola Pekerja
// PekerjaService menampung logika bisnis modul ini. CRUD pekerja dan akun login. Route lain: GET /admin/pekerja/baru, POST /admin/pekerja, GET /admin/pekerja/{id}, dst.
type PekerjaService struct {
	repo *repository.PekerjaRepository
}

func NewPekerjaService(repo *repository.PekerjaRepository) *PekerjaService {
	return &PekerjaService{repo: repo}
}
