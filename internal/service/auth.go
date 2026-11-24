package service

import (
	"context"
	"strings"

	"penggajian/internal/model"
	"penggajian/internal/repository"
)

// Pemilik: Anggota 1 | F-01 · UC-01 Login
type AuthService struct {
	repo *repository.AuthRepository
}

func NewAuthService(repo *repository.AuthRepository) *AuthService {
	return &AuthService{repo: repo}
}

// Login memvalidasi email dan password.
// Email tidak ditemukan dan password salah memberikan pesan yang sama.
func (s *AuthService) Login(ctx context.Context, email, password string) (*model.SessionUser, error) {
	email = strings.TrimSpace(strings.ToLower(email))

	if email == "" || strings.TrimSpace(password) == "" {
		return nil, ErrInputTidakValid
	}

	rec, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	if rec == nil || !CheckPassword(rec.User.PasswordHash, password) {
		return nil, ErrKredensialSalah
	}

	if rec.User.Status != "aktif" {
		return nil, ErrAkunNonaktif
	}

	return &model.SessionUser{
		UserID:    rec.User.ID,
		Username:  rec.User.Username,
		Nama:      rec.Nama,
		Role:      rec.User.Role,
		PekerjaID: rec.PekerjaID,
	}, nil
}
