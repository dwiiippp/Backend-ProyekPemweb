package service

import (
	"testing"
	"time"

	"penggajian/internal/model"

	"github.com/shopspring/decimal"
)

func TestHitungGajiPokokHarian(t *testing.T) {
	input := InputKalkulatorGaji{
		Pekerja:     model.Pekerja{KategoriKerja: model.KategoriHarian},
		NominalUpah: decimal.NewFromInt(100000),
		HariHadir:   3,
	}

	hasil := HitungGajiPokok(input)
	expected := decimal.NewFromInt(300000)

	if !hasil.Equal(expected) {
		t.Errorf("hasil %s, seharusnya %s", hasil, expected)
	}
}

func TestHitungPenyesuaianProporsional(t *testing.T) {
	input := InputKalkulatorGaji{
		Pekerja:        model.Pekerja{KategoriKerja: model.KategoriBulanan},
		NominalUpah:    decimal.NewFromInt(3500000),
		HariAktif:      15,
		HariDalamBulan: 30,
	}

	hasil := HitungPenyesuaianProporsional(input)
	expected := decimal.NewFromInt(-1750000)

	if !hasil.Equal(expected) {
		t.Errorf("hasil %s, seharusnya %s", hasil, expected)
	}
}

func TestHitungPotonganTerlambat(t *testing.T) {
	menit20 := 20
	menit10 := 10

	absensi := []model.Absensi{
		{
			StatusValidasi: model.ValidasiDisetujui,
			MenitTerlambat: &menit20,
		},
		{
			StatusValidasi: model.ValidasiDisetujui,
			MenitTerlambat: &menit10,
		},
		{
			StatusValidasi: model.ValidasiDitolak,
			MenitTerlambat: &menit20,
		},
	}

	hasil := HitungPotonganTerlambat(
		model.Pekerja{KategoriKerja: model.KategoriHarian},
		decimal.NewFromInt(80000),
		absensi,
	)

	expected := decimal.NewFromInt(10000)

	if !hasil.Equal(expected) {
		t.Errorf("hasil %s, seharusnya %s", hasil, expected)
	}
}

func TestHitungUpahLembur(t *testing.T) {
	absensi := []model.Absensi{
		{
			StatusValidasi: model.ValidasiDisetujui,
			JamLembur: decimal.NullDecimal{
				Decimal: decimal.NewFromInt(2),
				Valid:   true,
			},
		},
	}

	hasil := HitungUpahLembur(
		model.Pekerja{KategoriKerja: model.KategoriHarian},
		decimal.NewFromInt(80000),
		absensi,
	)

	expected := decimal.NewFromInt(35000)

	if !hasil.Equal(expected) {
		t.Errorf("hasil %s, seharusnya %s", hasil, expected)
	}
}

func TestHitungPotonganKasbon(t *testing.T) {
	gajiKotor := decimal.NewFromInt(1000000)

	hasil := HitungPotonganKasbon(
		gajiKotor,
		decimal.NewFromInt(500000),
	)

	expected := decimal.NewFromInt(300000)

	if !hasil.Equal(expected) {
		t.Errorf("hasil %s, seharusnya %s", hasil, expected)
	}
}

func TestHitungPotonganKasbonSisaLebihKecil(t *testing.T) {
	gajiKotor := decimal.NewFromInt(1000000)

	hasil := HitungPotonganKasbon(
		gajiKotor,
		decimal.NewFromInt(100000),
	)

	expected := decimal.NewFromInt(100000)

	if !hasil.Equal(expected) {
		t.Errorf("hasil %s, seharusnya %s", hasil, expected)
	}
}

func TestValidasiHariDalamBulan(t *testing.T) {
	tests := []struct {
		nama           string
		hariDalamBulan int
		wantError      bool
	}{
		{
			nama:           "hari dalam bulan nol",
			hariDalamBulan: 0,
			wantError:      true,
		},
		{
			nama:           "hari dalam bulan negatif",
			hariDalamBulan: -1,
			wantError:      true,
		},
		{
			nama:           "hari dalam bulan valid",
			hariDalamBulan: 30,
			wantError:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.nama, func(t *testing.T) {
			input := InputKalkulatorGaji{
				Pekerja: model.Pekerja{
					KategoriKerja: model.KategoriBulanan,
				},
				NominalUpah:    decimal.NewFromInt(3500000),
				HariAktif:      15,
				HariDalamBulan: tt.hariDalamBulan,
			}

			err := ValidasiInputKalkulator(input)

			if tt.wantError && err == nil {
				t.Error("seharusnya menghasilkan error")
			}

			if !tt.wantError && err != nil {
				t.Errorf("seharusnya tidak error, tetapi mendapat: %v", err)
			}
		})
	}
}

