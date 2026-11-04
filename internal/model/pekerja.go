package model

import (
	"time"

	"github.com/shopspring/decimal"
)

const (
	KategoriHarian  = "harian"
	KategoriBulanan = "bulanan"
)

// Pekerja memetakan tabel `pekerja`.
type Pekerja struct {
	ID            int64
	UserID        *int64
	NIK           string
	Nama          string
	KategoriKerja string // harian | bulanan
	TanggalMasuk  time.Time
	TanggalKeluar *time.Time
	Status        string // aktif | nonaktif
	DibuatPada    time.Time
	DiubahPada    time.Time
}

// RiwayatGaji memetakan tabel `riwayat_gaji` (upah harian atau bulanan sesuai kategori).
// Upah yang dipakai adalah baris dengan TanggalBerlaku terbesar <= tanggal terkait (BR-04).
type RiwayatGaji struct {
	ID             int64
	PekerjaID      int64
	NominalGaji    decimal.Decimal
	TanggalBerlaku time.Time
	DibuatPada     time.Time
}
