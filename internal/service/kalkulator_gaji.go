package service

import (
	"errors"
	"sort"
	"time"

	"penggajian/internal/model"

	"github.com/shopspring/decimal"
)

// ItemKasbonAktif merepresentasikan kasbon aktif untuk alokasi FIFO (BR-09).
type ItemKasbonAktif struct {
	ID               int64
	TanggalPengajuan time.Time
	Sisa             decimal.Decimal
}

// RincianPotonganKasbon menyimpan hasil pemotongan per kasbon.
type RincianPotonganKasbon struct {
	KasbonID        int64
	NominalPotongan decimal.Decimal
	SisaSesudah     decimal.Decimal
	Lunas           bool
}

// InputKalkulatorGaji menampung seluruh input murni untuk menghitung gaji satu pekerja.
type InputKalkulatorGaji struct {
	Pekerja        model.Pekerja
	Periode        model.PeriodeGaji
	RiwayatGaji    []model.RiwayatGaji
	NominalUpah    decimal.Decimal // opsional bila RiwayatGaji kosong (untuk unit test sederhana)
	HariHadir      int
	HariAktif      int
	HariDalamBulan int
	Absensi        []model.Absensi
	Koreksi        decimal.Decimal
	KasbonAktif    []ItemKasbonAktif
}

// HasilKalkulatorGaji menampung hasil perhitungan gaji satu pekerja.
type HasilKalkulatorGaji struct {
	GajiPokok               decimal.Decimal
	JumlahHariHadir         int
	PotonganTerlambat       decimal.Decimal
	UpahLembur              decimal.Decimal
	PenyesuaianProporsional decimal.Decimal
	TotalKoreksi            decimal.Decimal
	GajiKotor               decimal.Decimal
	PotonganKasbon          decimal.Decimal
	GajiBersih              decimal.Decimal
	RincianKasbon           []RincianPotonganKasbon
}

// CariUpahBerlaku mencari nominal upah yang berlaku pada tanggal tertentu (BR-04).
// Mengambil baris dengan tanggal_berlaku <= tanggal yang paling baru.
func CariUpahBerlaku(riwayat []model.RiwayatGaji, tanggal time.Time) (decimal.Decimal, bool) {
	target := normTanggal(tanggal)
	var terpilih model.RiwayatGaji
	ketemu := false

	for _, r := range riwayat {
		tglBerlaku := normTanggal(r.TanggalBerlaku)
		if tglBerlaku.After(target) {
			continue
		}
		if !ketemu {
			terpilih = r
			ketemu = true
			continue
		}
		tglTerpilih := normTanggal(terpilih.TanggalBerlaku)
		if tglBerlaku.After(tglTerpilih) || (tglBerlaku.Equal(tglTerpilih) && r.ID > terpilih.ID) {
			terpilih = r
		}
	}

	if !ketemu {
		return decimal.Zero, false
	}
	return terpilih.NominalGaji, true
}

// HitungHariBulanan menghitung jumlah hari aktif pekerja dalam satu periode bulanan (BR-03).
func HitungHariBulanan(pekerja model.Pekerja, tanggalMulai, tanggalSelesai time.Time) (int, int) {
	mulai := normTanggal(tanggalMulai)
	selesai := normTanggal(tanggalSelesai)
	hariDalamBulan := time.Date(mulai.Year(), mulai.Month()+1, 0, 0, 0, 0, 0, time.UTC).Day()

	awal := mulai
	if !pekerja.TanggalMasuk.IsZero() {
		masuk := normTanggal(pekerja.TanggalMasuk)
		if masuk.After(awal) {
			awal = masuk
		}
	}

	akhir := selesai
	if pekerja.TanggalKeluar != nil && !pekerja.TanggalKeluar.IsZero() {
		keluar := normTanggal(*pekerja.TanggalKeluar)
		if keluar.Before(akhir) {
			akhir = keluar
		}
	}

	if akhir.Before(awal) {
		return 0, hariDalamBulan
	}

	hariAktif := int(akhir.Sub(awal)/(24*time.Hour)) + 1
	return hariAktif, hariDalamBulan
}

