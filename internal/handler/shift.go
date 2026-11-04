package handler

import (
	"net/http"

	"penggajian/internal/repository"
	"penggajian/internal/service"
)

// Pemilik: Anggota 1 | F-04 Kelola Shift (BR-05)
// Tabel: shift (tulis)
type ShiftHandler struct {
	d   Deps
	svc *service.ShiftService
}

func NewShiftHandler(d Deps) *ShiftHandler {
	return &ShiftHandler{d: d, svc: service.NewShiftService(repository.NewShiftRepository(d.DB))}
}

func (h *ShiftHandler) Register(mux *http.ServeMux, g Guards) {
	mux.Handle("GET /admin/shift", g.Admin(http.HandlerFunc(h.List)))
	// TODO(Anggota 1): tambah route lain di sini. CRUD shift. Seed: Pagi 07-15, Siang 15-23.
}

func (h *ShiftHandler) List(w http.ResponseWriter, r *http.Request) {
	h.d.Render.Page(w, r, "shift/list.html", PageData{Title: "Master Shift", Menu: "shift"})
}
