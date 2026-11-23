package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"penggajian/internal/model"
	"penggajian/internal/repository"

	"github.com/shopspring/decimal"
)

var ErrPrasyaratPenggajian = errors.New("prasyarat proses penggajian belum terpenuhi")

// Pemilik: Anggota 3 | F-07 · F-09 · UC-12 Proses Penggajian · UC-13 Tutup Periode
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
type PenggajianService struct {
	repo *repository.PenggajianRepository
}

func NewPenggajianService(repo *repository.PenggajianRepository) *PenggajianService {
	return &PenggajianService{repo: repo}
}

type ItemPratinjauGaji struct {
	Pekerja                 model.Pekerja
	JumlahHariHadir         int
	GajiPokok               decimal.Decimal
	PotonganTerlambat       decimal.Decimal
	UpahLembur              decimal.Decimal
	PenyesuaianProporsional decimal.Decimal
	TotalKoreksi            decimal.Decimal
	GajiKotor               decimal.Decimal
	PotonganKasbon          decimal.Decimal
	GajiBersih              decimal.Decimal
}

type HasilPratinjau struct {
	Periode             model.PeriodeGaji
	Items               []ItemPratinjauGaji
	Kesalahan           []string
	BisaDikonfirmasi    bool
	TotalGajiKotor      decimal.Decimal
	TotalPotonganKasbon decimal.Decimal
	TotalGajiBersih     decimal.Decimal
}

type HasilPeriodeGaji struct {
	Periode             model.PeriodeGaji
	Items               []repository.BarisHasilGaji
	TotalGajiKotor      decimal.Decimal
	TotalPotonganKasbon decimal.Decimal
	TotalGajiBersih     decimal.Decimal
}

func (s *PenggajianService) DaftarPeriode(ctx context.Context) ([]model.PeriodeGaji, error) {
	return s.repo.DaftarPeriode(ctx)
}

