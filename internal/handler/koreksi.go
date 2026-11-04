package handler

import (
	"net/http"

	"penggajian/internal/repository"
	"penggajian/internal/service"
)

// Pemilik: Anggota 3 | F-10 · UC-14 Koreksi Gaji (BR-11)
// Tabel: koreksi_gaji (tulis)
type KoreksiHandler struct {
	d   Deps
	svc *service.KoreksiService
}

func NewKoreksiHandler(d Deps) *KoreksiHandler {
	return &KoreksiHandler{d: d, svc: service.NewKoreksiService(repository.NewKoreksiRepository(d.DB))}
}

func (h *KoreksiHandler) Register(mux *http.ServeMux, g Guards) {
	mux.Handle("GET /admin/koreksi", g.Admin(http.HandlerFunc(h.List)))
	// TODO(Anggota 3): tambah route lain di sini. Koreksi dihitung langsung dari koreksi_gaji per periode_gaji_tujuan_id agar tidak terhitung dua kali.
}

func (h *KoreksiHandler) List(w http.ResponseWriter, r *http.Request) {
	h.d.Render.Page(w, r, "koreksi/list.html", PageData{Title: "Koreksi Gaji", Menu: "koreksi"})
}
