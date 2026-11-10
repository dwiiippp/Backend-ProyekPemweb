package repository

import (
	"context"
	"database/sql"
	"time"
)

// Pemilik: Anggota 1 | F-02 · UC-08 Kelola Pekerja
// Tabel: pekerja, users (tulis)
// Semua query SQL modul ini ditulis di sini (pakai placeholder ?, jangan concat string).
type PekerjaRepository struct {
	db *sql.DB
}

func NewPekerjaRepository(db *sql.DB) *PekerjaRepository {
	return &PekerjaRepository{db: db}
}

// Pekerja record menanmpung data pekerja beserta akun login
type PekerjaRecord struct {
	ID            int64
	UserID        int64
	NIK           string
	Nama          string
	Email         string
	KategoriKerja string
	TanggalMasuk  time.Time
	TanggalKeluar *time.Time
	Status        string
}

// CreatePekerjaParams untuk membuat pekerja dan akun login
type CreatePekerjaParams struct {
	NIK           string
	Nama          string
	Email         string
	PasswordHash  string
	KategoriKerja string
	TanggalMasuk  time.Time
}

// GetAll mengambil seluruh pekerja yang belum dihapus
func (r *PekerjaRepository) GetAll(ctx context.Context) ([]PekerjaRecord, error) {
	rows, err := r.db.QueryContext(ctx, `
	SELECT
		p.id,
		p.user_id,
		p.nik,
		u.nama,
		u.email,
		p.kategori_kerja,
		p.tanggal_masuk,
		p.tanggal_keluar,
		p.status
	FROM pekerja p JOIN users u ON u.id = p.user_id
	WHERE p.deleted_at IS NULL AND u.deleted_at IS NULL
	ORDER BY p.id DESC
	`)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pekerja []PekerjaRecord

	for rows.Next() {
		var (
			p             PekerjaRecord
			tanggalKeluar sql.NullTime
		)

		err := rows.Scan(
			&p.ID,
			&p.UserID,
			&p.NIK,
			&p.Nama,
			&p.Email,
			&p.KategoriKerja,
			&p.TanggalMasuk,
			&tanggalKeluar,
			&p.Status,
		)
		if err != nil {
			return nil, err
		}

		if tanggalKeluar.Valid {
			p.TanggalKeluar = &tanggalKeluar.Time
		}

		pekerja = append(pekerja, p)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return pekerja, nil
}

// GetByID mengambil pekerja berdasarkan ID
func (r *PekerjaRepository) GetByID(ctx context.Context, id int64) (*PekerjaRecord, error) {
	var (
		p             PekerjaRecord
		tanggalKeluar sql.NullTime
	)

	err := r.db.QueryRowContext(ctx, `
	SELECT
		p.id,
		p.user_id,
		p.nik,
		u.nama,
		u.email,
		p.kategori_kerja,
		p.tanggal_masuk,
		p.tanggal_keluar,
		p.status
	FROM pekerja p JOIN users u ON u.id = p.user_id
	WHERE p.id = ? AND p.deleted_at IS NULL AND u.deleted_at IS NULL
	`, id).Scan(
		&p.ID,
		&p.UserID,
		&p.NIK,
		&p.Nama,
		&p.Email,
		&p.KategoriKerja,
		&p.TanggalMasuk,
		&tanggalKeluar,
		&p.Status,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if tanggalKeluar.Valid {
		p.TanggalKeluar = &tanggalKeluar.Time
	}

	return &p, nil
}

// CreayeWithAccount membuat akun login dan data pekerja
func (r *PekerjaRepository) CreateWithAccount(ctx context.Context, input CreatePekerjaParams) (int64, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	//buat akun login
	result, err := tx.ExecContext(ctx,
		`INSERT INTO users (nama, email, password_hash, role, status) VALUES (?, ?, ?, 'pekerja', 'aktif')`, input.Nama, input.Email, input.PasswordHash)
	if err != nil {
		return 0, err
	}

	userID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	// Membuat data pekerja
	result, err = tx.ExecContext(ctx,
		`INSERT INTO pekerja (user_id, nik, kategori_kerja, tanggal_masuk, status) VALUES (?, ?, ?, ?, 'aktif')`,
		userID, input.NIK, input.KategoriKerja, input.TanggalMasuk)

	if err != nil {
		return 0, err
	}

	pekerjaID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	//Menyimpan seluruh perubahan
	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return pekerjaID, nil
}

//UPDATE DAN DEACTIVATE AKUN PEKERJA

type UpdatePekerjaParams struct {
	ID            int64
	NIK           string
	Nama          string
	Email         string
	KategoriKerja string
	TanggalMasuk  time.Time
}

// Update akun dan data pekerja
func (r *PekerjaRepository) Update(ctx context.Context, input UpdatePekerjaParams) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var userID int64

	err = tx.QueryRowContext(ctx,
		`SELECT user_id FROM pekerja WHERE id = ? AND deleted_at IS NULL`,
		input.ID,
	).Scan(&userID)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx,
		`UPDATE users SET nama = ?, email = ? WHERE id = ? AND deleted_at IS NULL`,
		input.Nama, input.Email, userID,
	)

	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx,
		`UPDATE pekerja SET nik = ?, kategori_kerja = ?, tanggal_masuk = ? WHERE id = ? AND deleted_at IS NULL`,
		input.NIK, input.KategoriKerja, input.TanggalMasuk, input.ID,
	)

	if err != nil {
		return err
	}

	return tx.Commit()
}

// Deactivate akun dan data pekerja
func (r *PekerjaRepository) Deactivate(ctx context.Context, id int64) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var userID int64

	err = tx.QueryRowContext(ctx,
		`SELECT user_id FROM pekerja WHERE id = ? AND deleted_at IS NULL`,
		id,
	).Scan(&userID)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx,
		`UPDATE pekerja SET status = 'nonaktif', tanggal_keluar = COALESCE(tanggal_keluar, CURDATE()) WHERE id = ?`,
		id,
	)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx,
		`UPDATE users SET status = 'nonaktif' WHERE id=?`,
		userID,
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}
