package repository

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"penggajian/internal/model"
)

// Pemilik: Anggota 1 | F-01 · UC-01 Login
type AuthRepository struct {
	db *sql.DB
}

func NewAuthRepository(db *sql.DB) *AuthRepository {
	return &AuthRepository{db: db}
}

// AuthRecord menyimpan data user dan relasi pekerja jika tersedia.
type AuthRecord struct {
	User      model.User
	Nama      string
	PekerjaID int64
}

// FindByEmail mencari user berdasarkan email.
// Mengembalikan nil, nil jika email tidak ditemukan.
func (r *AuthRepository) FindByEmail(ctx context.Context, email string) (*AuthRecord, error) {
	var (
		rec       AuthRecord
		role      string
		pekerjaID sql.NullInt64
	)

	email = strings.TrimSpace(strings.ToLower(email))

	err := r.db.QueryRowContext(ctx, `
		SELECT
			u.id,
			u.email,
			u.nama,
			u.password_hash,
			u.role,
			u.status,
			p.id
		FROM users u
		LEFT JOIN pekerja p
			ON p.user_id = u.id
			AND p.deleted_at IS NULL
		WHERE u.email = ?
			AND u.deleted_at IS NULL
	`, email).Scan(
		&rec.User.ID,
		&rec.User.Username,
		&rec.Nama,
		&rec.User.PasswordHash,
		&role,
		&rec.User.Status,
		&pekerjaID,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	rec.User.Role = model.Role(role)

	if pekerjaID.Valid {
		rec.PekerjaID = pekerjaID.Int64
	}

	return &rec, nil
}
