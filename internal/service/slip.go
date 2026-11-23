package service

import (
	"context"
	"database/sql"
	"errors"

	"penggajian/internal/model"
	"penggajian/internal/repository"
)

// Pemilik: Anggota 3 | F-11 · UC-06 Lihat Slip · UC-15 Cetak Slip (BR-12)
// SlipService menampung logika bisnis modul ini. Pekerja hanya boleh melihat slip dengan pekerja_id miliknya.
type SlipService struct {
	repo *repository.SlipRepository
}

func NewSlipService(repo *repository.SlipRepository) *SlipService {
	return &SlipService{repo: repo}
}

type FilterOpsiSlip struct {
	Periode []repository.OpsiPeriodeSlip
	Pekerja []repository.OpsiPekerjaSlip
}

func (s *SlipService) FilterOptions(ctx context.Context) (FilterOpsiSlip, error) {
	periode, err := s.repo.DaftarPeriodeDitutup(ctx)
	if err != nil {
		return FilterOpsiSlip{}, err
	}
	pekerja, err := s.repo.DaftarPekerja(ctx)
	if err != nil {
		return FilterOpsiSlip{}, err
	}
	return FilterOpsiSlip{Periode: periode, Pekerja: pekerja}, nil
}

func (s *SlipService) DaftarAdmin(ctx context.Context, periodeID, pekerjaID int64) ([]repository.RingkasanSlip, error) {
	if periodeID < 0 || pekerjaID < 0 {
		return nil, ErrInputTidakValid
	}
	return s.repo.DaftarAdmin(ctx, periodeID, pekerjaID)
}

func (s *SlipService) DaftarMilikPekerja(ctx context.Context, pekerjaID int64) ([]repository.RingkasanSlip, error) {
	if pekerjaID <= 0 {
		return nil, ErrBukanMilikSendiri
	}
	return s.repo.DaftarMilikPekerja(ctx, pekerjaID)
}

func (s *SlipService) DetailAdmin(ctx context.Context, gajiID int64) (repository.DetailSlip, error) {
	if gajiID <= 0 {
		return repository.DetailSlip{}, ErrInputTidakValid
	}

	slip, err := s.repo.DetailByGajiID(ctx, gajiID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return repository.DetailSlip{}, ErrTidakDitemukan
		}
		return repository.DetailSlip{}, err
	}
	if slip.StatusPeriode != model.PeriodeDitutup {
		return repository.DetailSlip{}, ErrTidakDitemukan
	}
	return slip, nil
}

// CekKepemilikanSlip memastikan pekerja hanya mengakses slip gajinya sendiri (BR-12 / UC-06).
func CekKepemilikanSlip(slipPekerjaID, userPekerjaID int64) error {
	if userPekerjaID <= 0 || slipPekerjaID != userPekerjaID {
		return ErrBukanMilikSendiri
	}
	return nil
}

func (s *SlipService) DetailPekerja(ctx context.Context, gajiID, pekerjaID int64) (repository.DetailSlip, error) {
	if pekerjaID <= 0 {
		return repository.DetailSlip{}, ErrBukanMilikSendiri
	}
	if gajiID <= 0 {
		return repository.DetailSlip{}, ErrInputTidakValid
	}

	slip, err := s.repo.DetailByGajiID(ctx, gajiID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return repository.DetailSlip{}, ErrTidakDitemukan
		}
		return repository.DetailSlip{}, err
	}
	if err := CekKepemilikanSlip(slip.PekerjaID, pekerjaID); err != nil {
		return repository.DetailSlip{}, err
	}
	if slip.StatusPeriode != model.PeriodeDitutup {
		return repository.DetailSlip{}, ErrTidakDitemukan
	}
	return slip, nil
}
