// Package middleware berisi session, autentikasi/RBAC, logging, dan recover.
// FILE BERSAMA (frozen).
package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"sync"
	"time"

	"penggajian/internal/model"
)

const cookieName = "sid"

type sessionEntry struct {
	user    model.SessionUser
	expires time.Time
}

// Manager menyimpan session di memori server (cukup untuk proyek ini;
// session hilang saat server restart).
type Manager struct {
	mu   sync.RWMutex
	data map[string]sessionEntry
	ttl  time.Duration
}

func NewSessionManager() *Manager {
	return &Manager{data: make(map[string]sessionEntry), ttl: 8 * time.Hour}
}

// Create membuat session baru dan mengirim cookie ke browser. Dipanggil setelah login berhasil.
func (m *Manager) Create(w http.ResponseWriter, u model.SessionUser) error {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return err
	}
	token := hex.EncodeToString(b)

	m.mu.Lock()
	m.data[token] = sessionEntry{user: u, expires: time.Now().Add(m.ttl)}
	m.mu.Unlock()

	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(m.ttl.Seconds()),
	})
	return nil
}

// Destroy menghapus session dan cookie. Dipanggil saat logout.
func (m *Manager) Destroy(w http.ResponseWriter, r *http.Request) {
	if c, err := r.Cookie(cookieName); err == nil {
		m.mu.Lock()
		delete(m.data, c.Value)
		m.mu.Unlock()
	}
	http.SetCookie(w, &http.Cookie{Name: cookieName, Value: "", Path: "/", MaxAge: -1, HttpOnly: true})
}

func (m *Manager) lookup(r *http.Request) *model.SessionUser {
	c, err := r.Cookie(cookieName)
	if err != nil {
		return nil
	}
	m.mu.RLock()
	e, ok := m.data[c.Value]
	m.mu.RUnlock()
	if !ok || time.Now().After(e.expires) {
		return nil
	}
	u := e.user
	return &u
}
