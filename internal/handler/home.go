package handler

import (
	"net/http"

	"penggajian/internal/repository"
	"penggajian/internal/service"
)

// Pemilik: Anggota 1 | UC-02 · F-13 Dashboard & profil
// Tabel: pekerja (baca)
type HomeHandler struct {
	d   Deps
	svc *service.HomeService
}

func NewHomeHandler(d Deps) *HomeHandler {
	return &HomeHandler{d: d, svc: service.NewHomeService(repository.NewHomeRepository(d.DB))}
}

func (h *HomeHandler) Register(mux *http.ServeMux, g Guards) {
	mux.Handle("GET /{$}", g.Login(http.HandlerFunc(h.Index)))
	mux.Handle("GET /profil", g.Pekerja(http.HandlerFunc(h.Profil)))
	// TODO(Anggota 1): tambah route lain di sini. Dashboard per role dan halaman profil pekerja (NIK, nama, kategori kerja).
}

func (h *HomeHandler) Index(w http.ResponseWriter, r *http.Request) {
	h.d.Render.Page(w, r, "home.html", PageData{Title: "Dashboard", Menu: "home"})
}

func (h *HomeHandler) Profil(w http.ResponseWriter, r *http.Request) {
	h.d.Render.Page(w, r, "profil.html", PageData{Title: "Profil Saya", Menu: "profil"})
}
