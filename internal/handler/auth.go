package handler

import (
	"errors"
	"log"
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
}

func (h *AuthHandler) LoginForm(w http.ResponseWriter, r *http.Request) {
	if middleware.UserFrom(r.Context()) != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	h.d.Render.Page(w, r, "login.html", PageData{Title: "Login", Menu: ""})
}

func (h *AuthHandler) DoLogin(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		h.d.Render.Error(w, r, http.StatusBadRequest, "Form tidak valid")
		return
	}

	u, err := h.svc.Login(r.Context(), r.FormValue("email"), r.FormValue("password"))
	if err != nil {
		msg := "Terjadi kesalahan, coba lagi"
		switch {
		case errors.Is(err, service.ErrInputTidakValid):
			msg = "Email dan password wajib diisi"
		case errors.Is(err, service.ErrKredensialSalah), errors.Is(err, service.ErrAkunNonaktif):
			msg = "Email atau password salah"
		default:
			log.Printf("login: %v", err)
		}
		h.d.Render.Page(w, r, "login.html", PageData{Title: "Login", Error: msg})
		return
	}

	if err := h.d.Sessions.Create(r.Context(), w, *u); err != nil {
		log.Printf("login session: %v", err)
		h.d.Render.Error(w, r, http.StatusInternalServerError, "Gagal membuat sesi")
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther) // home menampilkan dashboard sesuai role
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	if err := h.d.Sessions.Destroy(r.Context(), w, r); err != nil {
		log.Printf("logout session: %v", err)
		h.d.Render.Error(w, r, http.StatusInternalServerError, "Gagal menghapus sesi")
		return
	}
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
