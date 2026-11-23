package service

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"unicode/utf8"

	"github.com/shopspring/decimal"

	"penggajian/internal/model"
	"penggajian/internal/repository"
)

// Pemilik: Anggota 3 | F-10 · UC-14 Koreksi Gaji (BR-11)
// KoreksiService menampung logika bisnis modul ini. Koreksi dihitung langsung dari koreksi_gaji per periode_gaji_tujuan_id agar tidak terhitung dua kali.
type KoreksiService struct {
	repo *repository.KoreksiRepository
}

func NewKoreksiService(repo *repository.KoreksiRepository) *KoreksiService {
	return &KoreksiService{repo: repo}
}

type InputKoreksi struct {
	GajiSumberID    int64
	PeriodeTujuanID int64
	JenisKoreksi    string
	NominalText     string
	Alasan          string
}

type DataFormKoreksi struct {
	OpsiGajiSumber []repository.OpsiGajiSumber
	PeriodeTerbuka []repository.OpsiPeriodeTujuan
}

func (s *KoreksiService) Daftar(ctx context.Context) ([]repository.KoreksiRecord, error) {
	return s.repo.Daftar(ctx)
}

func (s *KoreksiService) DataForm(ctx context.Context) (DataFormKoreksi, error) {
	sumber, err := s.repo.DaftarOpsiGajiSumber(ctx)
	if err != nil {
		return DataFormKoreksi{}, err
	}
	tujuan, err := s.repo.DaftarPeriodeTerbuka(ctx)
	if err != nil {
		return DataFormKoreksi{}, err
	}
	return DataFormKoreksi{
		OpsiGajiSumber: sumber,
		PeriodeTerbuka: tujuan,
	}, nil
}

// ValidasiDasarKoreksi memvalidasi input form koreksi gaji secara murni.
func ValidasiDasarKoreksi(input InputKoreksi) (decimal.Decimal, string, error) {
	if input.GajiSumberID <= 0 || input.PeriodeTujuanID <= 0 {
		return decimal.Zero, "", ErrInputTidakValid
	}
	if input.JenisKoreksi != model.KoreksiTambah && input.JenisKoreksi != model.KoreksiKurang {
		return decimal.Zero, "", ErrInputTidakValid
	}

	nominal, err := decimal.NewFromString(strings.TrimSpace(input.NominalText))
	if err != nil || !nominal.IsPositive() || !nominal.Equal(nominal.Round(2)) {
		return decimal.Zero, "", ErrInputTidakValid
	}

	alasan := strings.TrimSpace(input.Alasan)
	if alasan == "" || utf8.RuneCountInString(alasan) > 2000 {
		return decimal.Zero, "", ErrInputTidakValid
	}

	return nominal, alasan, nil
}

// CocokKategoriPeriode memastikan koreksi diarahkan ke jenis periode yang sesuai kategori pekerja.
func CocokKategoriPeriode(kategoriKerja, jenisPeriode string) bool {
	if kategoriKerja == model.KategoriHarian {
		return jenisPeriode == model.PeriodeMingguan
	}
	if kategoriKerja == model.KategoriBulanan {
		return jenisPeriode == model.PeriodeBulanan
	}
	return false
}

func (s *KoreksiService) Buat(ctx context.Context, input InputKoreksi) (int64, error) {
	nominal, alasan, err := ValidasiDasarKoreksi(input)
	if err != nil {
		return 0, err
	}

	sumber, err := s.repo.AmbilInfoGajiSumber(ctx, input.GajiSumberID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, ErrTidakDitemukan
		}
		return 0, err
	}
	if sumber.StatusPeriode != model.PeriodeDitutup {
		return 0, ErrInputTidakValid
	}

	tujuan, err := s.repo.AmbilInfoPeriodeTujuan(ctx, input.PeriodeTujuanID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, ErrTidakDitemukan
		}
		return 0, err
	}
	if tujuan.Status != model.PeriodeTerbuka {
		return 0, ErrPeriodeDitutup
	}
	if !CocokKategoriPeriode(sumber.KategoriKerja, tujuan.JenisPeriode) {
		return 0, ErrInputTidakValid
	}

	return s.repo.Buat(
		ctx,
		tujuan.ID,
		sumber.GajiID,
		sumber.PekerjaID,
		input.JenisKoreksi,
		nominal,
		alasan,
	)
}
