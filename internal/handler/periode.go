package handler

import (
	"net/http"

	"penggajian/internal/repository"
	"penggajian/internal/service"
)

// Pemilik: Anggota 3 | Periode Gaji (BR-01, BR-02, BR-10)
// Tabel: periode_gaji (tulis)
type PeriodeHandler struct {
	d   Deps
	svc *service.PeriodeService
}

func NewPeriodeHandler(d Deps) *PeriodeHandler {
	return &PeriodeHandler{d: d, svc: service.NewPeriodeService(repository.NewPeriodeRepository(d.DB))}
}

func (h *PeriodeHandler) Register(mux *http.ServeMux, g Guards) {
	mux.Handle("GET /admin/periode", g.Admin(http.HandlerFunc(h.List)))
	// TODO(Anggota 3): tambah route lain di sini. Buat periode baru; periode ditutup tidak boleh diubah.
}

func (h *PeriodeHandler) List(w http.ResponseWriter, r *http.Request) {
	h.d.Render.Page(w, r, "periode/list.html", PageData{Title: "Periode Gaji", Menu: "periode"})
}
