package handler

import (
	"net/http"

	"penggajian/internal/repository"
	"penggajian/internal/service"
)

// Pemilik: Anggota 3 | F-07 · F-09 · UC-12 Proses Penggajian · UC-13 Tutup Periode
// Tabel: gaji, potongan_kasbon_gaji (tulis); kasbon, periode_gaji (ubah); baca absensi, riwayat_gaji, pekerja, koreksi_gaji
type PenggajianHandler struct {
	d   Deps
	svc *service.PenggajianService
}

func NewPenggajianHandler(d Deps) *PenggajianHandler {
	return &PenggajianHandler{d: d, svc: service.NewPenggajianService(repository.NewPenggajianRepository(d.DB))}
}

func (h *PenggajianHandler) Register(mux *http.ServeMux, g Guards) {
	mux.Handle("GET /admin/penggajian", g.Admin(http.HandlerFunc(h.Index)))
	// TODO(Anggota 3): tambah route lain di sini. Rumus di service/penggajian.go. Prasyarat: semua hari periode sudah divalidasi. Konfirmasi = SATU transaksi (gaji, potongan_kasbon_gaji, kasbon, periode).
}

func (h *PenggajianHandler) Index(w http.ResponseWriter, r *http.Request) {
	h.d.Render.Page(w, r, "penggajian/index.html", PageData{Title: "Proses Penggajian", Menu: "penggajian"})
}
