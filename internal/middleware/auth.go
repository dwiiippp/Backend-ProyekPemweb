package middleware

import (
	"context"
	"log"
	"net/http"

	"penggajian/internal/model"
)

type ctxKey struct{}

// UserFrom mengambil user yang sedang login dari context (nil bila belum login).
func UserFrom(ctx context.Context) *model.SessionUser {
	u, _ := ctx.Value(ctxKey{}).(*model.SessionUser)
	return u
}

// Load menaruh user (jika ada session) ke context. Bila devRole terisi dan belum ada
// session, user dummy disuntikkan — HANYA untuk development.
func (m *Manager) Load(devRole string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			u, err := m.lookup(r.Context(), r)
			if err != nil {
				log.Printf("Session lookup: %v", err)
				http.Error(w, "Terjadi kesalahan pada server", http.StatusInternalServerError)
				return
			}

			if u == nil && devRole != "" {
				u = devUser(devRole)
			}

			ctx := context.WithValue(r.Context(), ctxKey{}, u)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func devUser(role string) *model.SessionUser {
	switch model.Role(role) {
	case model.RoleAdmin:
		return &model.SessionUser{
			UserID:   1,
			Username: "admin",
			Nama:     "Admin",
			Role:     model.RoleAdmin,
		}

	case model.RolePekerja:
		return &model.SessionUser{
			UserID:    2,
			Username:  "budi",
			Nama:      "Budi Santoso",
			Role:      model.RolePekerja,
			PekerjaID: 1,
		}
	}
	return nil
}

// RequireLogin mengalihkan ke /login bila belum login.
func RequireLogin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if UserFrom(r.Context()) == nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// RequireRole membatasi akses per role. Pemeriksaan dilakukan di server (RBAC).
func RequireRole(role model.Role) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			u := UserFrom(r.Context())
			if u == nil {
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}
			if u.Role != role {
				http.Error(w, "Akses ditolak", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
