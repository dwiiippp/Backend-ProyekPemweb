package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/shopspring/decimal"
)

// Pemilik: Anggota 3 | F-10 · UC-14 Koreksi Gaji (BR-11)
// Tabel: koreksi_gaji (tulis)
// Semua query SQL modul ini ditulis di sini (pakai placeholder ?, jangan concat string).
type KoreksiRepository struct {
	db *sql.DB
}

func NewKoreksiRepository(db *sql.DB) *KoreksiRepository {
	return &KoreksiRepository{db: db}
}

func (r *KoreksiRepository) BeginTx(ctx context.Context) (*sql.Tx, error) {
	return r.db.BeginTx(ctx, nil)
}

type KoreksiRecord struct {
	ID                  int64
	PeriodeTujuanID     int64
	JenisPeriodeTujuan  string
	MulaiTujuan         time.Time
	SelesaiTujuan       time.Time
	StatusPeriodeTujuan string
	GajiSumberID        int64
	MulaiSumber         time.Time
	SelesaiSumber       time.Time
	PekerjaID           int64
	NamaPekerja         string
	NIK                 string
	JenisKoreksi        string
	Nominal             decimal.Decimal
	Alasan              string
	DibuatPada          time.Time
}

type OpsiGajiSumber struct {
	GajiID         int64
	PekerjaID      int64
	NamaPekerja    string
	NIK            string
	KategoriKerja  string
	PeriodeID      int64
	JenisPeriode   string
	TanggalMulai   time.Time
	TanggalSelesai time.Time
	GajiBersih     decimal.Decimal
}

type OpsiPeriodeTujuan struct {
	ID             int64
	JenisPeriode   string
	TanggalMulai   time.Time
	TanggalSelesai time.Time
}

type InfoGajiSumber struct {
	GajiID        int64
	PekerjaID     int64
	KategoriKerja string
	StatusPeriode string
}

type InfoPeriodeTujuan struct {
	ID           int64
	JenisPeriode string
	Status       string
}

func (r *KoreksiRepository) Daftar(ctx context.Context) ([]KoreksiRecord, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT
			k.id,
			k.periode_gaji_tujuan_id,
			pt.jenis_periode,
			pt.tanggal_mulai,
			pt.tanggal_selesai,
			pt.status,
			k.gaji_sumber_id,
			ps.tanggal_mulai,
			ps.tanggal_selesai,
			k.pekerja_id,
			u.nama,
			p.nik,
			k.jenis_koreksi,
			k.nominal,
			k.alasan,
			k.created_at
		FROM koreksi_gaji k
		JOIN periode_gaji pt ON pt.id = k.periode_gaji_tujuan_id
		JOIN gaji gs ON gs.id = k.gaji_sumber_id
		JOIN periode_gaji ps ON ps.id = gs.periode_gaji_id
		JOIN pekerja p ON p.id = k.pekerja_id AND p.deleted_at IS NULL
		JOIN users u ON u.id = p.user_id AND u.deleted_at IS NULL
		ORDER BY k.created_at DESC, k.id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var hasil []KoreksiRecord
	for rows.Next() {
		var item KoreksiRecord
		var alasan sql.NullString
		if err := rows.Scan(
			&item.ID,
			&item.PeriodeTujuanID,
			&item.JenisPeriodeTujuan,
			&item.MulaiTujuan,
			&item.SelesaiTujuan,
			&item.StatusPeriodeTujuan,
			&item.GajiSumberID,
			&item.MulaiSumber,
			&item.SelesaiSumber,
			&item.PekerjaID,
			&item.NamaPekerja,
			&item.NIK,
			&item.JenisKoreksi,
			&item.Nominal,
			&alasan,
			&item.DibuatPada,
		); err != nil {
			return nil, err
		}
		if alasan.Valid {
			item.Alasan = alasan.String
		}
		hasil = append(hasil, item)
	}
	return hasil, rows.Err()
}

