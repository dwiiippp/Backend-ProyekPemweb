package service

import (
	"context"
	"errors"
	"testing"

	"penggajian/internal/model"
)

func TestValidasiDasarKoreksi(t *testing.T) {
	tests := []struct {
		nama    string
		input   InputKoreksi
		wantErr bool
	}{
		{
			nama: "valid tambah",
			input: InputKoreksi{
				GajiSumberID:    1,
				PeriodeTujuanID: 2,
				JenisKoreksi:    model.KoreksiTambah,
				NominalText:     "50000.50",
				Alasan:          "Kekurangan upah lembur periode lalu",
			},
			wantErr: false,
		},
		{
			nama: "gaji sumber tidak valid",
			input: InputKoreksi{
				GajiSumberID:    0,
				PeriodeTujuanID: 2,
				JenisKoreksi:    model.KoreksiTambah,
				NominalText:     "50000",
				Alasan:          "Alasan",
			},
			wantErr: true,
		},
		{
			nama: "jenis koreksi tidak valid",
			input: InputKoreksi{
				GajiSumberID:    1,
				PeriodeTujuanID: 2,
				JenisKoreksi:    "bonus",
				NominalText:     "50000",
				Alasan:          "Alasan",
			},
			wantErr: true,
		},
		{
			nama: "nominal nol ditolak",
			input: InputKoreksi{
				GajiSumberID:    1,
				PeriodeTujuanID: 2,
				JenisKoreksi:    model.KoreksiKurang,
				NominalText:     "0",
				Alasan:          "Alasan",
			},
			wantErr: true,
		},
		{
			nama: "nominal lebih dari 2 desimal ditolak",
			input: InputKoreksi{
				GajiSumberID:    1,
				PeriodeTujuanID: 2,
				JenisKoreksi:    model.KoreksiKurang,
				NominalText:     "1000.123",
				Alasan:          "Alasan",
			},
			wantErr: true,
		},
		{
			nama: "alasan kosong ditolak",
			input: InputKoreksi{
				GajiSumberID:    1,
				PeriodeTujuanID: 2,
				JenisKoreksi:    model.KoreksiTambah,
				NominalText:     "25000",
				Alasan:          "   ",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.nama, func(t *testing.T) {
			_, _, err := ValidasiDasarKoreksi(tt.input)
			if tt.wantErr && !errors.Is(err, ErrInputTidakValid) {
				t.Fatalf("ingin ErrInputTidakValid, mendapat: %v", err)
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("tidak seharusnya error, mendapat: %v", err)
			}
		})
	}
}

func TestCocokKategoriPeriode(t *testing.T) {
	if !CocokKategoriPeriode(model.KategoriHarian, model.PeriodeMingguan) {
		t.Error("harian harus cocok dengan mingguan")
	}
	if !CocokKategoriPeriode(model.KategoriBulanan, model.PeriodeBulanan) {
		t.Error("bulanan harus cocok dengan bulanan")
	}
	if CocokKategoriPeriode(model.KategoriHarian, model.PeriodeBulanan) {
		t.Error("harian tidak boleh cocok dengan bulanan")
	}
	if CocokKategoriPeriode(model.KategoriBulanan, model.PeriodeMingguan) {
		t.Error("bulanan tidak boleh cocok dengan mingguan")
	}
}

func TestCekKepemilikanSlip(t *testing.T) {
	if err := CekKepemilikanSlip(3, 3); err != nil {
		t.Fatalf("slip milik sendiri seharusnya diizinkan: %v", err)
	}
	if err := CekKepemilikanSlip(3, 4); !errors.Is(err, ErrBukanMilikSendiri) {
		t.Fatalf("slip pekerja lain harus ErrBukanMilikSendiri, mendapat: %v", err)
	}
	if err := CekKepemilikanSlip(3, 0); !errors.Is(err, ErrBukanMilikSendiri) {
		t.Fatalf("user tanpa pekerja_id harus ErrBukanMilikSendiri, mendapat: %v", err)
	}
}

func TestLaporanGajiValidasiFilter(t *testing.T) {
	svc := &LaporanGajiService{}

	if _, err := svc.StatusPeriode(context.Background(), "harian", ""); !errors.Is(err, ErrInputTidakValid) {
		t.Errorf("jenis tidak valid harus ErrInputTidakValid, mendapat: %v", err)
	}
	if _, err := svc.StatusPeriode(context.Background(), "", "selesai"); !errors.Is(err, ErrInputTidakValid) {
		t.Errorf("status tidak valid harus ErrInputTidakValid, mendapat: %v", err)
	}
	if _, err := svc.RekapGaji(context.Background(), "", "2026-10-31", "2026-10-01"); !errors.Is(err, ErrInputTidakValid) {
		t.Errorf("rentang tanggal terbalik harus ErrInputTidakValid, mendapat: %v", err)
	}
}
