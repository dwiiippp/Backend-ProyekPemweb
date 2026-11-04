package handler

import (
	"net/http"

	"penggajian/internal/repository"
	"penggajian/internal/service"
)

// Pemilik: Anggota 3 | F-12 Laporan status periode & rekap biaya gaji/lembur
// Tabel: periode_gaji, gaji (baca)
type LaporanGajiHandler struct {
	d   Deps
	svc *service.LaporanGajiService
}

func NewLaporanGajiHandler(d Deps) *LaporanGajiHandler {
	return &LaporanGajiHandler{d: d, svc: service.NewLaporanGajiService(repository.NewLaporanGajiRepository(d.DB))}
}

func (h *LaporanGajiHandler) Register(mux *http.ServeMux, g Guards) {
	mux.Handle("GET /admin/laporan/periode", g.Admin(http.HandlerFunc(h.Periode)))
	mux.Handle("GET /admin/laporan/rekap-gaji", g.Admin(http.HandlerFunc(h.RekapGaji)))
	// TODO(Anggota 3): tambah route lain di sini. Hanya baca data.
}

func (h *LaporanGajiHandler) Periode(w http.ResponseWriter, r *http.Request) {
	h.d.Render.Page(w, r, "laporan/periode.html", PageData{Title: "Status Periode", Menu: "laporan-periode"})
}

func (h *LaporanGajiHandler) RekapGaji(w http.ResponseWriter, r *http.Request) {
	h.d.Render.Page(w, r, "laporan/rekap_gaji.html", PageData{Title: "Rekap Gaji & Lembur", Menu: "laporan-gaji"})
}
