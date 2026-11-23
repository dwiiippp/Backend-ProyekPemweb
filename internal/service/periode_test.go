package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"penggajian/internal/model"
)

func TestBuatPeriodeValidasiAturan(t *testing.T) {
	s := &PeriodeService{}
	tgl := func(y int, m time.Month, d int) time.Time {
		return time.Date(y, m, d, 0, 0, 0, 0, time.Local)
	}
	tglBayarAwal := tgl(2026, 10, 23)

	tests := []struct {
		nama    string
		periode model.PeriodeGaji
	}{
		{
			nama: "jenis tidak valid",
			periode: model.PeriodeGaji{
				JenisPeriode:   "harian",
				TanggalMulai:   tgl(2026, 11, 2),
				TanggalSelesai: tgl(2026, 11, 7),
			},
		},
		{
			nama: "tanggal selesai sebelum mulai",
			periode: model.PeriodeGaji{
				JenisPeriode:   model.PeriodeMingguan,
				TanggalMulai:   tgl(2026, 11, 7),
				TanggalSelesai: tgl(2026, 11, 2),
			},
		},
		{
			nama: "mingguan tidak dimulai hari Senin",
			periode: model.PeriodeGaji{
				JenisPeriode:   model.PeriodeMingguan,
				TanggalMulai:   tgl(2026, 10, 20), // Selasa
				TanggalSelesai: tgl(2026, 10, 24), // Sabtu
			},
		},
		{
			nama: "mingguan tidak berakhir hari Sabtu",
			periode: model.PeriodeGaji{
				JenisPeriode:   model.PeriodeMingguan,
				TanggalMulai:   tgl(2026, 10, 19), // Senin
				TanggalSelesai: tgl(2026, 10, 25), // Minggu
			},
		},
		{
			nama: "mingguan rentang dua minggu",
			periode: model.PeriodeGaji{
				JenisPeriode:   model.PeriodeMingguan,
				TanggalMulai:   tgl(2026, 10, 19), // Senin
				TanggalSelesai: tgl(2026, 10, 31), // Sabtu minggu depan
			},
		},
		{
			nama: "bulanan tidak dimulai tanggal 1",
			periode: model.PeriodeGaji{
				JenisPeriode:   model.PeriodeBulanan,
				TanggalMulai:   tgl(2026, 10, 2),
				TanggalSelesai: tgl(2026, 10, 31),
			},
		},
		{
			nama: "bulanan tidak berakhir akhir bulan",
			periode: model.PeriodeGaji{
				JenisPeriode:   model.PeriodeBulanan,
				TanggalMulai:   tgl(2026, 10, 1),
				TanggalSelesai: tgl(2026, 10, 30),
			},
		},
		{
			nama: "tanggal pembayaran sebelum tanggal selesai",
			periode: model.PeriodeGaji{
				JenisPeriode:      model.PeriodeMingguan,
				TanggalMulai:      tgl(2026, 10, 19),
				TanggalSelesai:    tgl(2026, 10, 24),
				TanggalPembayaran: &tglBayarAwal,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.nama, func(t *testing.T) {
			_, err := s.Buat(context.Background(), tt.periode)
			if !errors.Is(err, ErrInputTidakValid) {
				t.Fatalf("ingin ErrInputTidakValid, mendapat: %v", err)
			}
		})
	}
}

func TestAmbilByIDTidakValid(t *testing.T) {
	s := &PeriodeService{}

	_, err := s.AmbilByID(context.Background(), 0)
	if !errors.Is(err, ErrInputTidakValid) {
		t.Fatalf("ingin ErrInputTidakValid, mendapat: %v", err)
	}
}

func TestDetailPeriodeIDTidakValid(t *testing.T) {
	s := &PeriodeService{}

	_, err := s.Detail(context.Background(), -5)
	if !errors.Is(err, ErrInputTidakValid) {
		t.Fatalf("ingin ErrInputTidakValid, mendapat: %v", err)
	}
}

func TestTutupInputTidakValid(t *testing.T) {
	s := &PeriodeService{}

	err := s.Tutup(context.Background(), 0, time.Time{})
	if !errors.Is(err, ErrInputTidakValid) {
		t.Fatalf("ingin ErrInputTidakValid, mendapat: %v", err)
	}
}
