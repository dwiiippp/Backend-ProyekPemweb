package repository

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"time"

	"penggajian/internal/model"
)

type SessionRepository struct {
	db *sql.DB
}

func NewSessionRepository(db *sql.DB) *SessionRepository {
	return &SessionRepository{db: db}
}

// Create membuat sesi baru dengan token acak 128-bit.
// Masa berlaku sesi adalah 7 hari.
func (r *SessionRepository) Create(ctx context.Context, userID int64) (string, error) {
	tokenBytes := make([]byte, 16)

	if _, err := rand.Read(tokenBytes); err != nil {
		return "", err
	}

	token := hex.EncodeToString(tokenBytes)
	expiresAt := time.Now().Add(7 * 24 * time.Hour)

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO sessions (token, user_id, expires_at)
		VALUES (?, ?, ?)
	`, token, userID, expiresAt)

	if err != nil {
		return "", err
	}

	return token, nil
}

// FindUserByToken mengambil pengguna dari token sesi yang masih berlaku.
func (r *SessionRepository) FindUserByToken(ctx context.Context, token string) (*model.SessionUser, error) {
	var user model.SessionUser

	err := r.db.QueryRowContext(ctx, `
		SELECT
			u.id,
			u.email,
			u.nama,
			u.role,
			COALESCE(p.id, 0)
		FROM sessions s
		JOIN users u ON u.id = s.user_id
		LEFT JOIN pekerja p
			ON p.user_id = u.id
			AND p.deleted_at IS NULL
		WHERE s.token = ?
			AND s.expires_at > NOW()
			AND u.status = 'aktif'
			AND u.deleted_at IS NULL
	`, token).Scan(
		&user.UserID,
		&user.Username,
		&user.Nama,
		&user.Role,
		&user.PekerjaID,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// Delete menghapus sesi berdasarkan token.
func (r *SessionRepository) Delete(ctx context.Context, token string) error {
	_, err := r.db.ExecContext(ctx, `
		DELETE FROM sessions
		WHERE token = ?
	`, token)

	return err
}
