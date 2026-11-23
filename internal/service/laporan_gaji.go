package service

import (
	"context"
	"strings"
	"time"

	"github.com/shopspring/decimal"

	"penggajian/internal/model"
	"penggajian/internal/repository"
)

// Pemilik: Anggota 3 | F-12 Laporan status periode & rekap biaya gaji/lembur
// LaporanGajiService menampung logika bisnis modul ini. Hanya baca data.
type LaporanGajiService struct {
	repo *repository.LaporanGajiRepository
}

func NewLaporanGajiService(repo *repository.LaporanGajiRepository) *LaporanGajiService {
	return &LaporanGajiService{repo: repo}
}

type RingkasanStatusPeriode struct {
	Items         []repository.BarisStatusPeriode
	JumlahTerbuka int
	JumlahDitutup int
}

type RingkasanRekapGaji struct {
	Items                  []repository.BarisRekapGaji
	TotalGajiPokok         decimal.Decimal
	TotalPotonganTerlambat decimal.Decimal
	TotalUpahLembur        decimal.Decimal
	TotalPenyesuaian       decimal.Decimal
	TotalGajiKotor         decimal.Decimal
	TotalPotonganKasbon    decimal.Decimal
	TotalGajiBersih        decimal.Decimal
}

func (s *LaporanGajiService) StatusPeriode(ctx context.Context, jenis, status string) (RingkasanStatusPeriode, error) {
	jenis = strings.TrimSpace(jenis)
	status = strings.TrimSpace(status)

	if jenis != "" && jenis != model.PeriodeMingguan && jenis != model.PeriodeBulanan {
		return RingkasanStatusPeriode{}, ErrInputTidakValid
	}
	if status != "" && status != model.PeriodeTerbuka && status != model.PeriodeDitutup {
		return RingkasanStatusPeriode{}, ErrInputTidakValid
	}

	items, err := s.repo.DaftarStatusPeriode(ctx, jenis, status)
	if err != nil {
		return RingkasanStatusPeriode{}, err
	}

	var terbuka, ditutup int
	for _, item := range items {
		if item.Status == model.PeriodeTerbuka {
			terbuka++
		} else if item.Status == model.PeriodeDitutup {
			ditutup++
		}
	}

	return RingkasanStatusPeriode{
		Items:         items,
		JumlahTerbuka: terbuka,
		JumlahDitutup: ditutup,
	}, nil
}

func (s *LaporanGajiService) RekapGaji(ctx context.Context, jenis, mulaiText, selesaiText string) (RingkasanRekapGaji, error) {
	jenis = strings.TrimSpace(jenis)
	mulaiText = strings.TrimSpace(mulaiText)
	selesaiText = strings.TrimSpace(selesaiText)

	if jenis != "" && jenis != model.PeriodeMingguan && jenis != model.PeriodeBulanan {
		return RingkasanRekapGaji{}, ErrInputTidakValid
	}

	var mulai, selesai time.Time
	var err error
	if mulaiText != "" {
		mulai, err = time.Parse("2006-01-02", mulaiText)
		if err != nil {
			return RingkasanRekapGaji{}, ErrInputTidakValid
		}
	}
	if selesaiText != "" {
		selesai, err = time.Parse("2006-01-02", selesaiText)
		if err != nil {
			return RingkasanRekapGaji{}, ErrInputTidakValid
		}
	}
	if !mulai.IsZero() && !selesai.IsZero() && selesai.Before(mulai) {
		return RingkasanRekapGaji{}, ErrInputTidakValid
	}

	items, err := s.repo.RekapBiayaGaji(ctx, jenis, mulaiText, selesaiText)
	if err != nil {
		return RingkasanRekapGaji{}, err
	}

	out := RingkasanRekapGaji{Items: items}
	for _, item := range items {
		out.TotalGajiPokok = out.TotalGajiPokok.Add(item.TotalGajiPokok)
		out.TotalPotonganTerlambat = out.TotalPotonganTerlambat.Add(item.TotalPotonganTerlambat)
		out.TotalUpahLembur = out.TotalUpahLembur.Add(item.TotalUpahLembur)
		out.TotalPenyesuaian = out.TotalPenyesuaian.Add(item.TotalPenyesuaian)
		out.TotalGajiKotor = out.TotalGajiKotor.Add(item.TotalGajiKotor)
		out.TotalPotonganKasbon = out.TotalPotonganKasbon.Add(item.TotalPotonganKasbon)
		out.TotalGajiBersih = out.TotalGajiBersih.Add(item.TotalGajiBersih)
	}
	return out, nil
}
