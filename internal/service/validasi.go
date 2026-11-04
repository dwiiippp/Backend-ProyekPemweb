package service

import "penggajian/internal/repository"

// Pemilik: Anggota 2 | F-14 · UC-10 Validasi Presensi Harian (BR-14)
// ValidasiService menampung logika bisnis modul ini. Saat validasi: hitung menit_terlambat & jam_lembur untuk yang disetujui, lalu INSERT validasi_presensi_harian — semuanya dalam satu transaksi.
type ValidasiService struct {
	repo *repository.ValidasiRepository
}

func NewValidasiService(repo *repository.ValidasiRepository) *ValidasiService {
	return &ValidasiService{repo: repo}
}
