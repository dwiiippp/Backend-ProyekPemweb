package service

import "penggajian/internal/repository"

// Pemilik: Anggota 1 | F-05 Jadwal Kerja · UC-03
// JadwalService menampung logika bisnis modul ini. Atur jadwal (UNIQUE pekerja_id+tanggal). Pekerja hanya melihat jadwal miliknya (PekerjaID dari session).
type JadwalService struct {
	repo *repository.JadwalRepository
}

func NewJadwalService(repo *repository.JadwalRepository) *JadwalService {
	return &JadwalService{repo: repo}
}
