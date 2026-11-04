package handler

import (
	"net/http"

	"penggajian/internal/repository"
	"penggajian/internal/service"
)

// Pemilik: Anggota 1 | F-02 · UC-08 Kelola Pekerja
// Tabel: pekerja, users (tulis)
type PekerjaHandler struct {
	d   Deps
	svc *service.PekerjaService
}

func NewPekerjaHandler(d Deps) *PekerjaHandler {
	return &PekerjaHandler{d: d, svc: service.NewPekerjaService(repository.NewPekerjaRepository(d.DB))}
}

func (h *PekerjaHandler) Register(mux *http.ServeMux, g Guards) {
	mux.Handle("GET /admin/pekerja", g.Admin(http.HandlerFunc(h.List)))
	// TODO(Anggota 1): tambah route lain di sini. CRUD pekerja dan akun login. Route lain: GET /admin/pekerja/baru, POST /admin/pekerja, GET /admin/pekerja/{id}, dst.
}

func (h *PekerjaHandler) List(w http.ResponseWriter, r *http.Request) {
	h.d.Render.Page(w, r, "pekerja/list.html", PageData{Title: "Data Pekerja", Menu: "pekerja"})
}
