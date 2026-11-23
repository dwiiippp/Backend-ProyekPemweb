package handler

import (
	"errors"
	"log"
	"net/http"
	"strconv"

	"penggajian/internal/middleware"
	"penggajian/internal/repository"
	"penggajian/internal/service"
)

// Pemilik: Anggota 3 | F-11 · UC-06 Lihat Slip · UC-15 Cetak Slip (BR-12)
// Tabel: gaji, potongan_kasbon_gaji (baca)
type SlipHandler struct {
	d   Deps
	svc *service.SlipService
}

type slipListAdminData struct {
	PeriodeID int64
	PekerjaID int64
	Periode   []repository.OpsiPeriodeSlip
	Pekerja   []repository.OpsiPekerjaSlip
	Slip      []repository.RingkasanSlip
}

type slipDetailData struct {
	Slip    repository.DetailSlip
	BackURL string
	IsAdmin bool
}

func NewSlipHandler(d Deps) *SlipHandler {
	return &SlipHandler{d: d, svc: service.NewSlipService(repository.NewSlipRepository(d.DB))}
}

func (h *SlipHandler) Register(mux *http.ServeMux, g Guards) {
	mux.Handle("GET /admin/slip", g.Admin(http.HandlerFunc(h.List)))
	mux.Handle("GET /admin/slip/{id}", g.Admin(http.HandlerFunc(h.DetailAdmin)))
	mux.Handle("GET /pekerja/slip", g.Pekerja(http.HandlerFunc(h.Saya)))
	mux.Handle("GET /pekerja/slip/{id}", g.Pekerja(http.HandlerFunc(h.DetailPekerja)))
}

func (h *SlipHandler) List(w http.ResponseWriter, r *http.Request) {
	var periodeID, pekerjaID int64
	if raw := r.URL.Query().Get("periode_id"); raw != "" {
		v, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || v < 0 {
			h.d.Render.Error(w, r, http.StatusBadRequest, "Filter periode tidak valid")
			return
		}
		periodeID = v
	}
	if raw := r.URL.Query().Get("pekerja_id"); raw != "" {
		v, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || v < 0 {
			h.d.Render.Error(w, r, http.StatusBadRequest, "Filter pekerja tidak valid")
			return
		}
		pekerjaID = v
	}

	opts, err := h.svc.FilterOptions(r.Context())
	if err != nil {
		log.Printf("opsi filter slip: %v", err)
		h.d.Render.Error(w, r, http.StatusInternalServerError, "Gagal memuat opsi filter slip gaji")
		return
	}

	items, err := h.svc.DaftarAdmin(r.Context(), periodeID, pekerjaID)
	if err != nil {
		log.Printf("daftar slip admin: %v", err)
		h.d.Render.Error(w, r, http.StatusInternalServerError, "Gagal memuat daftar slip gaji")
		return
	}

	h.d.Render.Page(w, r, "slip/list.html", PageData{
		Title: "Slip Gaji",
		Menu:  "slip",
		Data: slipListAdminData{
			PeriodeID: periodeID,
			PekerjaID: pekerjaID,
			Periode:   opts.Periode,
			Pekerja:   opts.Pekerja,
			Slip:      items,
		},
	})
}

func (h *SlipHandler) DetailAdmin(w http.ResponseWriter, r *http.Request) {
	gajiID, ok := parseIDParam(r.PathValue("id"))
	if !ok {
		h.d.Render.Error(w, r, http.StatusBadRequest, "ID slip gaji tidak valid")
		return
	}

	slip, err := h.svc.DetailAdmin(r.Context(), gajiID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrTidakDitemukan):
			h.d.Render.Error(w, r, http.StatusNotFound, "Slip gaji tidak ditemukan")
		default:
			log.Printf("detail slip admin: %v", err)
			h.d.Render.Error(w, r, http.StatusInternalServerError, "Gagal memuat detail slip gaji")
		}
		return
	}

	h.d.Render.Page(w, r, "slip/detail.html", PageData{
		Title: "Slip Gaji — " + slip.NamaPekerja,
		Menu:  "slip",
		Data: slipDetailData{
			Slip:    slip,
			BackURL: "/admin/slip",
			IsAdmin: true,
		},
	})
}

func (h *SlipHandler) Saya(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFrom(r.Context())
	items, err := h.svc.DaftarMilikPekerja(r.Context(), user.PekerjaID)
	if err != nil {
		if errors.Is(err, service.ErrBukanMilikSendiri) {
			h.d.Render.Error(w, r, http.StatusForbidden, "Akun pekerja tidak valid")
			return
		}
		log.Printf("daftar slip pekerja: %v", err)
		h.d.Render.Error(w, r, http.StatusInternalServerError, "Gagal memuat daftar slip gaji Anda")
		return
	}

	h.d.Render.Page(w, r, "slip/saya.html", PageData{
		Title: "Slip Gaji Saya",
		Menu:  "slip",
		Data: struct {
			Slip []repository.RingkasanSlip
		}{Slip: items},
	})
}

func (h *SlipHandler) DetailPekerja(w http.ResponseWriter, r *http.Request) {
	gajiID, ok := parseIDParam(r.PathValue("id"))
	if !ok {
		h.d.Render.Error(w, r, http.StatusBadRequest, "ID slip gaji tidak valid")
		return
	}

	user := middleware.UserFrom(r.Context())
	slip, err := h.svc.DetailPekerja(r.Context(), gajiID, user.PekerjaID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrBukanMilikSendiri):
			h.d.Render.Error(w, r, http.StatusForbidden, "Anda tidak berhak melihat slip gaji pekerja lain")
		case errors.Is(err, service.ErrTidakDitemukan):
			h.d.Render.Error(w, r, http.StatusNotFound, "Slip gaji tidak ditemukan")
		default:
			log.Printf("detail slip pekerja: %v", err)
			h.d.Render.Error(w, r, http.StatusInternalServerError, "Gagal memuat detail slip gaji")
		}
		return
	}

	h.d.Render.Page(w, r, "slip/detail.html", PageData{
		Title: "Slip Gaji Saya",
		Menu:  "slip",
		Data: slipDetailData{
			Slip:    slip,
			BackURL: "/pekerja/slip",
			IsAdmin: false,
		},
	})
}
