package handler

import (
	"net/http"

	"penggajian/internal/middleware"
	"penggajian/internal/repository"
	"penggajian/internal/service"
)

// Pemilik: Anggota 1 | F-01 · UC-01 Login
// Tabel: users (tulis/baca)
type AuthHandler struct {
	d   Deps
	svc *service.AuthService
}

func NewAuthHandler(d Deps) *AuthHandler {
	return &AuthHandler{d: d, svc: service.NewAuthService(repository.NewAuthRepository(d.DB))}
}

func (h *AuthHandler) Register(mux *http.ServeMux, g Guards) {
	mux.HandleFunc("GET /login", h.LoginForm)
	mux.HandleFunc("POST /login", h.DoLogin)
	mux.Handle("POST /logout", g.Login(http.HandlerFunc(h.Logout)))
	// TODO(Anggota 1): tambah route lain di sini. Validasi username+password (service.CheckPassword), tolak akun nonaktif, buat session via d.Sessions.Create, redirect sesuai role.
}

func (h *AuthHandler) LoginForm(w http.ResponseWriter, r *http.Request) {
	if middleware.UserFrom(r.Context()) != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	h.d.Render.Page(w, r, "login.html", PageData{Title: "Login", Menu: ""})
}

func (h *AuthHandler) DoLogin(w http.ResponseWriter, r *http.Request) {
	// TODO(Anggota 1): implementasikan.
	http.Error(w, "belum diimplementasi (Anggota 1)", http.StatusNotImplemented)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// TODO(Anggota 1): implementasikan.
	http.Error(w, "belum diimplementasi (Anggota 1)", http.StatusNotImplemented)
}
