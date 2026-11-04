package handler

import (
	"net/http"

	"penggajian/internal/repository"
	"penggajian/internal/service"
)

// Pemilik: Anggota 2 | F-08 · UC-11 Kelola Kasbon · UC-07 Riwayat Kasbon
// Tabel: kasbon (tulis; sisa_kasbon juga diubah modul penggajian saat tutup periode)
type KasbonHandler struct {
	d   Deps
	svc *service.KasbonService
}

func NewKasbonHandler(d Deps) *KasbonHandler {
	return &KasbonHandler{d: d, svc: service.NewKasbonService(repository.NewKasbonRepository(d.DB))}
}

func (h *KasbonHandler) Register(mux *http.ServeMux, g Guards) {
	mux.Handle("GET /admin/kasbon", g.Admin(http.HandlerFunc(h.List)))
	mux.Handle("GET /pekerja/kasbon", g.Pekerja(http.HandlerFunc(h.Saya)))
	// TODO(Anggota 2): tambah route lain di sini. Saat dibuat: sisa_kasbon = nominal_kasbon, status 'aktif'.
}

func (h *KasbonHandler) List(w http.ResponseWriter, r *http.Request) {
	h.d.Render.Page(w, r, "kasbon/list.html", PageData{Title: "Kasbon", Menu: "kasbon"})
}

func (h *KasbonHandler) Saya(w http.ResponseWriter, r *http.Request) {
	h.d.Render.Page(w, r, "kasbon/saya.html", PageData{Title: "Kasbon Saya", Menu: "kasbon"})
}
