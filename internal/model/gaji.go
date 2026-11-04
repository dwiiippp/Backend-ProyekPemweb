package model

import (
	"time"

	"github.com/shopspring/decimal"
)

const (
	PeriodeTerbuka  = "terbuka"
	PeriodeDitutup  = "ditutup"
	PeriodeMingguan = "mingguan"
	PeriodeBulanan  = "bulanan"

	KoreksiTambah = "tambah"
	KoreksiKurang = "kurang"
)

// PeriodeGaji memetakan tabel `periode_gaji`.
type PeriodeGaji struct {
	ID                int64
	JenisPeriode      string // mingguan | bulanan
	TanggalMulai      time.Time
	TanggalSelesai    time.Time
	TanggalPembayaran *time.Time
	Status            string // terbuka | ditutup
	DibuatPada        time.Time
}

// Gaji memetakan tabel `gaji` (UNIQUE periode_gaji_id + pekerja_id). Read-only setelah periode ditutup.
type Gaji struct {
	ID                      int64
	PeriodeGajiID           int64
	PekerjaID               int64
	GajiPokok               decimal.Decimal
	JumlahHariHadir         int
	PotonganTerlambat       decimal.Decimal
	UpahLembur              decimal.Decimal
	PenyesuaianProporsional decimal.Decimal
	GajiKotor               decimal.Decimal
	PotonganKasbon          decimal.Decimal
	GajiBersih              decimal.Decimal
	DibuatPada              time.Time
}

// PotonganKasbonGaji memetakan tabel `potongan_kasbon_gaji` (rincian potongan per kasbon).
type PotonganKasbonGaji struct {
	ID              int64
	GajiID          int64
	KasbonID        int64
	NominalPotongan decimal.Decimal
	DibuatPada      time.Time
}

// KoreksiGaji memetakan tabel `koreksi_gaji`.
type KoreksiGaji struct {
	ID                  int64
	PeriodeGajiTujuanID int64
	GajiSumberID        int64
	PekerjaID           int64
	JenisKoreksi        string // tambah | kurang
	Nominal             decimal.Decimal
	Alasan              *string
	DibuatPada          time.Time
}
