package handler

import (
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

	"penggajian/internal/repository"
	"penggajian/internal/service"
)

type LaporanPresensiKasbonHandler struct {
	d   Deps
	svc *service.LaporanPresensiKasbonService
}

type laporanPresensiData struct {
	Mulai, Selesai string
	PekerjaID      int64
	Pekerja        []repository.PekerjaFilter
	Presensi       []repository.BarisLaporanPresensi
}

func NewLaporanPresensiKasbonHandler(d Deps) *LaporanPresensiKasbonHandler {
	return &LaporanPresensiKasbonHandler{d: d, svc: service.NewLaporanPresensiKasbonService(repository.NewLaporanPresensiKasbonRepository(d.DB))}
}

func (h *LaporanPresensiKasbonHandler) Register(mux *http.ServeMux, g Guards) {
	mux.Handle("GET /admin/laporan/presensi", g.Admin(http.HandlerFunc(h.Presensi)))
	mux.Handle("GET /admin/laporan/kasbon", g.Admin(http.HandlerFunc(h.Kasbon)))
}

func (h *LaporanPresensiKasbonHandler) Presensi(w http.ResponseWriter, r *http.Request) {
	today := time.Now()
	startDefault := today.AddDate(0, 0, -30).Format("2006-01-02")
	endDefault := today.Format("2006-01-02")
	mulai := r.URL.Query().Get("mulai")
	selesai := r.URL.Query().Get("selesai")
	if mulai == "" {
		mulai = startDefault
	}
	if selesai == "" {
		selesai = endDefault
	}
	pekerjaID := int64(0)
	if value := r.URL.Query().Get("pekerja_id"); value != "" {
		var err error
		pekerjaID, err = strconv.ParseInt(value, 10, 64)
		if err != nil || pekerjaID < 0 {
			h.d.Render.Error(w, r, http.StatusBadRequest, "Filter pekerja tidak valid")
			return
		}
	}
	workers, err := h.svc.Pekerja(r.Context())
	if err != nil {
		log.Printf("laporan presensi pekerja: %v", err)
		h.d.Render.Error(w, r, http.StatusInternalServerError, "Gagal memuat filter pekerja")
		return
	}
	_, _, items, err := h.svc.Presensi(r.Context(), mulai, selesai, pekerjaID)
	if err != nil {
		if errors.Is(err, service.ErrInputTidakValid) {
			h.d.Render.Error(w, r, http.StatusBadRequest, "Rentang tanggal tidak valid")
			return
		}
		log.Printf("laporan presensi: %v", err)
		h.d.Render.Error(w, r, http.StatusInternalServerError, "Gagal memuat laporan presensi")
		return
	}
	h.d.Render.Page(w, r, "laporan/presensi.html", PageData{
		Title: "Laporan Presensi", Menu: "laporan-presensi",
		Data: laporanPresensiData{Mulai: mulai, Selesai: selesai, PekerjaID: pekerjaID, Pekerja: workers, Presensi: items},
	})
}

func (h *LaporanPresensiKasbonHandler) Kasbon(w http.ResponseWriter, r *http.Request) {
	items, err := h.svc.RekapKasbon(r.Context())
	if err != nil {
		log.Printf("laporan kasbon: %v", err)
		h.d.Render.Error(w, r, http.StatusInternalServerError, "Gagal memuat rekap kasbon")
		return
	}
	h.d.Render.Page(w, r, "laporan/kasbon.html", PageData{
		Title: "Rekap Kasbon", Menu: "laporan-kasbon",
		Data: struct {
			Rekap []repository.RekapKasbon
		}{Rekap: items},
	})
}
