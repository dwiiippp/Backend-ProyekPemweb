package handler

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"penggajian/internal/model"
	"penggajian/internal/repository"
	"penggajian/internal/service"
)

// Pemilik: Anggota 3 | F-07 · F-09 · UC-12 Proses Penggajian · UC-13 Tutup Periode
// Tabel: gaji, potongan_kasbon_gaji (tulis); kasbon, periode_gaji (ubah); baca absensi, riwayat_gaji, pekerja, koreksi_gaji
type PenggajianHandler struct {
	d   Deps
	svc *service.PenggajianService
}

type penggajianPeriodeItem struct {
	ID                int64
	JenisPeriode      string
	TanggalMulai      time.Time
	TanggalSelesai    time.Time
	TanggalPembayaran time.Time
	Status            string
	Terbuka           bool
}

type penggajianPratinjauData struct {
	Pratinjau                service.HasilPratinjau
	TanggalPembayaranDefault string
}

type penggajianHasilData struct {
	Hasil             service.HasilPeriodeGaji
	TanggalPembayaran time.Time
}

func NewPenggajianHandler(d Deps) *PenggajianHandler {
	return &PenggajianHandler{d: d, svc: service.NewPenggajianService(repository.NewPenggajianRepository(d.DB))}
}

func (h *PenggajianHandler) Register(mux *http.ServeMux, g Guards) {
	mux.Handle("GET /admin/penggajian", g.Admin(http.HandlerFunc(h.Index)))
	mux.Handle("GET /admin/penggajian/{id}/pratinjau", g.Admin(http.HandlerFunc(h.Pratinjau)))
	mux.Handle("POST /admin/penggajian/{id}/konfirmasi", g.Admin(http.HandlerFunc(h.Konfirmasi)))
	mux.Handle("GET /admin/penggajian/{id}/hasil", g.Admin(http.HandlerFunc(h.Hasil)))
}

func (h *PenggajianHandler) Index(w http.ResponseWriter, r *http.Request) {
	items, err := h.svc.DaftarPeriode(r.Context())
	if err != nil {
		log.Printf("daftar periode penggajian: %v", err)
		h.d.Render.Error(w, r, http.StatusInternalServerError, "Gagal memuat daftar periode penggajian")
		return
	}

	viewItems := make([]penggajianPeriodeItem, 0, len(items))
	for _, p := range items {
		var tglBayar time.Time
		if p.TanggalPembayaran != nil {
			tglBayar = *p.TanggalPembayaran
		}
		viewItems = append(viewItems, penggajianPeriodeItem{
			ID:                p.ID,
			JenisPeriode:      p.JenisPeriode,
			TanggalMulai:      p.TanggalMulai,
			TanggalSelesai:    p.TanggalSelesai,
			TanggalPembayaran: tglBayar,
			Status:            p.Status,
			Terbuka:           p.Status == model.PeriodeTerbuka,
		})
	}

	h.d.Render.Page(w, r, "penggajian/index.html", PageData{
		Title: "Proses Penggajian",
		Menu:  "penggajian",
		Data: struct {
			Periode []penggajianPeriodeItem
		}{Periode: viewItems},
	})
}

func (h *PenggajianHandler) Pratinjau(w http.ResponseWriter, r *http.Request) {
	periodeID, ok := parseIDParam(r.PathValue("id"))
	if !ok {
		h.d.Render.Error(w, r, http.StatusBadRequest, "ID periode tidak valid")
		return
	}

	pratinjau, err := h.svc.Pratinjau(r.Context(), periodeID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrTidakDitemukan):
			h.d.Render.Error(w, r, http.StatusNotFound, "Periode gaji tidak ditemukan")
		case errors.Is(err, service.ErrPeriodeDitutup):
			http.Redirect(w, r, fmt.Sprintf("/admin/penggajian/%d/hasil", periodeID), http.StatusSeeOther)
		default:
			log.Printf("pratinjau penggajian: %v", err)
			h.d.Render.Error(w, r, http.StatusInternalServerError, "Gagal memproses pratinjau penggajian")
		}
		return
	}

	tglDefault := time.Now().Format("2006-01-02")
	if pratinjau.Periode.TanggalPembayaran != nil && !pratinjau.Periode.TanggalPembayaran.IsZero() {
		tglDefault = pratinjau.Periode.TanggalPembayaran.Format("2006-01-02")
	}

	h.d.Render.Page(w, r, "penggajian/pratinjau.html", PageData{
		Title: "Pratinjau Penggajian",
		Menu:  "penggajian",
		Data: penggajianPratinjauData{
			Pratinjau:                pratinjau,
			TanggalPembayaranDefault: tglDefault,
		},
	})
}

