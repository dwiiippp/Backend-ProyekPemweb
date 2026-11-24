package handler

import (
	"errors"
	"log"
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

type laporanStatusPeriodeData struct {
	Jenis     string
	Status    string
	Ringkasan service.RingkasanStatusPeriode
}

type laporanRekapGajiData struct {
	Jenis     string
	Mulai     string
	Selesai   string
	Ringkasan service.RingkasanRekapGaji
}

func NewLaporanGajiHandler(d Deps) *LaporanGajiHandler {
	return &LaporanGajiHandler{d: d, svc: service.NewLaporanGajiService(repository.NewLaporanGajiRepository(d.DB))}
}

func (h *LaporanGajiHandler) Register(mux *http.ServeMux, g Guards) {
	mux.Handle("GET /admin/laporan/periode", g.Admin(http.HandlerFunc(h.Periode)))
	mux.Handle("GET /admin/laporan/rekap-gaji", g.Admin(http.HandlerFunc(h.RekapGaji)))
}

func (h *LaporanGajiHandler) Periode(w http.ResponseWriter, r *http.Request) {
	jenis := r.URL.Query().Get("jenis")
	status := r.URL.Query().Get("status")

	ringkasan, err := h.svc.StatusPeriode(r.Context(), jenis, status)
	if err != nil {
		if errors.Is(err, service.ErrInputTidakValid) {
			h.d.Render.Error(w, r, http.StatusBadRequest, "Filter jenis atau status periode tidak valid")
			return
		}
		log.Printf("laporan status periode: %v", err)
		h.d.Render.Error(w, r, http.StatusInternalServerError, "Gagal memuat laporan status periode")
		return
	}

	h.d.Render.Page(w, r, "laporan/periode.html", PageData{
		Title: "Laporan Status Periode",
		Menu:  "laporan-periode",
		Data: laporanStatusPeriodeData{
			Jenis:     jenis,
			Status:    status,
			Ringkasan: ringkasan,
		},
	})
}

func (h *LaporanGajiHandler) RekapGaji(w http.ResponseWriter, r *http.Request) {
	jenis := r.URL.Query().Get("jenis")
	mulai := r.URL.Query().Get("mulai")
	selesai := r.URL.Query().Get("selesai")

	ringkasan, err := h.svc.RekapGaji(r.Context(), jenis, mulai, selesai)
	if err != nil {
		if errors.Is(err, service.ErrInputTidakValid) {
			h.d.Render.Error(w, r, http.StatusBadRequest, "Filter jenis periode atau rentang tanggal tidak valid")
			return
		}
		log.Printf("laporan rekap gaji: %v", err)
		h.d.Render.Error(w, r, http.StatusInternalServerError, "Gagal memuat rekap biaya gaji dan lembur")
		return
	}

	h.d.Render.Page(w, r, "laporan/rekap_gaji.html", PageData{
		Title: "Rekap Biaya Gaji & Lembur",
		Menu:  "laporan-gaji",
		Data: laporanRekapGajiData{
			Jenis:     jenis,
			Mulai:     mulai,
			Selesai:   selesai,
			Ringkasan: ringkasan,
		},
	})
}
