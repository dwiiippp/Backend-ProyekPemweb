package handler

import (
	"net/http"

	"penggajian/internal/repository"
	"penggajian/internal/service"
)

// Pemilik: Anggota 1 | F-03 Riwayat Upah (BR-04)
// Tabel: riwayat_gaji (tulis)
type UpahHandler struct {
	d   Deps
	svc *service.UpahService
}

func NewUpahHandler(d Deps) *UpahHandler {
	return &UpahHandler{d: d, svc: service.NewUpahService(repository.NewUpahRepository(d.DB))}
}

func (h *UpahHandler) Register(mux *http.ServeMux, g Guards) {
	mux.Handle("GET /admin/pekerja/{id}/upah", g.Admin(http.HandlerFunc(h.List)))
	// TODO(Anggota 1): tambah route lain di sini. Tambah upah baru dengan tanggal berlaku; jangan UPDATE baris lama.
}

func (h *UpahHandler) List(w http.ResponseWriter, r *http.Request) {
	h.d.Render.Page(w, r, "upah/list.html", PageData{Title: "Riwayat Upah", Menu: "pekerja"})
}
