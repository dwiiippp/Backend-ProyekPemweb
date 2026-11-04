package model

import (
	"time"

	"github.com/shopspring/decimal"
)

const (
	ValidasiMenunggu  = "menunggu"
	ValidasiDisetujui = "disetujui"
	ValidasiDitolak   = "ditolak"
)

// Absensi memetakan tabel `absensi` (satu absensi per jadwal).
// MenitTerlambat dan JamLembur baru terisi saat Admin menyetujui (BR-14).
type Absensi struct {
	ID              int64
	PekerjaID       int64
	JadwalKerjaID   int64
	TanggalAbsensi  time.Time
	JamMasukAktual  *time.Time
	JamKeluarAktual *time.Time
	MenitTerlambat  *int
	JamLembur       decimal.NullDecimal
	StatusValidasi  string // menunggu | disetujui | ditolak
	CatatanValidasi *string
	DivalidasiOleh  *int64
	DivalidasiPada  *time.Time
	DibuatPada      time.Time
}

// ValidasiPresensiHarian memetakan tabel `validasi_presensi_harian`.
// Adanya baris untuk suatu tanggal berarti presensi pada tanggal itu DITUTUP (BR-15).
type ValidasiPresensiHarian struct {
	ID              int64
	TanggalPresensi time.Time
	DivalidasiOleh  int64
	DivalidasiPada  time.Time
	DibuatPada      time.Time
}
