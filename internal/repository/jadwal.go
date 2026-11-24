package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/go-sql-driver/mysql"
)

// Tabel: jadwal_kerja (tulis)
// Semua query SQL modul ini ditulis di sini (pakai placeholder ?, jangan concat string).

type CreateJadwalParams struct {
	PekerjaID     int64
	ShiftID       int64
	TanggalJadwal time.Time
}

type JadwalRecord struct {
	ID            int64
	PekerjaID     int64
	NamaPekerja   string
	NIK           string
	ShiftID       int64
	NamaShift     string
	JamMasuk      string
	JamKeluar     string
	TanggalJadwal time.Time
}

type PekerjaOption struct {
	ID   int64
	Nama string
}

type ShiftOption struct {
	ID        int64
	NamaShift string
}

type JadwalRepository struct {
	db *sql.DB
}

var (
	ErrJadwalNotFound = errors.New("jadwal tidak ditemukan")
	ErrJadwalDuplikat = errors.New("pekerja sudah memiliki jadwal pada tanggal tersebut")
)

func mapJadwalDBError(err error) error {
	if err == nil {
		return nil
	}

	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
		return ErrJadwalDuplikat
	}

	return err
}

func NewJadwalRepository(db *sql.DB) *JadwalRepository {
	return &JadwalRepository{db: db}
}

func (r *JadwalRepository) GetPekerjaAktif(ctx context.Context) ([]PekerjaOption, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT p.id, u.nama FROM pekerja p JOIN users u ON u.id = p.user_id
		WHERE p.status = 'aktif' AND p.deleted_at IS NULL AND u.status = 'aktif' AND u.deleted_at IS NULL ORDER BY u.nama ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []PekerjaOption

	for rows.Next() {
		var item PekerjaOption

		if err := rows.Scan(&item.ID, &item.Nama); err != nil {
			return nil, err
		}

		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func (r *JadwalRepository) GetShiftAktif(ctx context.Context) ([]ShiftOption, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, nama_shift FROM shift WHERE deleted_at IS NULL ORDER BY id ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []ShiftOption

	for rows.Next() {
		var item ShiftOption

		if err := rows.Scan(&item.ID, &item.NamaShift); err != nil {
			return nil, err
		}

		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func (r *JadwalRepository) GetKalender(ctx context.Context, tanggalMulai time.Time, tanggalSelesai time.Time) ([]JadwalRecord, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT
			j.id,
			j.pekerja_id,
			u.nama,
			p.nik,
			j.shift_id,
			s.nama_shift,
			TIME_FORMAT(s.jam_masuk, '%H:%i:%s'),
			TIME_FORMAT(s.jam_keluar, '%H:%i:%s'),
			j.tanggal_jadwal
		FROM jadwal_kerja j
		JOIN pekerja p ON p.id = j.pekerja_id
		JOIN users u ON U.id = p.user_id
		JOIN shift s ON s.id = j.shift_id
		WHERE j.tanggal_jadwal BETWEEN ? AND ?
			AND p.deleted_at IS NULL
			AND u.deleted_at IS NULL
		ORDER BY j.tanggal_jadwal ASC, u.nama ASC
	`, tanggalMulai.Format("2006-01-02"), tanggalSelesai.Format("2006-01-02"))

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []JadwalRecord

	for rows.Next() {
		var item JadwalRecord

		if err := rows.Scan(
			&item.ID,
			&item.PekerjaID,
			&item.NamaPekerja,
			&item.NIK,
			&item.ShiftID,
			&item.NamaShift,
			&item.JamMasuk,
			&item.JamKeluar,
			&item.TanggalJadwal,
		); err != nil {
			return nil, err
		}

		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func (r *JadwalRepository) GetJadwalPekerja(ctx context.Context, pekerjaID int64) ([]JadwalRecord, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT
			j.id,
			j.pekerja_id,
			u.nama,
			p.nik,
			j.shift_id,
			s.nama_shift,
			TIME_FORMAT(s.jam_masuk, '%H:%i:%s'),
			TIME_FORMAT(s.jam_keluar, '%H:%i:%s'),
			j.tanggal_jadwal
		FROM jadwal_kerja j
		JOIN pekerja p ON p.id = j.pekerja_id
		JOIN users u ON u.id = p.user_id
		JOIN shift s ON s.id = j.shift_id
		WHERE j.pekerja_id = ?
			AND p.deleted_at IS NULL
			AND u.deleted_at IS NULL
		ORDER BY j.tanggal_jadwal DESC
	`, pekerjaID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []JadwalRecord

	for rows.Next() {
		var item JadwalRecord

		if err := rows.Scan(
			&item.ID,
			&item.PekerjaID,
			&item.NamaPekerja,
			&item.NIK,
			&item.ShiftID,
			&item.NamaShift,
			&item.JamMasuk,
			&item.JamKeluar,
			&item.TanggalJadwal,
		); err != nil {
			return nil, err
		}

		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func (r *JadwalRepository) GetByID(ctx context.Context, id int64) (JadwalRecord, error) {
	var item JadwalRecord

	err := r.db.QueryRowContext(ctx, `
		SELECT
			j.id,
			j.pekerja_id,
			u.nama,
			p.nik,
			j.shift_id,
			s.nama_shift,
			TIME_FORMAT(s.jam_masuk, '%H:%i:%s'),
			TIME_FORMAT(s.jam_keluar, '%H:%i:%s'),
			j.tanggal_jadwal
		FROM jadwal_kerja j
		JOIN pekerja p ON p.id = j.pekerja_id
		JOIN users u ON u.id = p.user_id
		JOIN shift s ON s.id = j.shift_id
		WHERE j.id = ?
			AND p.deleted_at IS NULL
			AND u.deleted_at IS NULL 
	`, id).Scan(
		&item.ID,
		&item.PekerjaID,
		&item.NamaPekerja,
		&item.NIK,
		&item.ShiftID,
		&item.NamaShift,
		&item.JamMasuk,
		&item.JamKeluar,
		&item.TanggalJadwal,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return JadwalRecord{}, ErrJadwalNotFound
	}

	if err != nil {
		return JadwalRecord{}, err
	}

	return item, nil
}

func (r *JadwalRepository) Create(ctx context.Context, input CreateJadwalParams) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO jadwal_kerja (pekerja_id, shift_id, tanggal_jadwal) VALUES (?, ?, ?)
	`, input.PekerjaID, input.ShiftID, input.TanggalJadwal.Format("2006-01-02"))
	return mapJadwalDBError(err)
}

func (r *JadwalRepository) Update(ctx context.Context, id int64, input CreateJadwalParams) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE jadwal_kerja SET pekerja_id = ?, shift_id = ?, tanggal_jadwal = ? WHERE id = ?
	`, input.PekerjaID, input.ShiftID, input.TanggalJadwal.Format("2006-01-02"), id)

	if err != nil {
		return mapJadwalDBError(err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if affected == 0 {
		return ErrJadwalNotFound
	}

	return nil
}
