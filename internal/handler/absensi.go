package handler

import (
	"net/http"

	"penggajian/internal/repository"
	"penggajian/internal/service"
)

// Pemilik: Anggota 2 | F-06 · UC-04 Input Presensi · UC-05 Riwayat (BR-13, BR-15)
// Tabel: absensi (tulis); baca jadwal_kerja, shift, validasi_presensi_harian, periode_gaji
type AbsensiHandler struct {
	d   Deps
	svc *service.AbsensiService
}

func NewAbsensiHandler(d Deps) *AbsensiHandler {
	return &AbsensiHandler{d: d, svc: service.NewAbsensiService(repository.NewAbsensiRepository(d.DB))}
}

func (h *AbsensiHandler) Register(mux *http.ServeMux, g Guards) {
	mux.Handle("GET /pekerja/absensi", g.Pekerja(http.HandlerFunc(h.Form)))
	mux.Handle("POST /pekerja/absensi", g.Pekerja(http.HandlerFunc(h.Simpan)))
	mux.Handle("GET /pekerja/absensi/riwayat", g.Pekerja(http.HandlerFunc(h.Riwayat)))
	// TODO(Anggota 2): tambah route lain di sini. Tolak bila: hari sudah ditutup (ErrHariDitutup), periode ditutup, presensi ganda, jam keluar < jam keluar shift, atau jadwal bukan milik sendiri. Status awal 'menunggu'.
}

func (h *AbsensiHandler) Form(w http.ResponseWriter, r *http.Request) {
	h.d.Render.Page(w, r, "absensi/form.html", PageData{Title: "Input Presensi", Menu: "absensi"})
}

func (h *AbsensiHandler) Simpan(w http.ResponseWriter, r *http.Request) {
	// TODO(Anggota 2): implementasikan.
	http.Error(w, "belum diimplementasi (Anggota 2)", http.StatusNotImplemented)
}

func (h *AbsensiHandler) Riwayat(w http.ResponseWriter, r *http.Request) {
	h.d.Render.Page(w, r, "absensi/riwayat.html", PageData{Title: "Riwayat Presensi", Menu: "absensi"})
}
