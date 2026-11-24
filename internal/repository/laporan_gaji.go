package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/shopspring/decimal"
)

// Pemilik: Anggota 3 | F-12 Laporan status periode & rekap biaya gaji/lembur
// Tabel: periode_gaji, gaji (baca)
// Semua query SQL modul ini ditulis di sini (pakai placeholder ?, jangan concat string).
type LaporanGajiRepository struct {
	db *sql.DB
}

func NewLaporanGajiRepository(db *sql.DB) *LaporanGajiRepository {
	return &LaporanGajiRepository{db: db}
}

type BarisStatusPeriode struct {
	PeriodeID           int64
	JenisPeriode        string
	TanggalMulai        time.Time
	TanggalSelesai      time.Time
	TanggalPembayaran   time.Time
	Status              string
	TotalHariTerjadwal  int
	HariTervalidasi     int
	JumlahPekerjaDigaji int
	TotalGajiBersih     decimal.Decimal
}

type BarisRekapGaji struct {
	PeriodeID              int64
	JenisPeriode           string
	TanggalMulai           time.Time
	TanggalSelesai         time.Time
	TanggalPembayaran      time.Time
	JumlahPekerja          int
	TotalGajiPokok         decimal.Decimal
	TotalPotonganTerlambat decimal.Decimal
	TotalUpahLembur        decimal.Decimal
	TotalPenyesuaian       decimal.Decimal
	TotalGajiKotor         decimal.Decimal
	TotalPotonganKasbon    decimal.Decimal
	TotalGajiBersih        decimal.Decimal
}

func (r *LaporanGajiRepository) DaftarStatusPeriode(ctx context.Context, jenis, status string) ([]BarisStatusPeriode, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT
			pg.id,
			pg.jenis_periode,
			pg.tanggal_mulai,
			pg.tanggal_selesai,
			pg.tanggal_pembayaran,
			pg.status,
			(
				SELECT COUNT(DISTINCT j.tanggal_jadwal)
				FROM jadwal_kerja j
				WHERE j.tanggal_jadwal BETWEEN pg.tanggal_mulai AND pg.tanggal_selesai
			) AS total_hari_terjadwal,
			(
				SELECT COUNT(DISTINCT j.tanggal_jadwal)
				FROM jadwal_kerja j
				JOIN validasi_presensi_harian v ON v.tanggal_presensi = j.tanggal_jadwal
				WHERE j.tanggal_jadwal BETWEEN pg.tanggal_mulai AND pg.tanggal_selesai
			) AS hari_tervalidasi,
			COALESCE(g_agg.jumlah_pekerja, 0) AS jumlah_pekerja_digaji,
			COALESCE(g_agg.total_gaji_bersih, 0) AS total_gaji_bersih
		FROM periode_gaji pg
		LEFT JOIN (
			SELECT periode_gaji_id, COUNT(*) AS jumlah_pekerja, SUM(gaji_bersih) AS total_gaji_bersih
			FROM gaji
			GROUP BY periode_gaji_id
		) g_agg ON g_agg.periode_gaji_id = pg.id
		WHERE (? = '' OR pg.jenis_periode = ?)
		  AND (? = '' OR pg.status = ?)
		ORDER BY pg.tanggal_mulai DESC, pg.id DESC`,
		jenis, jenis, status, status,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var hasil []BarisStatusPeriode
	for rows.Next() {
		var item BarisStatusPeriode
		var tglBayar sql.NullTime
		if err := rows.Scan(
			&item.PeriodeID, &item.JenisPeriode, &item.TanggalMulai, &item.TanggalSelesai,
			&tglBayar, &item.Status, &item.TotalHariTerjadwal, &item.HariTervalidasi,
			&item.JumlahPekerjaDigaji, &item.TotalGajiBersih,
		); err != nil {
			return nil, err
		}
		if tglBayar.Valid {
			item.TanggalPembayaran = tglBayar.Time
		}
		hasil = append(hasil, item)
	}
	return hasil, rows.Err()
}

func (r *LaporanGajiRepository) RekapBiayaGaji(ctx context.Context, jenis, mulai, selesai string) ([]BarisRekapGaji, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT
			pg.id,
			pg.jenis_periode,
			pg.tanggal_mulai,
			pg.tanggal_selesai,
			pg.tanggal_pembayaran,
			COUNT(g.id) AS jumlah_pekerja,
			COALESCE(SUM(g.gaji_pokok), 0),
			COALESCE(SUM(g.potongan_terlambat), 0),
			COALESCE(SUM(g.upah_lembur), 0),
			COALESCE(SUM(g.penyesuaian_proporsional), 0),
			COALESCE(SUM(g.gaji_kotor), 0),
			COALESCE(SUM(g.potongan_kasbon), 0),
			COALESCE(SUM(g.gaji_bersih), 0)
		FROM periode_gaji pg
		LEFT JOIN gaji g ON g.periode_gaji_id = pg.id
		WHERE pg.status = 'ditutup'
		  AND (? = '' OR pg.jenis_periode = ?)
		  AND (? = '' OR pg.tanggal_selesai >= ?)
		  AND (? = '' OR pg.tanggal_mulai <= ?)
		GROUP BY pg.id, pg.jenis_periode, pg.tanggal_mulai, pg.tanggal_selesai, pg.tanggal_pembayaran
		ORDER BY pg.tanggal_mulai DESC, pg.id DESC`,
		jenis, jenis, mulai, mulai, selesai, selesai,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var hasil []BarisRekapGaji
	for rows.Next() {
		var item BarisRekapGaji
		var tglBayar sql.NullTime
		if err := rows.Scan(
			&item.PeriodeID, &item.JenisPeriode, &item.TanggalMulai, &item.TanggalSelesai, &tglBayar,
			&item.JumlahPekerja, &item.TotalGajiPokok, &item.TotalPotonganTerlambat,
			&item.TotalUpahLembur, &item.TotalPenyesuaian, &item.TotalGajiKotor,
			&item.TotalPotonganKasbon, &item.TotalGajiBersih,
		); err != nil {
			return nil, err
		}
		if tglBayar.Valid {
			item.TanggalPembayaran = tglBayar.Time
		}
		hasil = append(hasil, item)
	}
	return hasil, rows.Err()
}
