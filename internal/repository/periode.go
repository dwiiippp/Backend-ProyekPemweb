package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/shopspring/decimal"

	"penggajian/internal/model"
)

var ErrPeriodeDuplikat = errors.New("periode gaji sudah terdaftar")

type PeriodeRepository struct {
	db *sql.DB
}

func NewPeriodeRepository(db *sql.DB) *PeriodeRepository {
	return &PeriodeRepository{db: db}
}

type StatusHariPeriode struct {
	Tanggal         time.Time
	JumlahJadwal    int
	JumlahDisetujui int
	JumlahMenunggu  int
	JumlahDitolak   int
	SudahDivalidasi bool
	DivalidasiPada  time.Time
}

func (r *PeriodeRepository) Daftar(ctx context.Context) ([]model.PeriodeGaji, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, jenis_periode, tanggal_mulai, tanggal_selesai,
		       tanggal_pembayaran, status, created_at
		FROM periode_gaji
		ORDER BY tanggal_mulai DESC, id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var hasil []model.PeriodeGaji

	for rows.Next() {
		var item model.PeriodeGaji
		var tanggalPembayaran sql.NullTime

		if err := rows.Scan(
			&item.ID,
			&item.JenisPeriode,
			&item.TanggalMulai,
			&item.TanggalSelesai,
			&tanggalPembayaran,
			&item.Status,
			&item.DibuatPada,
		); err != nil {
			return nil, err
		}

		if tanggalPembayaran.Valid {
			t := tanggalPembayaran.Time
			item.TanggalPembayaran = &t
		}

		hasil = append(hasil, item)
	}

	return hasil, rows.Err()
}

func (r *PeriodeRepository) Buat(ctx context.Context, periode model.PeriodeGaji) (int64, error) {
	var tanggalPembayaran sql.NullString
	if periode.TanggalPembayaran != nil && !periode.TanggalPembayaran.IsZero() {
		tanggalPembayaran = sql.NullString{
			String: periode.TanggalPembayaran.Format("2006-01-02"),
			Valid:  true,
		}
	}

	result, err := r.db.ExecContext(ctx, `
		INSERT INTO periode_gaji (jenis_periode, tanggal_mulai, tanggal_selesai, tanggal_pembayaran)
		VALUES (?, ?, ?, ?)`,
		periode.JenisPeriode,
		periode.TanggalMulai.Format("2006-01-02"),
		periode.TanggalSelesai.Format("2006-01-02"),
		tanggalPembayaran,
	)
	if err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			return 0, ErrPeriodeDuplikat
		}
		return 0, err
	}

	return result.LastInsertId()
}

func (r *PeriodeRepository) AmbilByID(ctx context.Context, id int64) (model.PeriodeGaji, error) {
	var item model.PeriodeGaji
	var tanggalPembayaran sql.NullTime

	err := r.db.QueryRowContext(ctx, `
		SELECT id, jenis_periode, tanggal_mulai, tanggal_selesai,
		       tanggal_pembayaran, status, created_at
		FROM periode_gaji
		WHERE id = ?`, id,
	).Scan(
		&item.ID,
		&item.JenisPeriode,
		&item.TanggalMulai,
		&item.TanggalSelesai,
		&tanggalPembayaran,
		&item.Status,
		&item.DibuatPada,
	)
	if err != nil {
		return model.PeriodeGaji{}, err
	}

	if tanggalPembayaran.Valid {
		t := tanggalPembayaran.Time
		item.TanggalPembayaran = &t
	}

	return item, nil
}

func (r *PeriodeRepository) DaftarStatusHari(ctx context.Context, mulai, selesai time.Time) ([]StatusHariPeriode, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT
			j.tanggal_jadwal,
			COUNT(DISTINCT j.id) AS jumlah_jadwal,
			COALESCE(SUM(CASE WHEN a.status_validasi = 'disetujui' THEN 1 ELSE 0 END), 0) AS jumlah_disetujui,
			COALESCE(SUM(CASE WHEN a.status_validasi = 'menunggu' THEN 1 ELSE 0 END), 0) AS jumlah_menunggu,
			COALESCE(SUM(CASE WHEN a.status_validasi = 'ditolak' THEN 1 ELSE 0 END), 0) AS jumlah_ditolak,
			CASE WHEN v.id IS NOT NULL THEN 1 ELSE 0 END AS sudah_divalidasi,
			v.divalidasi_pada
		FROM jadwal_kerja j
		LEFT JOIN absensi a ON a.jadwal_kerja_id = j.id
		LEFT JOIN validasi_presensi_harian v ON v.tanggal_presensi = j.tanggal_jadwal
		WHERE j.tanggal_jadwal BETWEEN ? AND ?
		GROUP BY j.tanggal_jadwal, v.id, v.divalidasi_pada
		ORDER BY j.tanggal_jadwal ASC`,
		mulai.Format("2006-01-02"),
		selesai.Format("2006-01-02"),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var hasil []StatusHariPeriode
	for rows.Next() {
		var item StatusHariPeriode
		var sudahInt int
		var divalidasiPada sql.NullTime
		if err := rows.Scan(
			&item.Tanggal,
			&item.JumlahJadwal,
			&item.JumlahDisetujui,
			&item.JumlahMenunggu,
			&item.JumlahDitolak,
			&sudahInt,
			&divalidasiPada,
		); err != nil {
			return nil, err
		}
		item.SudahDivalidasi = sudahInt == 1
		if divalidasiPada.Valid {
			item.DivalidasiPada = divalidasiPada.Time
		}
		hasil = append(hasil, item)
	}
	return hasil, rows.Err()
}

func (r *PeriodeRepository) RingkasanGajiPeriode(ctx context.Context, periodeID int64) (int, decimal.Decimal, error) {
	var jumlahPekerja int
	var totalBersih decimal.Decimal
	err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*), COALESCE(SUM(gaji_bersih), 0)
		FROM gaji
		WHERE periode_gaji_id = ?`,
		periodeID,
	).Scan(&jumlahPekerja, &totalBersih)
	return jumlahPekerja, totalBersih, err
}

func (r *PeriodeRepository) Tutup(ctx context.Context, id int64, tanggalPembayaran time.Time) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE periode_gaji
		SET status = ?, tanggal_pembayaran = ?
		WHERE id = ? AND status = ?`,
		model.PeriodeDitutup,
		tanggalPembayaran.Format("2006-01-02"),
		id,
		model.PeriodeTerbuka,
	)
	if err != nil {
		return err
	}

	jumlah, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if jumlah == 0 {
		return sql.ErrNoRows
	}

	return nil
}
