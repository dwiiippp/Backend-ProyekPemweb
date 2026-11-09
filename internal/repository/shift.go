package repository

import (
	"context"
	"database/sql"
	"errors"

	"penggajian/internal/model"
)

var ErrShiftNotFound = errors.New("Shift tidak ditemukan")

// Tabel: shift (tulis)
// Semua query SQL modul ini ditulis di sini (pakai placeholder ?, jangan concat string).
type ShiftRepository struct {
	db *sql.DB
}

func NewShiftRepository(db *sql.DB) *ShiftRepository {
	return &ShiftRepository{db: db}
}

type CreateShiftParams struct {
	NamaShift string
	JamMasuk  string
	JamKeluar string
}

// GetAll mengambil semua shift yang belum dihapus.
func (r *ShiftRepository) GetAll(ctx context.Context) ([]model.Shift, error) {
	rows, err := r.db.QueryContext(ctx, `
	SELECT id, nama_shift, TIME_FORMAT(jam_masuk, '%H:%i:%s'), TIME_FORMAT(jam_keluar, '%H:%i:%s')
	FROM shift WHERE deleted_at IS NULL ORDER BY id ASC`)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var shifts []model.Shift

	for rows.Next() {
		var shift model.Shift

		if err := rows.Scan(
			&shift.ID,
			&shift.NamaShift,
			&shift.JamMasuk,
			&shift.JamKeluar,
		); err != nil {
			return nil, err
		}

		shifts = append(shifts, shift)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return shifts, nil
}

// GetByID mengambil satu shift berdasarkan ID
func (r *ShiftRepository) GetByID(ctx context.Context, id int64) (model.Shift, error) {
	var shift model.Shift

	err := r.db.QueryRowContext(ctx, `
		SELECT id, nama_shift, TIME_FORMAT(jam_masuk, '%H:%i:%s'), TIME_FORMAT(jam_keluar, '%H:%i:%s')
		FROM shift WHERE id = ? AND deleted_at IS NULL
	`, id).Scan(
		&shift.ID,
		&shift.NamaShift,
		&shift.JamMasuk,
		&shift.JamKeluar,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return shift, ErrShiftNotFound
	}

	return shift, err
}

func (r *ShiftRepository) Create(ctx context.Context, input CreateShiftParams) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO shift (nama_shift, jam_masuk, jam_keluar) VALUES (?, ?, ?)
	`, input.NamaShift, input.JamMasuk, input.JamKeluar)

	return err
}

func (r *ShiftRepository) Update(ctx context.Context, id int64, input CreateShiftParams) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE shift SET nama_shift = ?, jam_masuk = ?, jam_keluar = ? WHERE id = ? AND deleted_at IS NULL
	`, input.NamaShift, input.JamMasuk, input.JamKeluar, id)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if affected == 0 {
		return ErrShiftNotFound
	}

	return nil
}

func (r *ShiftRepository) Delete(ctx context.Context, id int64) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE shift SET deleted_at = CURRENT_TIMESTAMP WHERE id = ? AND deleted_at IS NULL
	`, id)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if affected == 0 {
		return ErrShiftNotFound
	}

	return nil
}