func (h *PenggajianHandler) Konfirmasi(w http.ResponseWriter, r *http.Request) {
	if !sameOriginRequest(r) {
		h.d.Render.Error(w, r, http.StatusForbidden, "Permintaan lintas situs ditolak")
		return
	}
	if err := parseLimitedForm(w, r); err != nil {
		h.d.Render.Error(w, r, http.StatusBadRequest, "Form konfirmasi tidak valid")
		return
	}

	periodeID, ok := parseIDParam(r.PathValue("id"))
	if !ok {
		h.d.Render.Error(w, r, http.StatusBadRequest, "ID periode tidak valid")
		return
	}

	var tglBayar time.Time
	if raw := strings.TrimSpace(r.PostFormValue("tanggal_pembayaran")); raw != "" {
		parsed, err := time.Parse("2006-01-02", raw)
		if err != nil {
			h.d.Render.Error(w, r, http.StatusBadRequest, "Format tanggal pembayaran tidak valid")
			return
		}
		tglBayar = parsed
	}

	err := h.svc.KonfirmasiDanTutup(r.Context(), periodeID, tglBayar)
	switch {
	case err == nil:
		http.Redirect(w, r, fmt.Sprintf("/admin/penggajian/%d/hasil?sukses=1", periodeID), http.StatusSeeOther)
	case errors.Is(err, service.ErrTidakDitemukan):
		h.d.Render.Error(w, r, http.StatusNotFound, "Periode gaji tidak ditemukan")
	case errors.Is(err, service.ErrPeriodeDitutup):
		http.Redirect(w, r, fmt.Sprintf("/admin/penggajian/%d/hasil", periodeID), http.StatusSeeOther)
	case errors.Is(err, service.ErrPrasyaratPenggajian):
		pratinjau, pErr := h.svc.Pratinjau(r.Context(), periodeID)
		if pErr != nil {
			h.d.Render.Error(w, r, http.StatusConflict, "Prasyarat penggajian belum terpenuhi")
			return
		}
		h.d.Render.Page(w, r, "penggajian/pratinjau.html", PageData{
			Title: "Pratinjau Penggajian",
			Menu:  "penggajian",
			Error: "Periode belum dapat ditutup. Selesaikan semua kesalahan prasyarat di bawah ini.",
			Data: penggajianPratinjauData{
				Pratinjau:                pratinjau,
				TanggalPembayaranDefault: r.PostFormValue("tanggal_pembayaran"),
			},
		})
	default:
		log.Printf("konfirmasi penggajian: %v", err)
		h.d.Render.Error(w, r, http.StatusInternalServerError, "Gagal mengonfirmasi dan menutup periode penggajian")
	}
}

func (h *PenggajianHandler) Hasil(w http.ResponseWriter, r *http.Request) {
	periodeID, ok := parseIDParam(r.PathValue("id"))
	if !ok {
		h.d.Render.Error(w, r, http.StatusBadRequest, "ID periode tidak valid")
		return
	}

	hasil, err := h.svc.Hasil(r.Context(), periodeID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrTidakDitemukan):
			h.d.Render.Error(w, r, http.StatusNotFound, "Periode gaji tidak ditemukan")
		default:
			log.Printf("hasil penggajian: %v", err)
			h.d.Render.Error(w, r, http.StatusInternalServerError, "Gagal memuat hasil penggajian")
		}
		return
	}

	var tglBayar time.Time
	if hasil.Periode.TanggalPembayaran != nil {
		tglBayar = *hasil.Periode.TanggalPembayaran
	}

	flash := ""
	if r.URL.Query().Get("sukses") == "1" {
		flash = "Penggajian berhasil dikonfirmasi, kasbon diperbarui, dan periode telah ditutup."
	}

	h.d.Render.Page(w, r, "penggajian/hasil.html", PageData{
		Title: "Hasil Penggajian",
		Menu:  "penggajian",
		Flash: flash,
		Data: penggajianHasilData{
			Hasil:             hasil,
			TanggalPembayaran: tglBayar,
		},
	})
}

func parseIDParam(raw string) (int64, bool) {
	id, err := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
	return id, err == nil && id > 0
}
