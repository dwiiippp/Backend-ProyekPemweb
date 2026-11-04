package service

import "penggajian/internal/repository"

// Pemilik: Anggota 2 | F-06 · UC-04 Input Presensi · UC-05 Riwayat (BR-13, BR-15)
// AbsensiService menampung logika bisnis modul ini. Tolak bila: hari sudah ditutup (ErrHariDitutup), periode ditutup, presensi ganda, jam keluar < jam keluar shift, atau jadwal bukan milik sendiri. Status awal 'menunggu'.
type AbsensiService struct {
	repo *repository.AbsensiRepository
}

func NewAbsensiService(repo *repository.AbsensiRepository) *AbsensiService {
	return &AbsensiService{repo: repo}
}