// AlokasiPotonganKasbon menghitung total potongan kasbon (maks 30% gaji kotor)
// dan mengalokasikannya ke kasbon aktif tertua lebih dahulu (FIFO, BR-09).
func AlokasiPotonganKasbon(gajiKotor decimal.Decimal, kasbon []ItemKasbonAktif) (decimal.Decimal, []RincianPotonganKasbon) {
	if !gajiKotor.IsPositive() || len(kasbon) == 0 {
		return decimal.Zero, nil
	}

	urutan := make([]ItemKasbonAktif, len(kasbon))
	copy(urutan, kasbon)
	sort.Slice(urutan, func(i, j int) bool {
		ti := normTanggal(urutan[i].TanggalPengajuan)
		tj := normTanggal(urutan[j].TanggalPengajuan)
		if !ti.Equal(tj) {
			return ti.Before(tj)
		}
		return urutan[i].ID < urutan[j].ID
	})

	sisaKuota := HitungBatasPotonganKasbon(gajiKotor).Round(2)
	totalPotongan := decimal.Zero
	var rincian []RincianPotonganKasbon

	for _, item := range urutan {
		if !sisaKuota.IsPositive() {
			break
		}
		if !item.Sisa.IsPositive() {
			continue
		}

		potong := item.Sisa
		if potong.GreaterThan(sisaKuota) {
			potong = sisaKuota
		}
		potong = potong.Round(2)
		if !potong.IsPositive() {
			continue
		}

		sisaSesudah := item.Sisa.Sub(potong).Round(2)
		if sisaSesudah.IsNegative() {
			sisaSesudah = decimal.Zero
		}

		rincian = append(rincian, RincianPotonganKasbon{
			KasbonID:        item.ID,
			NominalPotongan: potong,
			SisaSesudah:     sisaSesudah,
			Lunas:           sisaSesudah.IsZero(),
		})

		totalPotongan = totalPotongan.Add(potong)
		sisaKuota = sisaKuota.Sub(potong)
	}

	return totalPotongan.Round(2), rincian
}

// HitungGajiPekerja menghitung seluruh komponen gaji satu pekerja secara murni (F-07, BR-03..BR-09).
func HitungGajiPekerja(input InputKalkulatorGaji) (HasilKalkulatorGaji, error) {
	if input.Pekerja.KategoriKerja != model.KategoriHarian &&
		input.Pekerja.KategoriKerja != model.KategoriBulanan {
		return HasilKalkulatorGaji{}, errors.New("kategori pekerja tidak valid")
	}
	if input.NominalUpah.IsNegative() {
		return HasilKalkulatorGaji{}, errors.New("nominal upah tidak boleh negatif")
	}

	var disetujui []model.Absensi
	for _, a := range input.Absensi {
		if a.StatusValidasi == model.ValidasiDisetujui {
			disetujui = append(disetujui, a)
		}
	}

	gajiPokok := decimal.Zero
	potonganTerlambat := decimal.Zero
	upahLembur := decimal.Zero
	penyesuaianProporsional := decimal.Zero

	if input.Pekerja.KategoriKerja == model.KategoriHarian {
		for _, a := range disetujui {
			upahHari, ok := resolveUpah(input.RiwayatGaji, input.NominalUpah, a.TanggalAbsensi)
			if !ok {
				return HasilKalkulatorGaji{}, errors.New("riwayat upah pekerja harian tidak ditemukan pada tanggal absensi")
			}
			if upahHari.IsNegative() {
				return HasilKalkulatorGaji{}, errors.New("nominal upah tidak boleh negatif")
			}

			gajiPokok = gajiPokok.Add(upahHari)
			tarifPerJam := upahHari.Div(decimal.NewFromInt(8))
			potonganTerlambat = potonganTerlambat.Add(hitungPotonganSatuHari(a, tarifPerJam))
			upahLembur = upahLembur.Add(hitungLemburSatuHari(a, tarifPerJam))
		}
	} else {
		refDate := input.Periode.TanggalSelesai
		if refDate.IsZero() && len(disetujui) > 0 {
			refDate = disetujui[len(disetujui)-1].TanggalAbsensi
		}
		if refDate.IsZero() {
			refDate = time.Now()
		}

		gajiBulanan, ok := resolveUpah(input.RiwayatGaji, input.NominalUpah, refDate)
		if !ok {
			return HasilKalkulatorGaji{}, errors.New("riwayat gaji pekerja bulanan tidak ditemukan")
		}
		if gajiBulanan.IsNegative() {
			return HasilKalkulatorGaji{}, errors.New("nominal gaji bulanan tidak boleh negatif")
		}

		gajiPokok = gajiBulanan

		hariAktif := input.HariAktif
		hariDalamBulan := input.HariDalamBulan
		if hariDalamBulan <= 0 && !input.Periode.TanggalMulai.IsZero() && !input.Periode.TanggalSelesai.IsZero() {
			hariAktif, hariDalamBulan = HitungHariBulanan(input.Pekerja, input.Periode.TanggalMulai, input.Periode.TanggalSelesai)
		}
		if hariDalamBulan <= 0 {
			return HasilKalkulatorGaji{}, errors.New("jumlah hari dalam bulan harus lebih dari 0 untuk pekerja bulanan")
		}
		if hariAktif < hariDalamBulan {
			gajiProp := gajiBulanan.
				Mul(decimal.NewFromInt(int64(hariAktif))).
				Div(decimal.NewFromInt(int64(hariDalamBulan)))
			penyesuaianProporsional = gajiProp.Sub(gajiBulanan)
		}

		for _, a := range disetujui {
			upahBulanHari, ok := resolveUpah(input.RiwayatGaji, gajiBulanan, a.TanggalAbsensi)
			if !ok {
				upahBulanHari = gajiBulanan
			}
			tarifPerJam := upahBulanHari.Div(decimal.NewFromInt(173))
			potonganTerlambat = potonganTerlambat.Add(hitungPotonganSatuHari(a, tarifPerJam))
			upahLembur = upahLembur.Add(hitungLemburSatuHari(a, tarifPerJam))
		}
	}

	gajiPokok = gajiPokok.Round(2)
	potonganTerlambat = potonganTerlambat.Round(2)
	upahLembur = upahLembur.Round(2)
	penyesuaianProporsional = penyesuaianProporsional.Round(2)
	totalKoreksi := input.Koreksi.Round(2)

	gajiKotor := gajiPokok.
		Sub(potonganTerlambat).
		Add(upahLembur).
		Add(penyesuaianProporsional).
		Add(totalKoreksi).
		Round(2)

	potonganKasbon, rincianKasbon := AlokasiPotonganKasbon(gajiKotor, input.KasbonAktif)
	gajiBersih := gajiKotor.Sub(potonganKasbon).Round(2)

	return HasilKalkulatorGaji{
		GajiPokok:               gajiPokok,
		JumlahHariHadir:         len(disetujui),
		PotonganTerlambat:       potonganTerlambat,
		UpahLembur:              upahLembur,
		PenyesuaianProporsional: penyesuaianProporsional,
		TotalKoreksi:            totalKoreksi,
		GajiKotor:               gajiKotor,
		PotonganKasbon:          potonganKasbon,
		GajiBersih:              gajiBersih,
		RincianKasbon:           rincianKasbon,
	}, nil
}

