package service

import "penggajian/internal/repository"

// Pemilik: Anggota 1 | F-04 Kelola Shift (BR-05)
// ShiftService menampung logika bisnis modul ini. CRUD shift. Seed: Pagi 07-15, Siang 15-23.
type ShiftService struct {
	repo *repository.ShiftRepository
}

func NewShiftService(repo *repository.ShiftRepository) *ShiftService {
	return &ShiftService{repo: repo}
}
