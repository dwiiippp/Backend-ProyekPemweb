package handler

import (
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"penggajian/internal/model"
	"penggajian/internal/repository"
	"penggajian/internal/service"
)

// Pemilik: Anggota 3 | Periode Gaji (BR-01, BR-02, BR-10)
// Tabel: periode_gaji (tulis)
type PeriodeHandler struct {
	d   Deps
	svc *service.PeriodeService
}

type periodeListItem struct {
	ID                int64
	JenisPeriode      string
	TanggalMulai      time.Time
	TanggalSelesai    time.Time
	TanggalPembayaran time.Time
	Status            string
	Terbuka           bool
}

type periodeFormData struct {
	JenisPeriode      string
	TanggalMulai      string
	TanggalSelesai    string
	TanggalPembayaran string
}

type periodeDetailData struct {
	Detail            service.DetailPeriode
	TanggalPembayaran time.Time
	Terbuka           bool
}

func NewPeriodeHandler(d Deps) *PeriodeHandler {
	return &PeriodeHandler{
		d:   d,
		svc: service.NewPeriodeService(repository.NewPeriodeRepository(d.DB)),
	}
}

func (h *PeriodeHandler) Register(mux *http.ServeMux, g Guards) {
	mux.Handle("GET /admin/periode", g.Admin(http.HandlerFunc(h.List)))
	mux.Handle("GET /admin/periode/baru", g.Admin(http.HandlerFunc(h.Form)))
	mux.Handle("POST /admin/periode", g.Admin(http.HandlerFunc(h.Simpan)))
	mux.Handle("GET /admin/periode/{id}", g.Admin(http.HandlerFunc(h.Detail)))
}

func (h *PeriodeHandler) List(w http.ResponseWriter, r *http.Request) {
	items, err := h.svc.Daftar(r.Context())
	if err != nil {
		log.Printf("daftar periode: %v", err)
		h.d.Render.Error(w, r, http.StatusInternalServerError, "Gagal memuat data periode gaji")
		return
	}

	viewItems := make([]periodeListItem, 0, len(items))
	for _, p := range items {
		var tglBayar time.Time
		if p.TanggalPembayaran != nil {
			tglBayar = *p.TanggalPembayaran
		}
		viewItems = append(viewItems, periodeListItem{
			ID:                p.ID,
			JenisPeriode:      p.JenisPeriode,
			TanggalMulai:      p.TanggalMulai,
			TanggalSelesai:    p.TanggalSelesai,
			TanggalPembayaran: tglBayar,
			Status:            p.Status,
			Terbuka:           p.Status == model.PeriodeTerbuka,
		})
	}

	flash := ""
	if r.URL.Query().Get("sukses") == "1" {
		flash = "Periode gaji baru berhasil dibuat."
	}

	h.d.Render.Page(w, r, "periode/list.html", PageData{
		Title: "Periode Gaji",
		Menu:  "periode",
		Flash: flash,
		Data: struct {
			Periode []periodeListItem
		}{Periode: viewItems},
	})
}

func (h *PeriodeHandler) Form(w http.ResponseWriter, r *http.Request) {
	h.renderForm(w, r, "", periodeFormData{})
}

func (h *PeriodeHandler) Simpan(w http.ResponseWriter, r *http.Request) {
	if !sameOriginRequest(r) {
		h.d.Render.Error(w, r, http.StatusForbidden, "Permintaan lintas situs ditolak")
		return
	}

	if err := parseLimitedForm(w, r); err != nil {
		h.d.Render.Error(w, r, http.StatusBadRequest, "Form periode tidak valid")
		return
	}

	form := periodeFormData{
		JenisPeriode:      strings.TrimSpace(r.PostFormValue("jenis_periode")),
		TanggalMulai:      strings.TrimSpace(r.PostFormValue("tanggal_mulai")),
		TanggalSelesai:    strings.TrimSpace(r.PostFormValue("tanggal_selesai")),
		TanggalPembayaran: strings.TrimSpace(r.PostFormValue("tanggal_pembayaran")),
	}

	mulai, errMulai := time.Parse("2006-01-02", form.TanggalMulai)
	selesai, errSelesai := time.Parse("2006-01-02", form.TanggalSelesai)
	if errMulai != nil || errSelesai != nil {
		h.renderForm(w, r, "Tanggal mulai dan selesai harus diisi dengan format yang benar.", form)
		return
	}

	var tglBayarPtr *time.Time
	if form.TanggalPembayaran != "" {
		tglBayar, errBayar := time.Parse("2006-01-02", form.TanggalPembayaran)
		if errBayar != nil {
			h.renderForm(w, r, "Tanggal pembayaran tidak valid.", form)
			return
		}
		tglBayarPtr = &tglBayar
	}

	periode := model.PeriodeGaji{
		JenisPeriode:      form.JenisPeriode,
		TanggalMulai:      mulai,
		TanggalSelesai:    selesai,
		TanggalPembayaran: tglBayarPtr,
		Status:            model.PeriodeTerbuka,
	}

	_, err := h.svc.Buat(r.Context(), periode)
	switch {
	case err == nil:
		http.Redirect(w, r, "/admin/periode?sukses=1", http.StatusSeeOther)
	case errors.Is(err, service.ErrInputTidakValid):
		h.renderForm(w, r, "Jenis atau rentang tanggal periode tidak valid. Mingguan wajib Senin–Sabtu (6 hari); bulanan wajib tanggal 1 s.d. akhir bulan; tanggal pembayaran minimal sama dengan tanggal selesai.", form)
	case errors.Is(err, service.ErrPeriodeDuplikat):
		h.renderForm(w, r, "Periode dengan jenis dan rentang tanggal tersebut sudah terdaftar.", form)
	default:
		log.Printf("simpan periode: %v", err)
		h.d.Render.Error(w, r, http.StatusInternalServerError, "Gagal menyimpan periode gaji")
	}
}

func (h *PeriodeHandler) Detail(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(strings.TrimSpace(r.PathValue("id")), 10, 64)
	if err != nil || id <= 0 {
		h.d.Render.Error(w, r, http.StatusBadRequest, "ID periode tidak valid")
		return
	}

	detail, err := h.svc.Detail(r.Context(), id)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInputTidakValid):
			h.d.Render.Error(w, r, http.StatusBadRequest, "ID periode tidak valid")
		case errors.Is(err, service.ErrTidakDitemukan):
			h.d.Render.Error(w, r, http.StatusNotFound, "Periode gaji tidak ditemukan")
		default:
			log.Printf("detail periode: %v", err)
			h.d.Render.Error(w, r, http.StatusInternalServerError, "Gagal memuat detail periode gaji")
		}
		return
	}

	var tglBayar time.Time
	if detail.Periode.TanggalPembayaran != nil {
		tglBayar = *detail.Periode.TanggalPembayaran
	}

	h.d.Render.Page(w, r, "periode/detail.html", PageData{
		Title: "Detail Periode Gaji",
		Menu:  "periode",
		Data: periodeDetailData{
			Detail:            detail,
			TanggalPembayaran: tglBayar,
			Terbuka:           detail.Periode.Status == model.PeriodeTerbuka,
		},
	})
}

func (h *PeriodeHandler) renderForm(w http.ResponseWriter, r *http.Request, pageError string, form periodeFormData) {
	h.d.Render.Page(w, r, "periode/form.html", PageData{
		Title: "Buat Periode Gaji",
		Menu:  "periode",
		Error: pageError,
		Data:  form,
	})
}
