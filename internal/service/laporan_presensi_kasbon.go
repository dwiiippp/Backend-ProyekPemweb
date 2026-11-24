package service

import (
	"context"
	"time"

	"penggajian/internal/repository"
)

type LaporanPresensiKasbonService struct {
	repo *repository.LaporanPresensiKasbonRepository
}

func NewLaporanPresensiKasbonService(repo *repository.LaporanPresensiKasbonRepository) *LaporanPresensiKasbonService {
	return &LaporanPresensiKasbonService{repo: repo}
}

func (s *LaporanPresensiKasbonService) Pekerja(ctx context.Context) ([]repository.PekerjaFilter, error) {
	return s.repo.Pekerja(ctx)
}

func (s *LaporanPresensiKasbonService) Presensi(ctx context.Context, mulaiText, selesaiText string, pekerjaID int64) (time.Time, time.Time, []repository.BarisLaporanPresensi, error) {
	mulai, err := parseTanggal(mulaiText)
	if err != nil {
		return time.Time{}, time.Time{}, nil, ErrInputTidakValid
	}
	selesai, err := parseTanggal(selesaiText)
	if err != nil || mulai.After(selesai) || pekerjaID < 0 {
		return time.Time{}, time.Time{}, nil, ErrInputTidakValid
	}
	rows, err := s.repo.Presensi(ctx, mulai, selesai, pekerjaID)
	return mulai, selesai, rows, err
}

func (s *LaporanPresensiKasbonService) RekapKasbon(ctx context.Context) ([]repository.RekapKasbon, error) {
	return s.repo.RekapKasbon(ctx)
}