func resolveUpah(riwayat []model.RiwayatGaji, fallback decimal.Decimal, tanggal time.Time) (decimal.Decimal, bool) {
	if len(riwayat) > 0 {
		return CariUpahBerlaku(riwayat, tanggal)
	}
	if fallback.GreaterThan(decimal.Zero) {
		return fallback, true
	}
	return decimal.Zero, false
}

func hitungPotonganSatuHari(data model.Absensi, tarifPerJam decimal.Decimal) decimal.Decimal {
	if data.MenitTerlambat != nil && *data.MenitTerlambat > 15 {
		return tarifPerJam
	}
	return decimal.Zero
}

func hitungLemburSatuHari(data model.Absensi, tarifPerJam decimal.Decimal) decimal.Decimal {
	if !data.JamLembur.Valid || !data.JamLembur.Decimal.IsPositive() {
		return decimal.Zero
	}

	jamLembur := data.JamLembur.Decimal
	jamPertama := jamLembur
	satu := decimal.NewFromInt(1)
	if jamPertama.GreaterThan(satu) {
		jamPertama = satu
	}

	jamBerikutnya := jamLembur.Sub(jamPertama)
	if jamBerikutnya.IsNegative() {
		jamBerikutnya = decimal.Zero
	}

	upahJamPertama := jamPertama.Mul(tarifPerJam).Mul(decimal.RequireFromString("1.5"))
	upahJamBerikutnya := jamBerikutnya.Mul(tarifPerJam).Mul(decimal.NewFromInt(2))
	return upahJamPertama.Add(upahJamBerikutnya)
}

