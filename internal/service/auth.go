package service

import "penggajian/internal/repository"

// Pemilik: Anggota 1 | F-01 · UC-01 Login
// AuthService menampung logika bisnis modul ini. Validasi username+password (service.CheckPassword), tolak akun nonaktif, buat session via d.Sessions.Create, redirect sesuai role.
type AuthService struct {
	repo *repository.AuthRepository
}

func NewAuthService(repo *repository.AuthRepository) *AuthService {
	return &AuthService{repo: repo}
}
