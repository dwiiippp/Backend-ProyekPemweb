package service

import (
	"context"
	"net/mail"
	"strings"
	"time"

	"penggajian/internal/repository"
)

// Pemilik: Anggota 1 | F-02 · UC-08 Kelola Pekerja
// PekerjaService menampung logika bisnis modul ini. CRUD pekerja dan akun login. Route lain: GET /admin/pekerja/baru, POST /admin/pekerja, GET /admin/pekerja/{id}, dst.
type PekerjaService struct {
	repo *repository.PekerjaRepository
}

func NewPekerjaService(repo *repository.PekerjaRepository) *PekerjaService {
	return &PekerjaService{repo: repo}
}

type PekerjaInput struct {
	NIK           string
	Nama          string
	Email         string
	Password      string
	KategoriKerja string
	TanggalMasuk  time.Time
}

func (s *PekerjaService) List(ctx context.Context) ([]repository.PekerjaRecord, error) {
	return s.repo.GetAll(ctx)
}

// GetByID mengambil detail pekerja berdasarkan ID
func (s *PekerjaService) GetByID(ctx context.Context, id int64) (*repository.PekerjaRecord, error) {
	if id <= 0 {
		return nil, ErrInputTidakValid
	}

	pekerja, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if pekerja == nil {
		return nil, ErrTidakDitemukan
	}

	return pekerja, nil
}

// Create akun login dan data pekerja
func (s *PekerjaService) Create(ctx context.Context, input PekerjaInput) (int64, error) {
	input = NormalizePekerjaInput(input)

	if err := validatePekerjaInput(input, true); err != nil {
		return 0, err
	}

	//Hash password
	passwordHash, err := HashPassword(input.Password)
	if err != nil {
		return 0, err
	}

	return s.repo.CreateWithAccount(ctx, repository.CreatePekerjaParams{
		NIK:           input.NIK,
		Nama:          input.Nama,
		Email:         input.Email,
		PasswordHash:  passwordHash,
		KategoriKerja: input.KategoriKerja,
		TanggalMasuk:  input.TanggalMasuk,
	})
}

func (s *PekerjaService) Update(ctx context.Context, id int64, input PekerjaInput) error {
	input = NormalizePekerjaInput(input)

	if id <= 0 {
		return ErrInputTidakValid
	}

	if err := validatePekerjaInput(input, false); err != nil {
		return err
	}

	//Pastikan pekerja dengan id tersebut ada
	_, err := s.GetByID(ctx, id)
	if err != nil {
		return err
	}

	return s.repo.Update(ctx, repository.UpdatePekerjaParams{
		ID:            id,
		NIK:           input.NIK,
		Nama:          input.Nama,
		Email:         input.Email,
		KategoriKerja: input.KategoriKerja,
		TanggalMasuk:  input.TanggalMasuk,
	})
}

// Deactivate menonaktifkan akun dan data pekerja.
func (s *PekerjaService) Deactivate(ctx context.Context, id int64) error {
	if id <= 0 {
		return ErrInputTidakValid
	}

	// Pastikan pekrja tersedia sebelum dinonaktifkan
	_, err := s.GetByID(ctx, id)
	if err != nil {
		return err
	}

	return s.repo.Deactivate(ctx, id)
}

// normalizePekerjaInput membersihkan spasi dan menyerangamkan email
func NormalizePekerjaInput(input PekerjaInput) PekerjaInput {
	input.NIK = strings.TrimSpace(input.NIK)
	input.Nama = strings.TrimSpace(input.Nama)
	input.Email = strings.ToLower(strings.TrimSpace(input.Email))
	input.KategoriKerja = strings.TrimSpace(input.KategoriKerja)

	return input
}

func validatePekerjaInput(input PekerjaInput, requirePassword bool) error {
	if input.NIK == "" ||
		input.Nama == "" ||
		input.Email == "" ||
		input.KategoriKerja == "" ||
		input.TanggalMasuk.IsZero() {
		return ErrInputTidakValid
	}

	if input.KategoriKerja != "harian" && input.KategoriKerja != "bulanan" {
		return ErrInputTidakValid
	}

	email, err := mail.ParseAddress(input.Email)
	if err != nil || email.Address != input.Email {
		return ErrInputTidakValid
	}

	if requirePassword && len(input.Password) < 8 {
		return ErrInputTidakValid
	}

	return nil
}
