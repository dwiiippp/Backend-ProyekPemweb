package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/shopspring/decimal"
)

// Pemilik: Anggota 3 | F-11 · UC-06 Lihat Slip · UC-15 Cetak Slip (BR-12)
// Tabel: gaji, potongan_kasbon_gaji (baca)
// Semua query SQL modul ini ditulis di sini (pakai placeholder ?, jangan concat string).
type SlipRepository struct {
	db *sql.DB
}

func NewSlipRepository(db *sql.DB) *SlipRepository {
	return &SlipRepository{db: db}
}

type RingkasanSlip struct {
	GajiID            int64
	PeriodeID         int64
	JenisPeriode      string
	TanggalMulai      time.Time
	TanggalSelesai    time.Time
	TanggalPembayaran time.Time
	PekerjaID         int64
	NamaPekerja       string
	NIK               string
	KategoriKerja     string
	GajiKotor         decimal.Decimal
	PotonganKasbon    decimal.Decimal
	GajiBersih        decimal.Decimal
}

type RincianPotonganKasbonSlip struct {
	KasbonID         int64
	TanggalPengajuan time.Time
	NominalKasbon    decimal.Decimal
	NominalPotongan  decimal.Decimal
	SisaKasbon       decimal.Decimal
	StatusKasbon     string
}

type RincianKoreksiSlip struct {
	ID           int64
	JenisKoreksi string
	Nominal      decimal.Decimal
	Alasan       string
}

type DetailSlip struct {
	GajiID                  int64
	PeriodeID               int64
	JenisPeriode            string
	TanggalMulai            time.Time
	TanggalSelesai          time.Time
	TanggalPembayaran       time.Time
	StatusPeriode           string
	PekerjaID               int64
	NamaPekerja             string
	NIK                     string
	KategoriKerja           string
	JumlahHariHadir         int
	GajiPokok               decimal.Decimal
	PotonganTerlambat       decimal.Decimal
	UpahLembur              decimal.Decimal
	PenyesuaianProporsional decimal.Decimal
	TotalKoreksi            decimal.Decimal
	GajiKotor               decimal.Decimal
	PotonganKasbon          decimal.Decimal
	GajiBersih              decimal.Decimal
	PotonganKasbonRincian   []RincianPotonganKasbonSlip
	KoreksiRincian          []RincianKoreksiSlip
}

type OpsiPeriodeSlip struct {
	ID             int64
	JenisPeriode   string
	TanggalMulai   time.Time
	TanggalSelesai time.Time
}

type OpsiPekerjaSlip struct {
	ID   int64
	Nama string
	NIK  string
}

func (r *SlipRepository) DaftarAdmin(ctx context.Context, periodeID, pekerjaID int64) ([]RingkasanSlip, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT
			g.id, pg.id, pg.jenis_periode, pg.tanggal_mulai, pg.tanggal_selesai,
			pg.tanggal_pembayaran, p.id, u.nama, p.nik, p.kategori_kerja,
			g.gaji_kotor, g.potongan_kasbon, g.gaji_bersih
		FROM gaji g
		JOIN periode_gaji pg ON pg.id = g.periode_gaji_id
		JOIN pekerja p ON p.id = g.pekerja_id AND p.deleted_at IS NULL
		JOIN users u ON u.id = p.user_id AND u.deleted_at IS NULL
		WHERE pg.status = 'ditutup'
		  AND (? = 0 OR g.periode_gaji_id = ?)
		  AND (? = 0 OR g.pekerja_id = ?)
		ORDER BY pg.tanggal_mulai DESC, u.nama ASC`,
		periodeID, periodeID, pekerjaID, pekerjaID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanRingkasanSlip(rows)
}

func (r *SlipRepository) DaftarMilikPekerja(ctx context.Context, pekerjaID int64) ([]RingkasanSlip, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT
			g.id, pg.id, pg.jenis_periode, pg.tanggal_mulai, pg.tanggal_selesai,
			pg.tanggal_pembayaran, p.id, u.nama, p.nik, p.kategori_kerja,
			g.gaji_kotor, g.potongan_kasbon, g.gaji_bersih
		FROM gaji g
		JOIN periode_gaji pg ON pg.id = g.periode_gaji_id
		JOIN pekerja p ON p.id = g.pekerja_id AND p.deleted_at IS NULL
		JOIN users u ON u.id = p.user_id AND u.deleted_at IS NULL
		WHERE pg.status = 'ditutup'
		  AND g.pekerja_id = ?
		ORDER BY pg.tanggal_mulai DESC, g.id DESC`,
		pekerjaID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanRingkasanSlip(rows)
}

