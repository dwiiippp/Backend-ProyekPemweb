package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/shopspring/decimal"

	"penggajian/internal/repository"
)

type AbsensiService struct {
	repo *repository.AbsensiRepository
}

func NewAbsensiService(repo *repository.AbsensiRepository) *AbsensiService {
	return &AbsensiService{repo: repo}
}

func (s *AbsensiService) JadwalTerbuka(ctx context.Context, pekerjaID int64, today time.Time) ([]repository.JadwalPresensi, error) {
	if pekerjaID <= 0 {
		return nil, ErrBukanMilikSendiri
	}
	return s.repo.ListJadwalTerbuka(ctx, pekerjaID, today)
}

func (s *AbsensiService) Riwayat(ctx context.Context, pekerjaID int64) ([]repository.RiwayatPresensi, error) {
	if pekerjaID <= 0 {
		return nil, ErrBukanMilikSendiri
	}
	return s.repo.Riwayat(ctx, pekerjaID)
}

func (s *AbsensiService) Simpan(ctx context.Context, pekerjaID, jadwalID int64, jamMasuk, jamKeluar string) error {
	if pekerjaID <= 0 || jadwalID <= 0 {
		return ErrInputTidakValid
	}
	schedule, err := s.repo.Jadwal(ctx, pekerjaID, jadwalID)
	if errors.Is(err, repository.ErrJadwalTidakDitemukan) {
		return ErrBukanMilikSendiri
	}
	if err != nil {
		return err
	}

	startClock, err := parseClock(schedule.JamMasukShift)
	if err != nil {
		return err
	}
	endClock, err := parseClock(schedule.JamKeluarShift)
	if err != nil {
		return err
	}
	checkInClock, err := parseClock(jamMasuk)
	if err != nil {
		return ErrInputTidakValid
	}
	checkOutClock, err := parseClock(jamKeluar)
	if err != nil {
		return ErrInputTidakValid
	}

	date := time.Date(schedule.Tanggal.Year(), schedule.Tanggal.Month(), schedule.Tanggal.Day(), 0, 0, 0, 0, schedule.Tanggal.Location())
	today := time.Now()
	today = time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, today.Location())
	if date.After(today) {
		return ErrInputTidakValid
	}
	start := atClock(date, startClock)
	end := atClock(date, endClock)
	if !end.After(start) {
		end = end.AddDate(0, 0, 1)
	}

	checkIn := atClock(date, checkInClock)
	checkOut := atClock(date, checkOutClock)
	if checkOutClock.Before(startClock) {
		checkOut = checkOut.AddDate(0, 0, 1)
	}
	if !endClock.After(startClock) && checkInClock.Before(startClock) && checkInClock.Hour() < endClock.Hour() {
		checkIn = checkIn.AddDate(0, 0, 1)
	}
	if !checkOut.After(checkIn) || checkOut.Before(end) {
		return ErrInputTidakValid
	}

	if err := s.repo.Simpan(ctx, pekerjaID, jadwalID, checkIn, checkOut); err != nil {
		switch {
		case errors.Is(err, repository.ErrJadwalTidakDitemukan):
			return ErrBukanMilikSendiri
		case errors.Is(err, repository.ErrHariDitutup):
			return ErrHariDitutup
		case errors.Is(err, repository.ErrPeriodeDitutup):
			return ErrPeriodeDitutup
		case errors.Is(err, repository.ErrPresensiGanda):
			return ErrPresensiGanda
		default:
			return err
		}
	}
	return nil
}

func parseClock(value string) (time.Time, error) {
	value = strings.TrimSpace(value)
	for _, layout := range []string{"15:04", "15:04:05"} {
		if t, err := time.Parse(layout, value); err == nil {
			return t, nil
		}
	}
	return time.Time{}, ErrInputTidakValid
}

func atClock(date, clock time.Time) time.Time {
	return time.Date(date.Year(), date.Month(), date.Day(), clock.Hour(), clock.Minute(), clock.Second(), 0, date.Location())
}

func hitungPresensi(actualIn, actualOut time.Time, shiftIn, shiftOut string) (int, decimal.Decimal, error) {
	startClock, err := parseClock(shiftIn)
	if err != nil {
		return 0, decimal.Zero, err
	}
	endClock, err := parseClock(shiftOut)
	if err != nil {
		return 0, decimal.Zero, err
	}
	start := atClock(time.Date(actualIn.Year(), actualIn.Month(), actualIn.Day(), 0, 0, 0, 0, actualIn.Location()), startClock)
	end := atClock(start, endClock)
	if !end.After(start) {
		end = end.AddDate(0, 0, 1)
	}
	if actualOut.Before(end) {
		return 0, decimal.Zero, ErrInputTidakValid
	}
	late := 0
	if actualIn.After(start) {
		late = int(actualIn.Sub(start).Minutes())
	}
	overtimeMinutes := 0
	if actualOut.After(end) {
		overtimeMinutes = int(actualOut.Sub(end).Minutes())
	}
	overtime := decimal.NewFromInt(int64(overtimeMinutes)).Div(decimal.NewFromInt(60)).Round(2)
	return late, overtime, nil
}
