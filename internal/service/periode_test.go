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

func TestKategoriUntukPeriode(t *testing.T) {
	tests := []struct {
		nama         string
		jenisPeriode string
		want         string
	}{
		{
			nama:         "periode mingguan untuk pekerja harian",
			jenisPeriode: model.PeriodeMingguan,
			want:         model.KategoriHarian,
		},
		{
			nama:         "periode bulanan untuk pekerja bulanan",
			jenisPeriode: model.PeriodeBulanan,
			want:         model.KategoriBulanan,
		},
	}

	for _, tt := range tests {
		t.Run(tt.nama, func(t *testing.T) {
			got := kategoriUntukPeriode(tt.jenisPeriode)

			if got != tt.want {
				t.Errorf(
					"kategoriUntukPeriode(%q) = %q, ingin %q",
					tt.jenisPeriode,
					got,
					tt.want,
				)
			}
		})
	}
}

func TestValidasiDanNormalisasiPeriodeValid(t *testing.T) {
	tgl := func(y int, m time.Month, d int) time.Time {
		return time.Date(y, m, d, 0, 0, 0, 0, time.Local)
	}

	tests := []struct {
		nama        string
		periode     model.PeriodeGaji
		wantMulai   time.Time
		wantSelesai time.Time
		wantBayar   *time.Time
	}{
		{
			nama: "periode mingguan valid",
			periode: model.PeriodeGaji{
				JenisPeriode:   model.PeriodeMingguan,
				TanggalMulai:   tgl(2026, 10, 19),
				TanggalSelesai: tgl(2026, 10, 24),
				Status:         model.PeriodeDitutup,
			},
			wantMulai:   tgl(2026, 10, 19),
			wantSelesai: tgl(2026, 10, 24),
		},
		{
			nama: "periode bulanan valid",
			periode: model.PeriodeGaji{
				JenisPeriode:   model.PeriodeBulanan,
				TanggalMulai:   tgl(2026, 10, 1),
				TanggalSelesai: tgl(2026, 10, 31),
			},
			wantMulai:   tgl(2026, 10, 1),
			wantSelesai: tgl(2026, 10, 31),
		},
		{
			nama: "periode mingguan dengan tanggal pembayaran valid",
			periode: model.PeriodeGaji{
				JenisPeriode:   model.PeriodeMingguan,
				TanggalMulai:   tgl(2026, 10, 19),
				TanggalSelesai: tgl(2026, 10, 24),
				TanggalPembayaran: func() *time.Time {
					v := tgl(2026, 10, 25)
					return &v
				}(),
			},
			wantMulai:   tgl(2026, 10, 19),
			wantSelesai: tgl(2026, 10, 24),
			wantBayar: func() *time.Time {
				v := tgl(2026, 10, 25)
				return &v
			}(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.nama, func(t *testing.T) {
			got, err := validasiDanNormalisasiPeriode(tt.periode)

			if err != nil {
				t.Fatalf("seharusnya valid, mendapat error: %v", err)
			}

			if !got.TanggalMulai.Equal(normTanggal(tt.wantMulai)) {
				t.Errorf("tanggal mulai tidak sesuai")
			}

			if !got.TanggalSelesai.Equal(normTanggal(tt.wantSelesai)) {
				t.Errorf("tanggal selesai tidak sesuai")
			}

			if got.Status != model.PeriodeTerbuka {
				t.Errorf("status seharusnya terbuka, mendapat %q", got.Status)
			}

			if tt.wantBayar == nil {
				if got.TanggalPembayaran != nil {
					t.Errorf("tanggal pembayaran seharusnya nil")
				}
			} else {
				if got.TanggalPembayaran == nil ||
					!got.TanggalPembayaran.Equal(normTanggal(*tt.wantBayar)) {
					t.Errorf("tanggal pembayaran tidak sesuai")
				}
			}
		})
	}
}
