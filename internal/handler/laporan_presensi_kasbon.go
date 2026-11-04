package handler

import (
	"net/http"

	"penggajian/internal/repository"
	"penggajian/internal/service"
)

// Pemilik: Anggota 2 | F-12 Laporan presensi & rekap kasbon
// Tabel: absensi, kasbon (baca)
type LaporanPresensiKasbonHandler struct {
	d   Deps
	svc *service.LaporanPresensiKasbonService
}

func NewLaporanPresensiKasbonHandler(d Deps) *LaporanPresensiKasbonHandler {
	return &LaporanPresensiKasbonHandler{d: d, svc: service.NewLaporanPresensiKasbonService(repository.NewLaporanPresensiKasbonRepository(d.DB))}
}

func (h *LaporanPresensiKasbonHandler) Register(mux *http.ServeMux, g Guards) {
	mux.Handle("GET /admin/laporan/presensi", g.Admin(http.HandlerFunc(h.Presensi)))
	mux.Handle("GET /admin/laporan/kasbon", g.Admin(http.HandlerFunc(h.Kasbon)))
	// TODO(Anggota 2): tambah route lain di sini. Hanya baca data.
}

func (h *LaporanPresensiKasbonHandler) Presensi(w http.ResponseWriter, r *http.Request) {
	h.d.Render.Page(w, r, "laporan/presensi.html", PageData{Title: "Laporan Presensi", Menu: "laporan-presensi"})
}

func (h *LaporanPresensiKasbonHandler) Kasbon(w http.ResponseWriter, r *http.Request) {
	h.d.Render.Page(w, r, "laporan/kasbon.html", PageData{Title: "Rekap Kasbon", Menu: "laporan-kasbon"})
}