func TestHitungPotonganKasbonGajiKotorTidakValid(t *testing.T) {
	tests := []struct {
		nama      string
		gajiKotor int64
	}{
		{nama: "gaji kotor nol", gajiKotor: 0},
		{nama: "gaji kotor negatif", gajiKotor: -100000},
	}

	for _, tt := range tests {
		t.Run(tt.nama, func(t *testing.T) {
			hasil := HitungPotonganKasbon(
				decimal.NewFromInt(tt.gajiKotor),
				decimal.NewFromInt(500000),
			)

			if !hasil.Equal(decimal.Zero) {
				t.Errorf("hasil %s, seharusnya 0", hasil)
			}
		})
	}
}

func TestHitungGajiPekerjaSkenarioSeed(t *testing.T) {
	tgl := func(y int, m time.Month, d int) time.Time {
		return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
	}
	intPtr := func(v int) *int { return &v }
	lemburDec := func(s string) decimal.NullDecimal {
		return decimal.NullDecimal{Decimal: decimal.RequireFromString(s), Valid: true}
	}

	riwayatBudi := []model.RiwayatGaji{
		{ID: 1, PekerjaID: 1, NominalGaji: decimal.NewFromInt(100000), TanggalBerlaku: tgl(2025, 1, 10)},
		{ID: 2, PekerjaID: 1, NominalGaji: decimal.NewFromInt(110000), TanggalBerlaku: tgl(2026, 10, 28)},
	}
	riwayatSiti := []model.RiwayatGaji{
		{ID: 3, PekerjaID: 2, NominalGaji: decimal.NewFromInt(100000), TanggalBerlaku: tgl(2025, 3, 1)},
	}
	riwayatAndi := []model.RiwayatGaji{
		{ID: 4, PekerjaID: 3, NominalGaji: decimal.NewFromInt(95000), TanggalBerlaku: tgl(2025, 6, 15)},
	}
	riwayatDewi := []model.RiwayatGaji{
		{ID: 5, PekerjaID: 4, NominalGaji: decimal.NewFromInt(3500000), TanggalBerlaku: tgl(2026, 10, 15)},
	}

	periode1 := model.PeriodeGaji{
		ID: 1, JenisPeriode: model.PeriodeMingguan,
		TanggalMulai: tgl(2026, 10, 19), TanggalSelesai: tgl(2026, 10, 24),
	}
	periode2 := model.PeriodeGaji{
		ID: 2, JenisPeriode: model.PeriodeMingguan,
		TanggalMulai: tgl(2026, 10, 26), TanggalSelesai: tgl(2026, 10, 31),
	}
	periode3 := model.PeriodeGaji{
		ID: 3, JenisPeriode: model.PeriodeBulanan,
		TanggalMulai: tgl(2026, 10, 1), TanggalSelesai: tgl(2026, 10, 31),
	}

	tests := []struct {
		nama               string
		input              InputKalkulatorGaji
		wantHadir          int
		wantPokok          string
		wantTerlambat      string
		wantLembur         string
		wantProporsional   string
		wantKotor          string
		wantPotonganKasbon string
		wantBersih         string
		wantRincianLen     int
		wantErr            bool
	}{
		{
			nama: "Periode 1 Budi: 6 hari hadir, terlambat 5 menit tidak dipotong, kasbon 500rb dipotong 30%",
			input: InputKalkulatorGaji{
				Pekerja:     model.Pekerja{ID: 1, KategoriKerja: model.KategoriHarian},
				Periode:     periode1,
				RiwayatGaji: riwayatBudi,
				Absensi: []model.Absensi{
					{TanggalAbsensi: tgl(2026, 10, 19), StatusValidasi: model.ValidasiDisetujui, MenitTerlambat: intPtr(0)},
					{TanggalAbsensi: tgl(2026, 10, 20), StatusValidasi: model.ValidasiDisetujui, MenitTerlambat: intPtr(0)},
					{TanggalAbsensi: tgl(2026, 10, 21), StatusValidasi: model.ValidasiDisetujui, MenitTerlambat: intPtr(0)},
					{TanggalAbsensi: tgl(2026, 10, 22), StatusValidasi: model.ValidasiDisetujui, MenitTerlambat: intPtr(5)},
					{TanggalAbsensi: tgl(2026, 10, 23), StatusValidasi: model.ValidasiDisetujui, MenitTerlambat: intPtr(0)},
					{TanggalAbsensi: tgl(2026, 10, 24), StatusValidasi: model.ValidasiDisetujui, MenitTerlambat: intPtr(0)},
				},
				KasbonAktif: []ItemKasbonAktif{
					{ID: 1, TanggalPengajuan: tgl(2026, 10, 12), Sisa: decimal.NewFromInt(500000)},
				},
			},
			wantHadir:          6,
			wantPokok:          "600000.00",
			wantTerlambat:      "0.00",
			wantLembur:         "0.00",
			wantProporsional:   "0.00",
			wantKotor:          "600000.00",
			wantPotonganKasbon: "180000.00",
			wantBersih:         "420000.00",
			wantRincianLen:     1,
		},
		{
			nama: "Periode 1 Siti: 6 hari hadir, lembur 2 jam, kasbon sisa 100rb lunas",
			input: InputKalkulatorGaji{
				Pekerja:     model.Pekerja{ID: 2, KategoriKerja: model.KategoriHarian},
				Periode:     periode1,
				RiwayatGaji: riwayatSiti,
				Absensi: []model.Absensi{
					{TanggalAbsensi: tgl(2026, 10, 19), StatusValidasi: model.ValidasiDisetujui},
					{TanggalAbsensi: tgl(2026, 10, 20), StatusValidasi: model.ValidasiDisetujui},
					{TanggalAbsensi: tgl(2026, 10, 21), StatusValidasi: model.ValidasiDisetujui, JamLembur: lemburDec("2.00")},
					{TanggalAbsensi: tgl(2026, 10, 22), StatusValidasi: model.ValidasiDisetujui},
					{TanggalAbsensi: tgl(2026, 10, 23), StatusValidasi: model.ValidasiDisetujui},
					{TanggalAbsensi: tgl(2026, 10, 24), StatusValidasi: model.ValidasiDisetujui},
				},
				KasbonAktif: []ItemKasbonAktif{
					{ID: 2, TanggalPengajuan: tgl(2026, 10, 5), Sisa: decimal.NewFromInt(100000)},
				},
			},
			wantHadir:          6,
			wantPokok:          "600000.00",
			wantTerlambat:      "0.00",
			wantLembur:         "43750.00",
			wantProporsional:   "0.00",
			wantKotor:          "643750.00",
			wantPotonganKasbon: "100000.00",
			wantBersih:         "543750.00",
			wantRincianLen:     1,
		},
		{
			nama: "Periode 2 Budi: upah berubah tengah periode (100rb -> 110rb mulai 28 Okt), terlambat 20 mnt, presensi ditolak diabaikan",
			input: InputKalkulatorGaji{
				Pekerja:     model.Pekerja{ID: 1, KategoriKerja: model.KategoriHarian},
				Periode:     periode2,
				RiwayatGaji: riwayatBudi,
				Absensi: []model.Absensi{
					{TanggalAbsensi: tgl(2026, 10, 26), StatusValidasi: model.ValidasiDisetujui, MenitTerlambat: intPtr(0)},
					{TanggalAbsensi: tgl(2026, 10, 27), StatusValidasi: model.ValidasiDisetujui, MenitTerlambat: intPtr(20)},
					{TanggalAbsensi: tgl(2026, 10, 28), StatusValidasi: model.ValidasiDisetujui, MenitTerlambat: intPtr(0)},
					{TanggalAbsensi: tgl(2026, 10, 29), StatusValidasi: model.ValidasiDitolak, MenitTerlambat: intPtr(30)},
					{TanggalAbsensi: tgl(2026, 10, 30), StatusValidasi: model.ValidasiDisetujui, MenitTerlambat: intPtr(0)},
				},
			},
			wantHadir:          4,
			wantPokok:          "420000.00",
			wantTerlambat:      "12500.00",
			wantLembur:         "0.00",
			wantProporsional:   "0.00",
			wantKotor:          "407500.00",
			wantPotonganKasbon: "0.00",
			wantBersih:         "407500.00",
		},
		{
			nama: "Periode 2 Andi: terlambat 10 mnt (tidak dipotong) + lembur 1 jam + 1 hari tidak hadir",
			input: InputKalkulatorGaji{
				Pekerja:     model.Pekerja{ID: 3, KategoriKerja: model.KategoriHarian},
				Periode:     periode2,
				RiwayatGaji: riwayatAndi,
				Absensi: []model.Absensi{
					{TanggalAbsensi: tgl(2026, 10, 26), StatusValidasi: model.ValidasiDisetujui},
					{TanggalAbsensi: tgl(2026, 10, 27), StatusValidasi: model.ValidasiDisetujui},
					{TanggalAbsensi: tgl(2026, 10, 28), StatusValidasi: model.ValidasiDisetujui},
					{TanggalAbsensi: tgl(2026, 10, 29), StatusValidasi: model.ValidasiDisetujui, MenitTerlambat: intPtr(10), JamLembur: lemburDec("1.00")},
				},
			},
			wantHadir:          4,
			wantPokok:          "380000.00",
			wantTerlambat:      "0.00",
			wantLembur:         "17812.50",
			wantProporsional:   "0.00",
			wantKotor:          "397812.50",
			wantPotonganKasbon: "0.00",
			wantBersih:         "397812.50",
		},
		{
			nama: "Periode 3 Dewi: pekerja bulanan masuk 15 Okt (17/31 hari aktif) + koreksi tambah 50rb",
			input: InputKalkulatorGaji{
				Pekerja: model.Pekerja{
					ID:            4,
					KategoriKerja: model.KategoriBulanan,
					TanggalMasuk:  tgl(2026, 10, 15),
				},
				Periode:     periode3,
				RiwayatGaji: riwayatDewi,
				Koreksi:     decimal.NewFromInt(50000),
				Absensi: []model.Absensi{
					{TanggalAbsensi: tgl(2026, 10, 19), StatusValidasi: model.ValidasiDisetujui},
					{TanggalAbsensi: tgl(2026, 10, 20), StatusValidasi: model.ValidasiDisetujui},
				},
			},
			wantHadir:          2,
			wantPokok:          "3500000.00",
			wantTerlambat:      "0.00",
			wantLembur:         "0.00",
			wantProporsional:   "-1580645.16",
			wantKotor:          "1969354.84",
			wantPotonganKasbon: "0.00",
			wantBersih:         "1969354.84",
		},
		{
			nama: "Gagal jika riwayat upah tidak tersedia",
			input: InputKalkulatorGaji{
				Pekerja: model.Pekerja{ID: 99, KategoriKerja: model.KategoriHarian},
				Periode: periode1,
				Absensi: []model.Absensi{
					{TanggalAbsensi: tgl(2026, 10, 19), StatusValidasi: model.ValidasiDisetujui},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.nama, func(t *testing.T) {
			hasil, err := HitungGajiPekerja(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("seharusnya error, tetapi berhasil")
				}
				return
			}
			if err != nil {
				t.Fatalf("error tidak terduga: %v", err)
			}

			assertDec := func(field string, got decimal.Decimal, want string) {
				t.Helper()
				if !got.Equal(decimal.RequireFromString(want)) {
					t.Errorf("%s = %s, ingin %s", field, got.StringFixed(2), want)
				}
			}

			if hasil.JumlahHariHadir != tt.wantHadir {
				t.Errorf("JumlahHariHadir = %d, ingin %d", hasil.JumlahHariHadir, tt.wantHadir)
			}
			assertDec("GajiPokok", hasil.GajiPokok, tt.wantPokok)
			assertDec("PotonganTerlambat", hasil.PotonganTerlambat, tt.wantTerlambat)
			assertDec("UpahLembur", hasil.UpahLembur, tt.wantLembur)
			assertDec("PenyesuaianProporsional", hasil.PenyesuaianProporsional, tt.wantProporsional)
			assertDec("GajiKotor", hasil.GajiKotor, tt.wantKotor)
			assertDec("PotonganKasbon", hasil.PotonganKasbon, tt.wantPotonganKasbon)
			assertDec("GajiBersih", hasil.GajiBersih, tt.wantBersih)
			if len(hasil.RincianKasbon) != tt.wantRincianLen {
				t.Errorf("len(RincianKasbon) = %d, ingin %d", len(hasil.RincianKasbon), tt.wantRincianLen)
			}
		})
	}
}

func TestAlokasiPotonganKasbonFIFO(t *testing.T) {
	tgl := func(d int) time.Time {
		return time.Date(2026, 10, d, 0, 0, 0, 0, time.UTC)
	}
	// Sengaja tidak urut untuk menguji pengurutan FIFO berdasarkan tanggal_pengajuan ASC
	kasbon := []ItemKasbonAktif{
		{ID: 20, TanggalPengajuan: tgl(15), Sisa: decimal.NewFromInt(250000)},
		{ID: 10, TanggalPengajuan: tgl(5), Sisa: decimal.NewFromInt(200000)},
	}
	// 30% dari 1.000.000 = 300.000 -> kasbon ID 10 terpotong 200.000 (lunas), ID 20 terpotong 100.000 (sisa 150.000)
	total, rincian := AlokasiPotonganKasbon(decimal.NewFromInt(1000000), kasbon)
	if !total.Equal(decimal.NewFromInt(300000)) {
		t.Fatalf("total potongan = %s, ingin 300000", total)
	}
	if len(rincian) != 2 {
		t.Fatalf("jumlah rincian = %d, ingin 2", len(rincian))
	}
	if rincian[0].KasbonID != 10 || !rincian[0].NominalPotongan.Equal(decimal.NewFromInt(200000)) || !rincian[0].Lunas {
		t.Errorf("rincian[0] tidak sesuai FIFO: %+v", rincian[0])
	}
	if rincian[1].KasbonID != 20 || !rincian[1].NominalPotongan.Equal(decimal.NewFromInt(100000)) || rincian[1].Lunas || !rincian[1].SisaSesudah.Equal(decimal.NewFromInt(150000)) {
		t.Errorf("rincian[1] tidak sesuai FIFO: %+v", rincian[1])
	}
}
