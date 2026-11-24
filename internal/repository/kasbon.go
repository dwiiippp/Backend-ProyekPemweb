package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/shopspring/decimal"
)

type KasbonRepository struct {
	db *sql.DB
}

func NewKasbonRepository(db *sql.DB) *KasbonRepository {
	return &KasbonRepository{db: db}
}

type KasbonRecord struct {
	ID               int64
	PekerjaID        int64
	NamaPekerja      string
	NIK              string
	TanggalPengajuan time.Time
	Nominal          decimal.Decimal
	Sisa             decimal.Decimal
	Keterangan       sql.NullString
	Status           string
}

var ErrPekerjaKasbonTidakDitemukan = errors.New("pekerja tidak ditemukan atau tidak aktif")

func (r *KasbonRepository) Daftar(ctx context.Context) ([]KasbonRecord, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT k.id, k.pekerja_id, u.nama, p.nik, k.tanggal_pengajuan,
		       k.nominal_kasbon, k.sisa_kasbon, k.keterangan, k.status
		FROM kasbon k
		JOIN pekerja p ON p.id = k.pekerja_id AND p.deleted_at IS NULL
		JOIN users u ON u.id = p.user_id AND u.deleted_at IS NULL
		ORDER BY k.tanggal_pengajuan DESC, k.id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanKasbon(rows)
}

func (r *KasbonRepository) MilikPekerja(ctx context.Context, pekerjaID int64) ([]KasbonRecord, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT k.id, k.pekerja_id, u.nama, p.nik, k.tanggal_pengajuan,
		       k.nominal_kasbon, k.sisa_kasbon, k.keterangan, k.status
		FROM kasbon k
		JOIN pekerja p ON p.id = k.pekerja_id AND p.deleted_at IS NULL
		JOIN users u ON u.id = p.user_id AND u.deleted_at IS NULL
		WHERE k.pekerja_id = ?
		ORDER BY k.tanggal_pengajuan DESC, k.id DESC`, pekerjaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanKasbon(rows)
}

func (r *KasbonRepository) Buat(ctx context.Context, pekerjaID int64, nominal decimal.Decimal, keterangan string, tanggal time.Time) error {
	result, err := r.db.ExecContext(ctx, `
		INSERT INTO kasbon
			(pekerja_id, tanggal_pengajuan, nominal_kasbon, sisa_kasbon, keterangan, status)
		SELECT id, ?, ?, ?, NULLIF(?, ''), 'aktif'
		FROM pekerja
		WHERE id = ? AND status = 'aktif' AND deleted_at IS NULL
		  AND EXISTS (
			SELECT 1 FROM users
			WHERE users.id = pekerja.user_id AND users.status = 'aktif' AND users.deleted_at IS NULL
		  )`,
		tanggal.Format("2006-01-02"), nominal, nominal, keterangan, pekerjaID)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected != 1 {
		return ErrPekerjaKasbonTidakDitemukan
	}
	return nil
}

func scanKasbon(rows *sql.Rows) ([]KasbonRecord, error) {
	var result []KasbonRecord
	for rows.Next() {
		var item KasbonRecord
		if err := rows.Scan(&item.ID, &item.PekerjaID, &item.NamaPekerja, &item.NIK,
			&item.TanggalPengajuan, &item.Nominal, &item.Sisa, &item.Keterangan, &item.Status); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, rows.Err()
}
