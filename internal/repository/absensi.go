package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/shopspring/decimal"
)

type AbsensiRepository struct {
	db *sql.DB
}

func NewAbsensiRepository(db *sql.DB) *AbsensiRepository {
	return &AbsensiRepository{db: db}
}

func adaPeriodeDitutup(ctx context.Context, tx *sql.Tx, tanggal string) (bool, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT status FROM periode_gaji
		WHERE ? BETWEEN tanggal_mulai AND tanggal_selesai
		FOR UPDATE`, tanggal)
	if err != nil {
		return false, err
	}
	defer rows.Close()
	for rows.Next() {
		var status string
		if err := rows.Scan(&status); err != nil {
			return false, err
		}
		if status == "ditutup" {
			return true, nil
		}
	}
	return false, rows.Err()
}

type JadwalPresensi struct {
	ID             int64
	Tanggal        time.Time
	NamaShift      string
	JamMasukShift  string
	JamKeluarShift string
}

type RiwayatPresensi struct {
	ID              int64
	Tanggal         time.Time
	NamaShift       string
	JamMasukAktual  time.Time
	JamKeluarAktual time.Time
	MenitTerlambat  sql.NullInt64
	JamLembur       decimal.NullDecimal
	Status          string
	Catatan         sql.NullString
}

var (
	ErrJadwalTidakDitemukan = errors.New("jadwal tidak ditemukan atau bukan milik pekerja")
	ErrPresensiGanda        = errors.New("presensi sudah ada")
	ErrHariDitutup          = errors.New("hari sudah ditutup")
	ErrPeriodeDitutup       = errors.New("periode sudah ditutup")
)

func (r *AbsensiRepository) ListJadwalTerbuka(ctx context.Context, pekerjaID int64, sampai time.Time) ([]JadwalPresensi, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT j.id, j.tanggal_jadwal, s.nama_shift, s.jam_masuk, s.jam_keluar
		FROM jadwal_kerja j
		JOIN shift s ON s.id = j.shift_id AND s.deleted_at IS NULL
		JOIN pekerja p ON p.id = j.pekerja_id AND p.status = 'aktif' AND p.deleted_at IS NULL
		JOIN users u ON u.id = p.user_id AND u.deleted_at IS NULL
		WHERE j.pekerja_id = ? AND j.tanggal_jadwal <= ?
		  AND NOT EXISTS (
			SELECT 1 FROM absensi a WHERE a.jadwal_kerja_id = j.id
		  )
		  AND NOT EXISTS (
			SELECT 1 FROM validasi_presensi_harian v WHERE v.tanggal_presensi = j.tanggal_jadwal
		  )
		  AND NOT EXISTS (
			SELECT 1 FROM periode_gaji pg
			WHERE pg.status = 'ditutup'
			  AND j.tanggal_jadwal BETWEEN pg.tanggal_mulai AND pg.tanggal_selesai
		  )
		ORDER BY j.tanggal_jadwal DESC, s.jam_masuk`, pekerjaID, sampai.Format("2006-01-02"))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []JadwalPresensi
	for rows.Next() {
		var item JadwalPresensi
		if err := rows.Scan(&item.ID, &item.Tanggal, &item.NamaShift, &item.JamMasukShift, &item.JamKeluarShift); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

func (r *AbsensiRepository) Jadwal(ctx context.Context, pekerjaID, jadwalID int64) (JadwalPresensi, error) {
	var item JadwalPresensi
	err := r.db.QueryRowContext(ctx, `
		SELECT j.id, j.tanggal_jadwal, s.nama_shift, s.jam_masuk, s.jam_keluar
		FROM jadwal_kerja j
		JOIN shift s ON s.id = j.shift_id AND s.deleted_at IS NULL
		JOIN pekerja p ON p.id = j.pekerja_id AND p.status = 'aktif' AND p.deleted_at IS NULL
		JOIN users u ON u.id = p.user_id AND u.deleted_at IS NULL
		WHERE j.id = ? AND j.pekerja_id = ?`, jadwalID, pekerjaID).
		Scan(&item.ID, &item.Tanggal, &item.NamaShift, &item.JamMasukShift, &item.JamKeluarShift)
	if errors.Is(err, sql.ErrNoRows) {
		return JadwalPresensi{}, ErrJadwalTidakDitemukan
	}
	return item, err
}

func (r *AbsensiRepository) Simpan(ctx context.Context, pekerjaID, jadwalID int64, masuk, keluar time.Time) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var tanggal time.Time
	err = tx.QueryRowContext(ctx, `
		SELECT tanggal_jadwal FROM jadwal_kerja
		WHERE id = ? AND pekerja_id = ?
		FOR UPDATE`, jadwalID, pekerjaID).Scan(&tanggal)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrJadwalTidakDitemukan
	}
	if err != nil {
		return err
	}

	var exists int
	err = tx.QueryRowContext(ctx, `
		SELECT EXISTS(SELECT 1 FROM validasi_presensi_harian WHERE tanggal_presensi = ?)`,
		tanggal.Format("2006-01-02")).Scan(&exists)
	if err != nil {
		return err
	}
	if exists != 0 {
		return ErrHariDitutup
	}

	closed, err := adaPeriodeDitutup(ctx, tx, tanggal.Format("2006-01-02"))
	if err != nil {
		return err
	}
	if closed {
		return ErrPeriodeDitutup
	}

	err = tx.QueryRowContext(ctx, `
		SELECT EXISTS(SELECT 1 FROM absensi WHERE jadwal_kerja_id = ?)`, jadwalID).Scan(&exists)
	if err != nil {
		return err
	}
	if exists != 0 {
		return ErrPresensiGanda
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO absensi
			(pekerja_id, jadwal_kerja_id, tanggal_absensi, jam_masuk_aktual, jam_keluar_aktual, status_validasi)
		VALUES (?, ?, ?, ?, ?, 'menunggu')`,
		pekerjaID, jadwalID, tanggal.Format("2006-01-02"), masuk, keluar); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *AbsensiRepository) Riwayat(ctx context.Context, pekerjaID int64) ([]RiwayatPresensi, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT a.id, a.tanggal_absensi, s.nama_shift, a.jam_masuk_aktual,
		       a.jam_keluar_aktual, a.menit_terlambat, a.jam_lembur,
		       a.status_validasi, a.catatan_validasi
		FROM absensi a
		JOIN jadwal_kerja j ON j.id = a.jadwal_kerja_id
		JOIN shift s ON s.id = j.shift_id
		WHERE a.pekerja_id = ?
		ORDER BY a.tanggal_absensi DESC, a.id DESC`, pekerjaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []RiwayatPresensi
	for rows.Next() {
		var item RiwayatPresensi
		if err := rows.Scan(&item.ID, &item.Tanggal, &item.NamaShift, &item.JamMasukAktual,
			&item.JamKeluarAktual, &item.MenitTerlambat, &item.JamLembur, &item.Status, &item.Catatan); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, rows.Err()
}
