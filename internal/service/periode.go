package service

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/shopspring/decimal"

	"penggajian/internal/model"
	"penggajian/internal/repository"
)

var ErrPeriodeDuplikat = errors.New("periode gaji dengan jenis dan rentang tanggal tersebut sudah ada")

// Pemilik: Anggota 3 | Periode Gaji (BR-01, BR-02, BR-10)
type PeriodeService struct {
	repo *repository.PeriodeRepository
}

func NewPeriodeService(repo *repository.PeriodeRepository) *PeriodeService {
	return &PeriodeService{repo: repo}
}

type DetailPeriode struct {
	Periode              model.PeriodeGaji
	Hari                 []repository.StatusHariPeriode
	SemuaHariTervalidasi bool
	JumlahPekerja        int
	TotalGajiBersih      decimal.Decimal
}

func (s *PeriodeService) Daftar(ctx context.Context) ([]model.PeriodeGaji, error) {
	return s.repo.Daftar(ctx)
}

func (s *PeriodeService) Buat(ctx context.Context, periode model.PeriodeGaji) (int64, error) {
	periode, err := validasiDanNormalisasiPeriode(periode)
	if err != nil {
		return 0, err
	}

	id, err := s.repo.Buat(ctx, periode)
	if errors.Is(err, repository.ErrPeriodeDuplikat) {
		return 0, ErrPeriodeDuplikat
	}

	return id, err
}

func validasiDanNormalisasiPeriode(
	periode model.PeriodeGaji,
) (model.PeriodeGaji, error) {

	if periode.JenisPeriode != model.PeriodeMingguan &&
		periode.JenisPeriode != model.PeriodeBulanan {
		return model.PeriodeGaji{}, ErrInputTidakValid
	}

	if periode.TanggalMulai.IsZero() || periode.TanggalSelesai.IsZero() {
		return model.PeriodeGaji{}, ErrInputTidakValid
	}

	mulai := normTanggal(periode.TanggalMulai)
	selesai := normTanggal(periode.TanggalSelesai)

	if selesai.Before(mulai) {
		return model.PeriodeGaji{}, ErrInputTidakValid
	}

	if periode.JenisPeriode == model.PeriodeMingguan {
		// BR-01: Senin sampai Sabtu (6 hari kalender).
		if mulai.Weekday() != time.Monday ||
			selesai.Weekday() != time.Saturday {
			return model.PeriodeGaji{}, ErrInputTidakValid
		}

		if !mulai.AddDate(0, 0, 5).Equal(selesai) {
			return model.PeriodeGaji{}, ErrInputTidakValid
		}
	}

	if periode.JenisPeriode == model.PeriodeBulanan {
		// BR-02: Tanggal 1 sampai akhir bulan yang sama.
		if mulai.Day() != 1 {
			return model.PeriodeGaji{}, ErrInputTidakValid
		}

		hariTerakhir := time.Date(
			mulai.Year(),
			mulai.Month()+1,
			0,
			0, 0, 0, 0,
			time.UTC,
		)

		if !selesai.Equal(hariTerakhir) {
			return model.PeriodeGaji{}, ErrInputTidakValid
		}
	}

	var tanggalPembayaran *time.Time

	if periode.TanggalPembayaran != nil &&
		!periode.TanggalPembayaran.IsZero() {

		tglBayar := normTanggal(*periode.TanggalPembayaran)

		if tglBayar.Before(selesai) {
			return model.PeriodeGaji{}, ErrInputTidakValid
		}

		tanggalPembayaran = &tglBayar
	}

	periode.TanggalMulai = mulai
	periode.TanggalSelesai = selesai
	periode.TanggalPembayaran = tanggalPembayaran
	periode.Status = model.PeriodeTerbuka

	return periode, nil
}

func (s *PeriodeService) AmbilByID(ctx context.Context, id int64) (model.PeriodeGaji, error) {
	if id <= 0 {
		return model.PeriodeGaji{}, ErrInputTidakValid
	}

	p, err := s.repo.AmbilByID(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return model.PeriodeGaji{}, ErrTidakDitemukan
	}
	return p, err
}

func (s *PeriodeService) Detail(ctx context.Context, id int64) (DetailPeriode, error) {
	periode, err := s.AmbilByID(ctx, id)
	if err != nil {
		return DetailPeriode{}, err
	}

	hari, err := s.repo.DaftarStatusHari(ctx, periode.TanggalMulai, periode.TanggalSelesai)
	if err != nil {
		return DetailPeriode{}, err
	}

	semuaTervalidasi := len(hari) > 0
	for _, h := range hari {
		if !h.SudahDivalidasi || h.JumlahMenunggu > 0 {
			semuaTervalidasi = false
			break
		}
	}

	jumlahPekerja, totalBersih, err := s.repo.RingkasanGajiPeriode(ctx, periode.ID)
	if err != nil {
		return DetailPeriode{}, err
	}

	return DetailPeriode{
		Periode:              periode,
		Hari:                 hari,
		SemuaHariTervalidasi: semuaTervalidasi,
		JumlahPekerja:        jumlahPekerja,
		TotalGajiBersih:      totalBersih,
	}, nil
}

func normTanggal(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}
