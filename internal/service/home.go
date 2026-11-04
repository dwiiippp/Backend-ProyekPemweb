package service

import "penggajian/internal/repository"

// Pemilik: Anggota 1 | UC-02 · F-13 Dashboard & profil
// HomeService menampung logika bisnis modul ini. Dashboard per role dan halaman profil pekerja (NIK, nama, kategori kerja).
type HomeService struct {
	repo *repository.HomeRepository
}

func NewHomeService(repo *repository.HomeRepository) *HomeService {
	return &HomeService{repo: repo}
}
