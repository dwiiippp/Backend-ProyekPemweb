package service

import "penggajian/internal/repository"

// Pemilik: Anggota 3 | F-12 Laporan status periode & rekap biaya gaji/lembur
// LaporanGajiService menampung logika bisnis modul ini. Hanya baca data.
type LaporanGajiService struct {
	repo *repository.LaporanGajiRepository
}

func NewLaporanGajiService(repo *repository.LaporanGajiRepository) *LaporanGajiService {
	return &LaporanGajiService{repo: repo}
}