// Pratinjau menghitung simulasi gaji seluruh pekerja yang memenuhi syarat untuk satu periode (UC-12).
// Jika ada hari yang belum divalidasi atau presensi menunggu, daftar kesalahan dikembalikan.
func (s *PenggajianService) Pratinjau(ctx context.Context, periodeID int64) (HasilPratinjau, error) {
	if periodeID <= 0 {
		return HasilPratinjau{}, ErrInputTidakValid
	}

	periode, err := s.repo.AmbilPeriodeByID(ctx, periodeID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return HasilPratinjau{}, ErrTidakDitemukan
		}
		return HasilPratinjau{}, err
	}
	if periode.Status == model.PeriodeDitutup {
		return HasilPratinjau{}, ErrPeriodeDitutup
	}

	var kesalahan []string

	belumValidasi, err := s.repo.TanggalBelumDivalidasi(ctx, periode.TanggalMulai, periode.TanggalSelesai)
	if err != nil {
		return HasilPratinjau{}, err
	}
	for _, tgl := range belumValidasi {
		kesalahan = append(kesalahan, fmt.Sprintf("Presensi tanggal %s belum divalidasi harian.", tgl.Format("02-01-2006")))
	}

	menunggu, err := s.repo.PresensiMenunggu(ctx, periode.TanggalMulai, periode.TanggalSelesai)
	if err != nil {
		return HasilPratinjau{}, err
	}
	for _, m := range menunggu {
		kesalahan = append(kesalahan, fmt.Sprintf("Masih ada %d presensi berstatus menunggu pada tanggal %s.", m.Jumlah, m.Tanggal.Format("02-01-2006")))
	}

	kategori := kategoriUntukPeriode(periode.JenisPeriode)
	pekerjaList, err := s.repo.ListPekerjaEligible(ctx, kategori, periode.TanggalMulai, periode.TanggalSelesai)
	if err != nil {
		return HasilPratinjau{}, err
	}
	if len(pekerjaList) == 0 {
		kesalahan = append(kesalahan, fmt.Sprintf("Tidak ada pekerja aktif berkategori %s dalam rentang periode ini.", kategori))
	}

	items := make([]ItemPratinjauGaji, 0, len(pekerjaList))
	totalKotor := decimal.Zero
	totalKasbon := decimal.Zero
	totalBersih := decimal.Zero

	for _, p := range pekerjaList {
		absensi, err := s.repo.ListAbsensiDisetujui(ctx, p.ID, periode.TanggalMulai, periode.TanggalSelesai)
		if err != nil {
			return HasilPratinjau{}, err
		}

		riwayat, err := s.repo.ListRiwayatGajiPekerja(ctx, p.ID, periode.TanggalSelesai)
		if err != nil {
			return HasilPratinjau{}, err
		}

		koreksi, err := s.repo.TotalKoreksiPeriode(ctx, periode.ID, p.ID)
		if err != nil {
			return HasilPratinjau{}, err
		}

		kasbonDB, err := s.repo.ListKasbonAktif(ctx, p.ID)
		if err != nil {
			return HasilPratinjau{}, err
		}

		hasil, err := HitungGajiPekerja(InputKalkulatorGaji{
			Pekerja:     p,
			Periode:     periode,
			RiwayatGaji: riwayat,
			Absensi:     absensi,
			Koreksi:     koreksi,
			KasbonAktif: konversiKasbonAktif(kasbonDB),
		})
		if err != nil {
			kesalahan = append(kesalahan, fmt.Sprintf("Pekerja %s (%s): %s.", p.Nama, p.NIK, err.Error()))
			continue
		}

		items = append(items, ItemPratinjauGaji{
			Pekerja:                 p,
			JumlahHariHadir:         hasil.JumlahHariHadir,
			GajiPokok:               hasil.GajiPokok,
			PotonganTerlambat:       hasil.PotonganTerlambat,
			UpahLembur:              hasil.UpahLembur,
			PenyesuaianProporsional: hasil.PenyesuaianProporsional,
			TotalKoreksi:            hasil.TotalKoreksi,
			GajiKotor:               hasil.GajiKotor,
			PotonganKasbon:          hasil.PotonganKasbon,
			GajiBersih:              hasil.GajiBersih,
		})

		totalKotor = totalKotor.Add(hasil.GajiKotor)
		totalKasbon = totalKasbon.Add(hasil.PotonganKasbon)
		totalBersih = totalBersih.Add(hasil.GajiBersih)
	}

	return HasilPratinjau{
		Periode:             periode,
		Items:               items,
		Kesalahan:           kesalahan,
		BisaDikonfirmasi:    len(kesalahan) == 0 && len(items) > 0,
		TotalGajiKotor:      totalKotor,
		TotalPotonganKasbon: totalKasbon,
		TotalGajiBersih:     totalBersih,
	}, nil
}

