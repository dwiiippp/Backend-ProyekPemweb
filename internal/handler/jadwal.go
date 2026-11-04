package handler

import (
	"net/http"

	"penggajian/internal/repository"
	"penggajian/internal/service"
)

// Pemilik: Anggota 1 | F-05 Jadwal Kerja · UC-03
// Tabel: jadwal_kerja (tulis)
type JadwalHandler struct {
	d   Deps
	svc *service.JadwalService
}

func NewJadwalHandler(d Deps) *JadwalHandler {
	return &JadwalHandler{d: d, svc: service.NewJadwalService(repository.NewJadwalRepository(d.DB))}
}

func (h *JadwalHandler) Register(mux *http.ServeMux, g Guards) {
	mux.Handle("GET /admin/jadwal", g.Admin(http.HandlerFunc(h.Kalender)))
	mux.Handle("GET /pekerja/jadwal", g.Pekerja(http.HandlerFunc(h.Saya)))
	// TODO(Anggota 1): tambah route lain di sini. Atur jadwal (UNIQUE pekerja_id+tanggal). Pekerja hanya melihat jadwal miliknya (PekerjaID dari session).
}

func (h *JadwalHandler) Kalender(w http.ResponseWriter, r *http.Request) {
	h.d.Render.Page(w, r, "jadwal/kalender.html", PageData{Title: "Jadwal Kerja", Menu: "jadwal"})
}

func (h *JadwalHandler) Saya(w http.ResponseWriter, r *http.Request) {
	h.d.Render.Page(w, r, "jadwal/saya.html", PageData{Title: "Jadwal Saya", Menu: "jadwal"})
}
