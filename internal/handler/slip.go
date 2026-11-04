package handler

import (
	"net/http"

	"penggajian/internal/repository"
	"penggajian/internal/service"
)

// Pemilik: Anggota 3 | F-11 · UC-06 Lihat Slip · UC-15 Cetak Slip (BR-12)
// Tabel: gaji, potongan_kasbon_gaji (baca)
type SlipHandler struct {
	d   Deps
	svc *service.SlipService
}

func NewSlipHandler(d Deps) *SlipHandler {
	return &SlipHandler{d: d, svc: service.NewSlipService(repository.NewSlipRepository(d.DB))}
}

func (h *SlipHandler) Register(mux *http.ServeMux, g Guards) {
	mux.Handle("GET /admin/slip", g.Admin(http.HandlerFunc(h.List)))
	mux.Handle("GET /pekerja/slip", g.Pekerja(http.HandlerFunc(h.Saya)))
	// TODO(Anggota 3): tambah route lain di sini. Pekerja hanya boleh melihat slip dengan pekerja_id miliknya.
}

func (h *SlipHandler) List(w http.ResponseWriter, r *http.Request) {
	h.d.Render.Page(w, r, "slip/list.html", PageData{Title: "Slip Gaji", Menu: "slip"})
}

func (h *SlipHandler) Saya(w http.ResponseWriter, r *http.Request) {
	h.d.Render.Page(w, r, "slip/saya.html", PageData{Title: "Slip Gaji Saya", Menu: "slip"})
}
