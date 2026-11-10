// Package middleware berisi session, autentikasi/RBAC, logging, dan recover.
// FILE BERSAMA (frozen).
package middleware

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"penggajian/internal/model"
	"penggajian/internal/repository"
)

const (
	cookieName = "session"
	sessionTTL = 7 * 24 * time.Hour
)

type Manager struct {
	repo   *repository.SessionRepository
	secure bool
}

func NewSessionManager(db *sql.DB, cookieSecure bool) *Manager {
	return &Manager{
		repo:   repository.NewSessionRepository(db),
		secure: cookieSecure,
	}
}

// Create menyimpan sesi ke database dan mengirim token melalui cookie
func (m *Manager) Create(ctx context.Context, w http.ResponseWriter, u model.SessionUser) error {
	token, err := m.repo.Create(ctx, u.UserID)
	if err != nil {
		return err
	}

	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    token,
		Path:     "/",
		MaxAge:   int(sessionTTL.Seconds()),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   m.secure,
	})
	return nil
}

// Destroy menghapus session dan cookie
func (m *Manager) Destroy(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	var deleteErr error

	cookie, err := r.Cookie(cookieName)
	if err == nil {
		deleteErr = m.repo.Delete(ctx, cookie.Value)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   m.secure,
	})
	return deleteErr

}

func (m *Manager) lookup(ctx context.Context, r *http.Request) (*model.SessionUser, error) {
	cookie, err := r.Cookie(cookieName)
	if err != nil {
		return nil, nil
	}
	return m.repo.FindUserByToken(ctx, cookie.Value)
}
