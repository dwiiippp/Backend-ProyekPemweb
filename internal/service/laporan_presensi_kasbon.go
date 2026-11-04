package service

import "penggajian/internal/repository"

// Pemilik: Anggota 2 | F-12 Laporan presensi & rekap kasbon
// LaporanPresensiKasbonService menampung logika bisnis modul ini. Hanya baca data.
type LaporanPresensiKasbonService struct {
	repo *repository.LaporanPresensiKasbonRepository
}

func NewLaporanPresensiKasbonService(repo *repository.LaporanPresensiKasbonRepository) *LaporanPresensiKasbonService {
	return &LaporanPresensiKasbonService{repo: repo}
}
