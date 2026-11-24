package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/shopspring/decimal"

	"penggajian/internal/repository"
)

type KasbonService struct {
	repo *repository.KasbonRepository
}

func NewKasbonService(repo *repository.KasbonRepository) *KasbonService {
	return &KasbonService{repo: repo}
}

func (s *KasbonService) Daftar(ctx context.Context) ([]repository.KasbonRecord, error) {
	return s.repo.Daftar(ctx)
}

func (s *KasbonService) MilikPekerja(ctx context.Context, pekerjaID int64) ([]repository.KasbonRecord, error) {
	if pekerjaID <= 0 {
		return nil, ErrBukanMilikSendiri
	}
	return s.repo.MilikPekerja(ctx, pekerjaID)
}

func (s *KasbonService) Buat(ctx context.Context, pekerjaID int64, nominalText, keterangan string, tanggal time.Time) error {
	if pekerjaID <= 0 {
		return ErrBukanMilikSendiri
	}
	nominal, err := parseNominalKasbon(nominalText)
	if err != nil {
		return ErrInputTidakValid
	}
	keterangan = strings.TrimSpace(keterangan)
	if len(keterangan) > 2000 {
		return ErrInputTidakValid
	}
	err = s.repo.Buat(ctx, pekerjaID, nominal, keterangan, tanggal)
	if errors.Is(err, repository.ErrPekerjaKasbonTidakDitemukan) {
		return ErrBukanMilikSendiri
	}
	return err
}

func parseNominalKasbon(value string) (decimal.Decimal, error) {
	value = strings.TrimSpace(strings.ReplaceAll(value, ",", "."))
	nominal, err := decimal.NewFromString(value)
	if err != nil || !nominal.IsPositive() || !nominal.Equal(nominal.Round(2)) {
		return decimal.Zero, ErrInputTidakValid
	}
	if nominal.GreaterThan(decimal.RequireFromString("9999999999.99")) {
		return decimal.Zero, ErrInputTidakValid
	}
	return nominal, nil
}
