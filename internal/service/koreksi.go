package service

import "penggajian/internal/repository"

// Pemilik: Anggota 3 | F-10 · UC-14 Koreksi Gaji (BR-11)
// KoreksiService menampung logika bisnis modul ini. Koreksi dihitung langsung dari koreksi_gaji per periode_gaji_tujuan_id agar tidak terhitung dua kali.
type KoreksiService struct {
	repo *repository.KoreksiRepository
}

func NewKoreksiService(repo *repository.KoreksiRepository) *KoreksiService {
	return &KoreksiService{repo: repo}
}
