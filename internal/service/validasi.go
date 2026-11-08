package service

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"penggajian/internal/repository"
)

type ValidasiService struct {
	repo *repository.ValidasiRepository
}

func NewValidasiService(repo *repository.ValidasiRepository) *ValidasiService {
	return &ValidasiService{repo: repo}
}

func (s *ValidasiService) DaftarHari(ctx context.Context) ([]repository.HariValidasi, error) {
	return s.repo.DaftarHari(ctx)
}

func (s *ValidasiService) Detail(ctx context.Context, tanggalText string) (time.Time, []repository.PresensiValidasi, bool, error) {
	tanggal, err := parseTanggal(tanggalText)
	if err != nil {
		return time.Time{}, nil, false, ErrInputTidakValid
	}
	items, err := s.repo.Detail(ctx, tanggal)
	if err != nil {
		return time.Time{}, nil, false, err
	}
	hasSchedule, closed, err := s.repo.InfoTanggal(ctx, tanggal)
	if err != nil {
		return time.Time{}, nil, false, err
	}
	if !hasSchedule {
		return time.Time{}, nil, false, ErrTidakDitemukan
	}
	return tanggal, items, closed, nil
}

func (s *ValidasiService) Simpan(ctx context.Context, tanggalText string, validatorID int64, input map[int64]InputKeputusanValidasi) error {
	if validatorID <= 0 {
		return ErrInputTidakValid
	}
	tanggal, err := parseTanggal(tanggalText)
	if err != nil {
		return ErrInputTidakValid
	}
	items, err := s.repo.Detail(ctx, tanggal)
	if err != nil {
		return err
	}
	hasSchedule, closed, err := s.repo.InfoTanggal(ctx, tanggal)
	if err != nil {
		return err
	}
	if !hasSchedule {
		return ErrTidakDitemukan
	}
	if closed {
		return ErrHariDitutup
	}
	if len(items) != len(input) {
		return ErrInputTidakValid
	}

	decisions := make([]repository.KeputusanValidasi, 0, len(items))
	for _, item := range items {
		if item.Status != "menunggu" {
			return ErrHariDitutup
		}
		choice, ok := input[item.ID]
		if !ok {
			return ErrInputTidakValid
		}
		choice.Catatan = strings.TrimSpace(choice.Catatan)
		decision := repository.KeputusanValidasi{AbsensiID: item.ID, Status: choice.Status}
		switch choice.Status {
		case "disetujui":
			late, overtime, err := hitungPresensi(item.JamMasukAktual, item.JamKeluarAktual, item.JamMasukShift, item.JamKeluarShift)
			if err != nil {
				return ErrInputTidakValid
			}
			decision.MenitTerlambat = sql.NullInt64{Int64: int64(late), Valid: true}
			decision.JamLembur = sql.NullString{String: overtime.StringFixed(2), Valid: true}
		case "ditolak":
			if choice.Catatan == "" {
				return ErrInputTidakValid
			}
			decision.Catatan = choice.Catatan
		default:
			return ErrInputTidakValid
		}
		decisions = append(decisions, decision)
	}

	err = s.repo.Simpan(ctx, tanggal, validatorID, decisions)
	switch {
	case errors.Is(err, repository.ErrValidasiSudahDitutup):
		return ErrHariDitutup
	case errors.Is(err, repository.ErrValidasiPeriodeDitutup):
		return ErrPeriodeDitutup
	case errors.Is(err, repository.ErrDataValidasiBerubah):
		return ErrInputTidakValid
	case errors.Is(err, repository.ErrTanggalValidasiTidakDitemukan):
		return ErrTidakDitemukan
	default:
		return err
	}
}

type InputKeputusanValidasi struct {
	Status  string
	Catatan string
}

func parseTanggal(value string) (time.Time, error) {
	return time.ParseInLocation("2006-01-02", strings.TrimSpace(value), time.Local)
}
