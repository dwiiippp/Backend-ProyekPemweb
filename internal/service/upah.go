package service

import "penggajian/internal/repository"

// Pemilik: Anggota 1 | F-03 Riwayat Upah (BR-04)
// UpahService menampung logika bisnis modul ini. Tambah upah baru dengan tanggal berlaku; jangan UPDATE baris lama.
type UpahService struct {
	repo *repository.UpahRepository
}

func NewUpahService(repo *repository.UpahRepository) *UpahService {
	return &UpahService{repo: repo}
}
