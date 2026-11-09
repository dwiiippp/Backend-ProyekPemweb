package service

import (
	"context"
	"errors"
	"strings"
	"time"
	"unicode/utf8"

	"penggajian/internal/model"
	"penggajian/internal/repository"
)

var ErrShiftTidakValid = errors.New("data shift tidak valid")

type ShiftInput struct {
	NamaShift string
	JamMasuk  string
	JamKeluar string
}

// ShiftService menampung logika bisnis modul ini. CRUD shift. Seed: Pagi 07-15, Siang 15-23.
type ShiftService struct {
	repo *repository.ShiftRepository
}

func NewShiftService(repo *repository.ShiftRepository) *ShiftService {
	return &ShiftService{repo: repo}
}

func (s *ShiftService) List(ctx context.Context) ([]model.Shift, error) {
	return s.repo.GetAll(ctx)
}

func (s *ShiftService) GetByID(ctx context.Context, id int64) (model.Shift, error) {
	if id <= 0 {
		return model.Shift{}, ErrShiftTidakValid
	}
	return s.repo.GetByID(ctx, id)
}

func (s *ShiftService) Create(ctx context.Context, input ShiftInput) error {
	params, err := validateShift(input)
	if err != nil {
		return err
	}

	return s.repo.Create(ctx, params)
}

func (s *ShiftService) Update(ctx context.Context, id int64, input ShiftInput) error {
	if id <= 0 {
		return ErrShiftTidakValid
	}

	params, err := validateShift(input)
	if err != nil {
		return err
	}

	return s.repo.Update(ctx, id, params)
}

func (s *ShiftService) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return ErrShiftTidakValid
	}

	return s.repo.Delete(ctx, id)
}

func validateShift(input ShiftInput) (repository.CreateShiftParams, error) {
	nama := strings.TrimSpace(input.NamaShift)
	jamMasuk := strings.TrimSpace(input.JamMasuk)
	jamKeluar := strings.TrimSpace(input.JamKeluar)

	if nama == "" || utf8.RuneCountInString(nama) > 50 {
		return repository.CreateShiftParams{}, ErrShiftTidakValid
	}

	masuk, err := parseJam(jamMasuk)
	if err != nil {
		return repository.CreateShiftParams{}, ErrShiftTidakValid
	}

	keluar, err := parseJam(jamKeluar)
	if err != nil {
		return repository.CreateShiftParams{}, ErrShiftTidakValid
	}

	if masuk.Equal(keluar) {
		return repository.CreateShiftParams{}, ErrShiftTidakValid
	}

	return repository.CreateShiftParams{
		NamaShift: nama,
		JamMasuk:  masuk.Format("15:04:05"),
		JamKeluar: keluar.Format("15:04:05"),
	}, nil
}

func parseJam(value string) (time.Time, error) {
	layouts := []string{"15:04", "15:04:05"}

	for _, layout := range layouts {
		parsed, err := time.Parse(layout, value)
		if err == nil {
			return parsed, nil
		}
	}

	return time.Time{}, ErrShiftTidakValid
}