func (r *KoreksiRepository) DaftarOpsiGajiSumber(ctx context.Context) ([]OpsiGajiSumber, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT
			g.id, g.pekerja_id, u.nama, p.nik, p.kategori_kerja,
			pg.id, pg.jenis_periode, pg.tanggal_mulai, pg.tanggal_selesai, g.gaji_bersih
		FROM gaji g
		JOIN periode_gaji pg ON pg.id = g.periode_gaji_id
		JOIN pekerja p ON p.id = g.pekerja_id AND p.deleted_at IS NULL
		JOIN users u ON u.id = p.user_id AND u.deleted_at IS NULL
		WHERE pg.status = 'ditutup'
		ORDER BY pg.tanggal_mulai DESC, u.nama ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var hasil []OpsiGajiSumber
	for rows.Next() {
		var item OpsiGajiSumber
		if err := rows.Scan(
			&item.GajiID, &item.PekerjaID, &item.NamaPekerja, &item.NIK, &item.KategoriKerja,
			&item.PeriodeID, &item.JenisPeriode, &item.TanggalMulai, &item.TanggalSelesai, &item.GajiBersih,
		); err != nil {
			return nil, err
		}
		hasil = append(hasil, item)
	}
	return hasil, rows.Err()
}

func (r *KoreksiRepository) DaftarPeriodeTerbuka(ctx context.Context) ([]OpsiPeriodeTujuan, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, jenis_periode, tanggal_mulai, tanggal_selesai
		FROM periode_gaji
		WHERE status = 'terbuka'
		ORDER BY tanggal_mulai ASC, id ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var hasil []OpsiPeriodeTujuan
	for rows.Next() {
		var item OpsiPeriodeTujuan
		if err := rows.Scan(&item.ID, &item.JenisPeriode, &item.TanggalMulai, &item.TanggalSelesai); err != nil {
			return nil, err
		}
		hasil = append(hasil, item)
	}
	return hasil, rows.Err()
}

func (r *KoreksiRepository) AmbilInfoGajiSumberTx(ctx context.Context, tx *sql.Tx, gajiID int64) (InfoGajiSumber, error) {
	var info InfoGajiSumber

	err := tx.QueryRowContext(ctx, `
        SELECT
            g.id,
            g.pekerja_id,
            p.kategori_kerja,
            pg.status
        FROM gaji g
        JOIN pekerja p ON p.id = g.pekerja_id
            AND p.deleted_at IS NULL
        JOIN periode_gaji pg ON pg.id = g.periode_gaji_id
        WHERE g.id = ?`,
		gajiID,
	).Scan(
		&info.GajiID,
		&info.PekerjaID,
		&info.KategoriKerja,
		&info.StatusPeriode,
	)

	return info, err
}

func (r *KoreksiRepository) AmbilInfoPeriodeTujuanForUpdateTx(ctx context.Context, tx *sql.Tx, periodeID int64) (InfoPeriodeTujuan, error) {
	var info InfoPeriodeTujuan

	err := tx.QueryRowContext(ctx, `
        SELECT id, jenis_periode, status
        FROM periode_gaji
        WHERE id = ?
        FOR UPDATE`,
		periodeID,
	).Scan(
		&info.ID,
		&info.JenisPeriode,
		&info.Status,
	)

	return info, err
}

func (r *KoreksiRepository) BuatTx(ctx context.Context, tx *sql.Tx, periodeTujuanID int64, gajiSumberID int64, pekerjaID int64, jenisKoreksi string, nominal decimal.Decimal, alasan string) (int64, error) {
	res, err := tx.ExecContext(ctx, `
        INSERT INTO koreksi_gaji (
            periode_gaji_tujuan_id,
            gaji_sumber_id,
            pekerja_id,
            jenis_koreksi,
            nominal,
            alasan
        ) VALUES (?, ?, ?, ?, ?, ?)`,
		periodeTujuanID,
		gajiSumberID,
		pekerjaID,
		jenisKoreksi,
		nominal,
		alasan,
	)
	if err != nil {
		return 0, err
	}

	return res.LastInsertId()
}
