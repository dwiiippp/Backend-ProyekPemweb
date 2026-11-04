package service

import "penggajian/internal/repository"

// Pemilik: Anggota 3 | F-07 · F-09 · UC-12 Proses Penggajian · UC-13 Tutup Periode
//
// SARAN AGAR TIDAK BERGANTUNG PADA ANGGOTA LAIN: tulis perhitungan sebagai fungsi MURNI
// (input struct -> output struct, tanpa akses DB) lalu uji dengan table-driven test
// di penggajian_test.go memakai angka dari db/seed.sql. Setelah itu baru sambungkan ke repository.
//
// Rumus (F-07), hanya dari absensi berstatus 'disetujui':
//
//	upah per jam      = upah harian / 8 (harian)  |  gaji bulanan / 173 (bulanan)
//	gaji pokok        = jumlah hari hadir x upah harian (harian)  |  gaji bulanan (bulanan)
//	penyesuaian prop. = (hari aktif / hari dalam bulan x gaji bulanan) - gaji bulanan   (negatif)
//	potongan terlambat= upah per jam, untuk tiap hari dengan menit_terlambat > 15
//	upah lembur       = (min(jam lembur,1) x 1,5 + max(jam lembur-1,0) x 2) x upah per jam, per hari
//	gaji kotor        = gaji pokok - potongan terlambat + upah lembur + penyesuaian +/- koreksi
//	potongan kasbon   = min(total sisa kasbon aktif, 30% x gaji kotor), kasbon tertua dulu
//	gaji bersih       = gaji kotor - potongan kasbon
//
// Gunakan decimal.Decimal (bukan float64) untuk semua nilai uang.
type PenggajianService struct {
	repo *repository.PenggajianRepository
}

func NewPenggajianService(repo *repository.PenggajianRepository) *PenggajianService {
	return &PenggajianService{repo: repo}
}
