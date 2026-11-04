package model

import (
	"time"

	"github.com/shopspring/decimal"
)

const (
	KasbonAktif = "aktif"
	KasbonLunas = "lunas"
)

// Kasbon memetakan tabel `kasbon`.
type Kasbon struct {
	ID               int64
	PekerjaID        int64
	TanggalPengajuan time.Time
	NominalKasbon    decimal.Decimal
	SisaKasbon       decimal.Decimal
	Keterangan       *string
	Status           string // aktif | lunas
	DibuatPada       time.Time
}
