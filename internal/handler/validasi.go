package handler

import (
	"net/http"

	"penggajian/internal/repository"
	"penggajian/internal/service"
)

// Pemilik: Anggota 2 | F-14 · UC-10 Validasi Presensi Harian (BR-14)
// Tabel: absensi (ubah), validasi_presensi_harian (tulis)
type ValidasiHandler struct {
	d   Deps
	svc *service.ValidasiService
}

func NewValidasiHandler(d Deps) *ValidasiHandler {
	return &ValidasiHandler{d: d, svc: service.NewValidasiService(repository.NewValidasiRepository(d.DB))}
}

func (h *ValidasiHandler) Register(mux *http.ServeMux, g Guards) {
	mux.Handle("GET /admin/validasi", g.Admin(http.HandlerFunc(h.List)))
	mux.Handle("GET /admin/validasi/{tanggal}", g.Admin(http.HandlerFunc(h.Detail)))
	mux.Handle("POST /admin/validasi/{tanggal}", g.Admin(http.HandlerFunc(h.Simpan)))
	// TODO(Anggota 2): tambah route lain di sini. Saat validasi: hitung menit_terlambat & jam_lembur untuk yang disetujui, lalu INSERT validasi_presensi_harian — semuanya dalam satu transaksi.
}

func (h *ValidasiHandler) List(w http.ResponseWriter, r *http.Request) {
	h.d.Render.Page(w, r, "validasi/list.html", PageData{Title: "Validasi Presensi Harian", Menu: "validasi"})
}

func (h *ValidasiHandler) Detail(w http.ResponseWriter, r *http.Request) {
	h.d.Render.Page(w, r, "validasi/detail.html", PageData{Title: "Validasi Presensi", Menu: "validasi"})
}

func (h *ValidasiHandler) Simpan(w http.ResponseWriter, r *http.Request) {
	// TODO(Anggota 2): implementasikan.
	http.Error(w, "belum diimplementasi (Anggota 2)", http.StatusNotImplemented)
}
