package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/go-sql-driver/mysql"
)

var ErrDuplicateUpah = errors.New("tanggal berlaku upah sudah digunakan")

type UpahRecord struct {
	ID             int64
	PekerjaID      int64
	NominalGaji    float64
	TanggalBerlaku time.Time
	CreatedAt      time.Time
}

type CreateUpahParams struct {
	PekerjaID      int64
	NominalGaji    float64
	TanggalBerlaku time.Time
}

type UpahRepository struct {
	db *sql.DB
}

func NewUpahRepository(db *sql.DB) *UpahRepository {
	return &UpahRepository{db: db}
}

func (r *UpahRepository) GetByPekerjaID(ctx context.Context, pekerjaID int64) ([]UpahRecord, error) {
	rows, err := r.db.QueryContext(ctx, `
	SELECT id, pekerja_id, nominal_gaji, tanggal_berlaku, created_at FROM riwayat_gaji WHERE pekerja_id = ? ORDER BY tanggal_berlaku DESC, id DESC
	`, pekerjaID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []UpahRecord

	for rows.Next() {
		var record UpahRecord

		err := rows.Scan(
			&record.ID,
			&record.PekerjaID,
			&record.NominalGaji,
			&record.TanggalBerlaku,
			&record.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		records = append(records, record)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return records, nil
}

func (r *UpahRepository) Create(ctx context.Context, input CreateUpahParams) error {
	_, err := r.db.ExecContext(ctx, `
	INSERT INTO riwayat_gaji (pekerja_id, nominal_gaji, tanggal_berlaku) VALUES (?, ?, ?)
	`,
		input.PekerjaID, input.NominalGaji, input.TanggalBerlaku.Format("2006-01-02"),
	)

	if err != nil {
		var mysqlErr *mysql.MySQLError

		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			return ErrDuplicateUpah
		}

		return err
	}

	return nil

}