// KonfirmasiDanTutup memproses penggajian seluruh pekerja dan menutup periode dalam SATU transaksi (UC-13 / F-09).
func (s *PenggajianService) KonfirmasiDanTutup(ctx context.Context, periodeID int64, tanggalPembayaran time.Time) error {
	if periodeID <= 0 {
		return ErrInputTidakValid
	}

	tx, err := s.repo.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	periode, err := s.repo.AmbilPeriodeByIDForUpdateTx(ctx, tx, periodeID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrTidakDitemukan
		}
		return err
	}
	if periode.Status == model.PeriodeDitutup {
		return ErrPeriodeDitutup
	}

	belumValidasi, err := s.repo.TanggalBelumDivalidasiTx(ctx, tx, periode.TanggalMulai, periode.TanggalSelesai)
	if err != nil {
		return err
	}
	if len(belumValidasi) > 0 {
		return ErrPrasyaratPenggajian
	}

	menunggu, err := s.repo.PresensiMenungguTx(ctx, tx, periode.TanggalMulai, periode.TanggalSelesai)
	if err != nil {
		return err
	}
	if len(menunggu) > 0 {
		return ErrPrasyaratPenggajian
	}

	kategori := kategoriUntukPeriode(periode.JenisPeriode)
	pekerjaList, err := s.repo.ListPekerjaEligibleTx(ctx, tx, kategori, periode.TanggalMulai, periode.TanggalSelesai)
	if err != nil {
		return err
	}
	if len(pekerjaList) == 0 {
		return ErrPrasyaratPenggajian
	}

	for _, p := range pekerjaList {
		absensi, err := s.repo.ListAbsensiDisetujuiTx(ctx, tx, p.ID, periode.TanggalMulai, periode.TanggalSelesai)
		if err != nil {
			return err
		}

		riwayat, err := s.repo.ListRiwayatGajiPekerjaTx(ctx, tx, p.ID, periode.TanggalSelesai)
		if err != nil {
			return err
		}

		koreksi, err := s.repo.TotalKoreksiPeriodeTx(ctx, tx, periode.ID, p.ID)
		if err != nil {
			return err
		}

		kasbonDB, err := s.repo.ListKasbonAktifForUpdateTx(ctx, tx, p.ID)
		if err != nil {
			return err
		}

		hasil, err := HitungGajiPekerja(InputKalkulatorGaji{
			Pekerja:     p,
			Periode:     periode,
			RiwayatGaji: riwayat,
			Absensi:     absensi,
			Koreksi:     koreksi,
			KasbonAktif: konversiKasbonAktif(kasbonDB),
		})
		if err != nil {
			return ErrPrasyaratPenggajian
		}

		gajiID, err := s.repo.SimpanGaji(ctx, tx, periode.ID, p.ID, repository.HasilSimpanGaji{
			GajiPokok:               hasil.GajiPokok,
			JumlahHariHadir:         hasil.JumlahHariHadir,
			PotonganTerlambat:       hasil.PotonganTerlambat,
			UpahLembur:              hasil.UpahLembur,
			PenyesuaianProporsional: hasil.PenyesuaianProporsional,
			GajiKotor:               hasil.GajiKotor,
			PotonganKasbon:          hasil.PotonganKasbon,
			GajiBersih:              hasil.GajiBersih,
		})
		if err != nil {
			return err
		}

		for _, r := range hasil.RincianKasbon {
			if err := s.repo.CatatPotonganKasbon(ctx, tx, gajiID, r.KasbonID, r.NominalPotongan); err != nil {
				return err
			}
		}
	}

	tglBayar := tanggalPembayaran
	if tglBayar.IsZero() {
		if periode.TanggalPembayaran != nil && !periode.TanggalPembayaran.IsZero() {
			tglBayar = *periode.TanggalPembayaran
		} else {
			tglBayar = time.Now()
		}
	}

	if err := s.repo.TutupPeriodeTx(ctx, tx, periode.ID, tglBayar); err != nil {
		return err
	}

	return tx.Commit()
}

// Hasil mengambil daftar gaji yang sudah tersimpan untuk periode tertentu.
func (s *PenggajianService) Hasil(ctx context.Context, periodeID int64) (HasilPeriodeGaji, error) {
	if periodeID <= 0 {
		return HasilPeriodeGaji{}, ErrInputTidakValid
	}

	periode, err := s.repo.AmbilPeriodeByID(ctx, periodeID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return HasilPeriodeGaji{}, ErrTidakDitemukan
		}
		return HasilPeriodeGaji{}, err
	}

	items, err := s.repo.DaftarHasilGajiPeriode(ctx, periodeID)
	if err != nil {
		return HasilPeriodeGaji{}, err
	}

	totalKotor := decimal.Zero
	totalKasbon := decimal.Zero
	totalBersih := decimal.Zero
	for _, item := range items {
		totalKotor = totalKotor.Add(item.GajiKotor)
		totalKasbon = totalKasbon.Add(item.PotonganKasbon)
		totalBersih = totalBersih.Add(item.GajiBersih)
	}

	return HasilPeriodeGaji{
		Periode:             periode,
		Items:               items,
		TotalGajiKotor:      totalKotor,
		TotalPotonganKasbon: totalKasbon,
		TotalGajiBersih:     totalBersih,
	}, nil
}

func kategoriUntukPeriode(jenisPeriode string) string {
	if jenisPeriode == model.PeriodeBulanan {
		return model.KategoriBulanan
	}
	return model.KategoriHarian
}

func konversiKasbonAktif(items []repository.KasbonAktifPenggajian) []ItemKasbonAktif {
	hasil := make([]ItemKasbonAktif, 0, len(items))
	for _, item := range items {
		hasil = append(hasil, ItemKasbonAktif{
			ID:               item.ID,
			TanggalPengajuan: item.TanggalPengajuan,
			Sisa:             item.Sisa,
		})
	}
	return hasil
}
