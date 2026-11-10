package service

import (
	"context"
	"errors"
	"math"
	"penggajian/internal/repository"
	"time"
)

var (
	ErrUpahTidakValid      = errors.New("data upah tidak valid")
	ErrTanggalUpahDuplikat = errors.New("tanggal berlaku upah sudah digunakan")
)

type UpahInput struct {
	PekerjaID      int64
	NominalGaji    float64
	TanggalBerlaku time.Time
}

// UpahService menampung logika bisnis modul ini. Tambah upah baru dengan tanggal berlaku; jangan UPDATE baris lama.
type UpahService struct {
	repo *repository.UpahRepository
}

func NewUpahService(repo *repository.UpahRepository) *UpahService {
	return &UpahService{repo: repo}
}

func (s *UpahService) List(ctx context.Context, pekerjaID int64) ([]repository.UpahRecord, error) {
	if pekerjaID <= 0 {
		return nil, ErrUpahTidakValid
	}

	return s.repo.GetByPekerjaID(ctx, pekerjaID)
}

func (s *UpahService) Create(ctx context.Context, input UpahInput) error {
	if input.PekerjaID <= 0 {
		return ErrUpahTidakValid
	}

	if math.IsNaN(input.NominalGaji) || math.IsInf(input.NominalGaji, 0) || input.NominalGaji <= 0 || input.NominalGaji > 9999999999.99 {
		return ErrUpahTidakValid
	}

	if input.TanggalBerlaku.IsZero() {
		return ErrUpahTidakValid
	}

	err := s.repo.Create(ctx, repository.CreateUpahParams{
		PekerjaID:      input.PekerjaID,
		NominalGaji:    input.NominalGaji,
		TanggalBerlaku: input.TanggalBerlaku,
	})

	if errors.Is(err, repository.ErrDuplicateUpah) {
		return ErrTanggalUpahDuplikat
	}

	return err
}