// ValidasiInputKalkulator memvalidasi input dasar kalkulator gaji.
func ValidasiInputKalkulator(input InputKalkulatorGaji) error {
	if input.NominalUpah.IsNegative() {
		return errors.New("nominal upah tidak boleh negatif")
	}

	if input.HariHadir < 0 {
		return errors.New("jumlah hari hadir tidak boleh negatif")
	}

	if input.HariAktif < 0 {
		return errors.New("jumlah hari aktif tidak boleh negatif")
	}

	if input.Pekerja.KategoriKerja == model.KategoriBulanan &&
		input.HariDalamBulan <= 0 {
		return errors.New("jumlah hari dalam bulan harus lebih dari 0 untuk pekerja bulanan")
	}

	if input.Pekerja.KategoriKerja != model.KategoriHarian &&
		input.Pekerja.KategoriKerja != model.KategoriBulanan {
		return errors.New("kategori pekerja tidak valid")
	}

	return nil
}

// HitungGajiPokok menghitung gaji pokok berdasarkan kategori pekerja.
func HitungGajiPokok(input InputKalkulatorGaji) decimal.Decimal {
	if input.Pekerja.KategoriKerja == model.KategoriHarian {
		return input.NominalUpah.Mul(
			decimal.NewFromInt(int64(input.HariHadir)),
		)
	}

	return input.NominalUpah
}

// HitungPenyesuaianProporsional menghitung penyesuaian
// untuk pekerja bulanan yang aktif sebagian bulan.
func HitungPenyesuaianProporsional(input InputKalkulatorGaji) decimal.Decimal {
	if input.Pekerja.KategoriKerja != model.KategoriBulanan {
		return decimal.Zero
	}

	if input.HariDalamBulan <= 0 || input.HariAktif <= 0 {
		return decimal.Zero
	}

	gajiProporsional := input.NominalUpah.
		Mul(decimal.NewFromInt(int64(input.HariAktif))).
		Div(decimal.NewFromInt(int64(input.HariDalamBulan)))

	return gajiProporsional.Sub(input.NominalUpah)
}

// HitungPotonganTerlambat menghitung potongan keterlambatan.
// Potongan berlaku jika keterlambatan lebih dari 15 menit.
func HitungPotonganTerlambat(
	pekerja model.Pekerja,
	nominalUpah decimal.Decimal,
	absensi []model.Absensi,
) decimal.Decimal {
	var tarifPerJam decimal.Decimal

	if pekerja.KategoriKerja == model.KategoriHarian {
		tarifPerJam = nominalUpah.Div(decimal.NewFromInt(8))
	} else {
		tarifPerJam = nominalUpah.Div(decimal.NewFromInt(173))
	}

	totalPotongan := decimal.Zero

	for _, data := range absensi {
		if data.StatusValidasi != model.ValidasiDisetujui {
			continue
		}
		totalPotongan = totalPotongan.Add(hitungPotonganSatuHari(data, tarifPerJam))
	}

	return totalPotongan
}

// HitungUpahLembur menghitung upah lembur dari seluruh absensi.
func HitungUpahLembur(
	pekerja model.Pekerja,
	nominalUpah decimal.Decimal,
	absensi []model.Absensi,
) decimal.Decimal {
	var tarifPerJam decimal.Decimal

	if pekerja.KategoriKerja == model.KategoriHarian {
		tarifPerJam = nominalUpah.Div(decimal.NewFromInt(8))
	} else {
		tarifPerJam = nominalUpah.Div(decimal.NewFromInt(173))
	}

	totalLembur := decimal.Zero

	for _, data := range absensi {
		if data.StatusValidasi != model.ValidasiDisetujui {
			continue
		}
		totalLembur = totalLembur.Add(hitungLemburSatuHari(data, tarifPerJam))
	}

	return totalLembur
}

// HitungBatasPotonganKasbon menghitung batas potongan kasbon,
// maksimal 30 persen dari gaji kotor.
func HitungBatasPotonganKasbon(gajiKotor decimal.Decimal) decimal.Decimal {
	if gajiKotor.IsNegative() || gajiKotor.IsZero() {
		return decimal.Zero
	}

	return gajiKotor.Mul(decimal.RequireFromString("0.30"))
}

// HitungPotonganKasbon menentukan jumlah kasbon
// yang boleh dipotong dari gaji pada periode ini.
func HitungPotonganKasbon(
	gajiKotor decimal.Decimal,
	totalSisaKasbon decimal.Decimal,
) decimal.Decimal {
	if gajiKotor.IsNegative() || gajiKotor.IsZero() {
		return decimal.Zero
	}

	if totalSisaKasbon.IsNegative() || totalSisaKasbon.IsZero() {
		return decimal.Zero
	}

	batasPotongan := HitungBatasPotonganKasbon(gajiKotor)

	if totalSisaKasbon.LessThan(batasPotongan) {
		return totalSisaKasbon
	}

	return batasPotongan
}
