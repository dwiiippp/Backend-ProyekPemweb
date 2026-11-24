package service

import (
	"context"
	"errors"

	"penggajian/internal/repository"
)

// HomeService menampung logika bisnis modul ini. Dashboard per role dan halaman profil pekerja (NIK, nama, kategori kerja).
type HomeService struct {
	repo *repository.HomeRepository
}

func NewHomeService(repo *repository.HomeRepository) *HomeService {
	return &HomeService{repo: repo}
}

// Mengambil ringkasan dashboard admin
func (s *HomeService) GetAdminSummary(ctx context.Context) (*repository.HomeAdminSummary, error) {
	return s.repo.GetAdminSummary(ctx)
}

// Mengambil jadwal pekerja hari ini
func (s *HomeService) GetTodaySchedule(ctx context.Context, pekerjaID int64) (*repository.HomeTodaySchedule, error) {
	if pekerjaID <= 0 {
		return nil, errors.New("ID pekerja tidak valid")
	}

	return s.repo.GetTodaySchedule(ctx, pekerjaID)
}

// Mengambil profil pekerja
func (s *HomeService) GetProfile(ctx context.Context, pekerjaID int64) (*repository.HomeProfile, error) {
	if pekerjaID <= 0 {
		return nil, errors.New("ID pekerja tidak valid")
	}

	return s.repo.GetProfile(ctx, pekerjaID)
}