func scanRingkasanSlip(rows *sql.Rows) ([]RingkasanSlip, error) {
	var hasil []RingkasanSlip
	for rows.Next() {
		var item RingkasanSlip
		var tglBayar sql.NullTime
		if err := rows.Scan(
			&item.GajiID, &item.PeriodeID, &item.JenisPeriode,
			&item.TanggalMulai, &item.TanggalSelesai, &tglBayar,
			&item.PekerjaID, &item.NamaPekerja, &item.NIK, &item.KategoriKerja,
			&item.GajiKotor, &item.PotonganKasbon, &item.GajiBersih,
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

func (r *SlipRepository) DaftarPeriodeDitutup(ctx context.Context) ([]OpsiPeriodeSlip, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, jenis_periode, tanggal_mulai, tanggal_selesai
		FROM periode_gaji
		WHERE status = 'ditutup'
		ORDER BY tanggal_mulai DESC, id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var hasil []OpsiPeriodeSlip
	for rows.Next() {
		var item OpsiPeriodeSlip
		if err := rows.Scan(&item.ID, &item.JenisPeriode, &item.TanggalMulai, &item.TanggalSelesai); err != nil {
			return nil, err
		}
		hasil = append(hasil, item)
	}
	return hasil, rows.Err()
}

func (r *SlipRepository) DaftarPekerja(ctx context.Context) ([]OpsiPekerjaSlip, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT p.id, u.nama, p.nik
		FROM pekerja p
		JOIN users u ON u.id = p.user_id
		WHERE p.deleted_at IS NULL
		  AND u.deleted_at IS NULL
		ORDER BY u.nama ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var hasil []OpsiPekerjaSlip
	for rows.Next() {
		var item OpsiPekerjaSlip
		if err := rows.Scan(&item.ID, &item.Nama, &item.NIK); err != nil {
			return nil, err
		}
		hasil = append(hasil, item)
	}
	return hasil, rows.Err()
}

func (r *SlipRepository) DetailByGajiID(ctx context.Context, gajiID int64) (DetailSlip, error) {
	var slip DetailSlip
	var tglBayar sql.NullTime

	err := r.db.QueryRowContext(ctx, `
		SELECT
			g.id, pg.id, pg.jenis_periode, pg.tanggal_mulai, pg.tanggal_selesai,
			pg.tanggal_pembayaran, pg.status,
			p.id, u.nama, p.nik, p.kategori_kerja,
			g.jumlah_hari_hadir, g.gaji_pokok, g.potongan_terlambat,
			g.upah_lembur, g.penyesuaian_proporsional,
			g.gaji_kotor, g.potongan_kasbon, g.gaji_bersih
		FROM gaji g
		JOIN periode_gaji pg ON pg.id = g.periode_gaji_id
		JOIN pekerja p ON p.id = g.pekerja_id AND p.deleted_at IS NULL
		JOIN users u ON u.id = p.user_id AND u.deleted_at IS NULL
		WHERE g.id = ?`,
		gajiID,
	).Scan(
		&slip.GajiID, &slip.PeriodeID, &slip.JenisPeriode,
		&slip.TanggalMulai, &slip.TanggalSelesai, &tglBayar, &slip.StatusPeriode,
		&slip.PekerjaID, &slip.NamaPekerja, &slip.NIK, &slip.KategoriKerja,
		&slip.JumlahHariHadir, &slip.GajiPokok, &slip.PotonganTerlambat,
		&slip.UpahLembur, &slip.PenyesuaianProporsional,
		&slip.GajiKotor, &slip.PotonganKasbon, &slip.GajiBersih,
	)
	if err != nil {
		return DetailSlip{}, err
	}
	if tglBayar.Valid {
		slip.TanggalPembayaran = tglBayar.Time
	}

	kasbonRows, err := r.db.QueryContext(ctx, `
		SELECT
			pk.kasbon_id,
			k.tanggal_pengajuan,
			k.nominal_kasbon,
			pk.nominal_potongan,
			COALESCE((
				SELECT SUM(pk2.nominal_potongan)
				FROM potongan_kasbon_gaji pk2
				JOIN gaji g2 ON g2.id = pk2.gaji_id
				JOIN periode_gaji pg2 ON pg2.id = g2.periode_gaji_id
				WHERE pk2.kasbon_id = pk.kasbon_id
				AND (
					pg2.tanggal_selesai < pg.tanggal_selesai
					OR (
						pg2.tanggal_selesai = pg.tanggal_selesai
						AND pg2.tanggal_mulai <= pg.tanggal_mulai
					)
				)
			), 0) AS total_potongan_historis
		FROM potongan_kasbon_gaji pk
		JOIN kasbon k ON k.id = pk.kasbon_id
		JOIN gaji g ON g.id = pk.gaji_id
		JOIN periode_gaji pg ON pg.id = g.periode_gaji_id
		WHERE pk.gaji_id = ?
		ORDER BY k.tanggal_pengajuan ASC, k.id ASC`,
		gajiID,
	)
	if err != nil {
		return DetailSlip{}, err
	}
	defer kasbonRows.Close()

	for kasbonRows.Next() {
		var k RincianPotonganKasbonSlip
		var totalPotonganHistoris decimal.Decimal

		if err := kasbonRows.Scan(
			&k.KasbonID,
			&k.TanggalPengajuan,
			&k.NominalKasbon,
			&k.NominalPotongan,
			&totalPotonganHistoris,
		); err != nil {
			return DetailSlip{}, err
		}

		k.SisaKasbon = k.NominalKasbon.
			Sub(totalPotonganHistoris).
			Round(2)

		if k.SisaKasbon.IsNegative() {
			k.SisaKasbon = decimal.Zero
		}

		if k.SisaKasbon.IsZero() {
			k.StatusKasbon = "lunas"
		} else {
			k.StatusKasbon = "aktif"
		}

		slip.PotonganKasbonRincian = append(
			slip.PotonganKasbonRincian,
			k,
		)
	}
	if err := kasbonRows.Err(); err != nil {
		return DetailSlip{}, err
	}

	koreksiRows, err := r.db.QueryContext(ctx, `
		SELECT id, jenis_koreksi, nominal, alasan
		FROM koreksi_gaji
		WHERE periode_gaji_tujuan_id = ?
		  AND pekerja_id = ?
		ORDER BY id ASC`,
		slip.PeriodeID, slip.PekerjaID,
	)
	if err != nil {
		return DetailSlip{}, err
	}
	defer koreksiRows.Close()

	totalKoreksi := decimal.Zero
	for koreksiRows.Next() {
		var kr RincianKoreksiSlip
		var alasan sql.NullString
		if err := koreksiRows.Scan(&kr.ID, &kr.JenisKoreksi, &kr.Nominal, &alasan); err != nil {
			return DetailSlip{}, err
		}
		if alasan.Valid {
			kr.Alasan = alasan.String
		}
		if kr.JenisKoreksi == "tambah" {
			totalKoreksi = totalKoreksi.Add(kr.Nominal)
		} else {
			totalKoreksi = totalKoreksi.Sub(kr.Nominal)
		}
		slip.KoreksiRincian = append(slip.KoreksiRincian, kr)
	}
	if err := koreksiRows.Err(); err != nil {
		return DetailSlip{}, err
	}
	slip.TotalKoreksi = totalKoreksi

	return slip, nil
}
