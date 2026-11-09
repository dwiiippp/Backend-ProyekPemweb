package service

import (
	"context"
	"errors"
	"time"

	"penggajian/internal/repository"
)

// JadwalService menampung logika bisnis modul ini. Atur jadwal (UNIQUE pekerja_id+tanggal). Pekerja hanya melihat jadwal miliknya (PekerjaID dari session).
type JadwalService struct {
	repo *repository.JadwalRepository
}

func NewJadwalService(repo *repository.JadwalRepository) *JadwalService {
	return &JadwalService{repo: repo}
}

func (s *JadwalService) GetPekerjaAktif(ctx context.Context) ([]repository.PekerjaOption, error) {
	return s.repo.GetPekerjaAktif(ctx)
}

func (s *JadwalService) GetShiftAktif(ctx context.Context) ([]repository.ShiftOption, error) {
	return s.repo.GetShiftAktif(ctx)
}

func (s *JadwalService) GetKalender(ctx context.Context, tanggalMulai time.Time, tanggalSelesai time.Time) ([]repository.JadwalRecord, error) {
	if tanggalMulai.IsZero() || tanggalSelesai.IsZero() || tanggalMulai.After(tanggalSelesai) {
		return nil, ErrInputTidakValid
	}

	return s.repo.GetKalender(ctx, tanggalMulai, tanggalSelesai)
}

func (s *JadwalService) GetByID(ctx context.Context, id int64) (repository.JadwalRecord, error) {
	if id <= 0 {
		return repository.JadwalRecord{}, ErrInputTidakValid
	}

	jadwal, err := s.repo.GetByID(ctx, id)
	if errors.Is(err, repository.ErrJadwalNotFound) {
		return repository.JadwalRecord{}, ErrTidakDitemukan
	}

	return jadwal, err
}

func (s *JadwalService) GetJadwalPekerja(ctx context.Context, pekerjaID int64) ([]repository.JadwalRecord, error) {
	if pekerjaID <= 0 {
		return nil, ErrInputTidakValid
	}

	return s.repo.GetJadwalPekerja(ctx, pekerjaID)
}

func (s *JadwalService) Create(ctx context.Context, input repository.CreateJadwalParams) error {
	if err := s.validateJadwalInput(ctx, input); err != nil {
		return err
	}

	return s.repo.Create(ctx, input)
}

func (s *JadwalService) Update(ctx context.Context, id int64, input repository.CreateJadwalParams) error {
	if id <= 0 {
		return ErrInputTidakValid
	}

	if err := s.validateJadwalInput(ctx, input); err != nil {
		return err
	}

	err := s.repo.Update(ctx, id, input)
	if errors.Is(err, repository.ErrJadwalNotFound) {
		return ErrTidakDitemukan
	}

	return err
}

func (s *JadwalService) validateJadwalInput(ctx context.Context, input repository.CreateJadwalParams) error {
	if input.PekerjaID <= 0 || input.ShiftID <= 0 || input.TanggalJadwal.IsZero() {
		return ErrInputTidakValid
	}

	pekerja, err := s.repo.GetPekerjaAktif(ctx)
	if err != nil {
		return err
	}

	pekerjaValid := false
	for _, item := range pekerja {
		if item.ID == input.PekerjaID {
			pekerjaValid = true
			break
		}
	}

	if !pekerjaValid {
		return ErrInputTidakValid
	}

	shifts, err := s.repo.GetShiftAktif(ctx)
	if err != nil {
		return err
	}

	shiftValid := false
	for _, item := range shifts {
		if item.ID == input.ShiftID {
			shiftValid = true
			break
		}
	}

	if !shiftValid {
		return ErrInputTidakValid
	}

	return nil
}
