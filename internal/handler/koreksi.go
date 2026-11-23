package handler

import (
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"

	"penggajian/internal/repository"
	"penggajian/internal/service"
)

// Pemilik: Anggota 3 | F-10 · UC-14 Koreksi Gaji (BR-11)
// Tabel: koreksi_gaji (tulis)
type KoreksiHandler struct {
	d   Deps
	svc *service.KoreksiService
}

type koreksiFormInput struct {
	GajiSumberID    int64
	PeriodeTujuanID int64
	JenisKoreksi    string
	Nominal         string
	Alasan          string
}

type koreksiFormData struct {
	Form           koreksiFormInput
	OpsiGajiSumber []repository.OpsiGajiSumber
	PeriodeTerbuka []repository.OpsiPeriodeTujuan
}

func NewKoreksiHandler(d Deps) *KoreksiHandler {
	return &KoreksiHandler{d: d, svc: service.NewKoreksiService(repository.NewKoreksiRepository(d.DB))}
}

func (h *KoreksiHandler) Register(mux *http.ServeMux, g Guards) {
	mux.Handle("GET /admin/koreksi", g.Admin(http.HandlerFunc(h.List)))
	mux.Handle("GET /admin/koreksi/baru", g.Admin(http.HandlerFunc(h.Form)))
	mux.Handle("POST /admin/koreksi", g.Admin(http.HandlerFunc(h.Simpan)))
}

func (h *KoreksiHandler) List(w http.ResponseWriter, r *http.Request) {
	items, err := h.svc.Daftar(r.Context())
	if err != nil {
		log.Printf("daftar koreksi: %v", err)
		h.d.Render.Error(w, r, http.StatusInternalServerError, "Gagal memuat daftar koreksi gaji")
		return
	}

	flash := ""
	if r.URL.Query().Get("sukses") == "1" {
		flash = "Koreksi gaji berhasil disimpan ke periode tujuan."
	}

	h.d.Render.Page(w, r, "koreksi/list.html", PageData{
		Title: "Koreksi Gaji",
		Menu:  "koreksi",
		Flash: flash,
		Data: struct {
			Koreksi []repository.KoreksiRecord
		}{Koreksi: items},
	})
}

func (h *KoreksiHandler) Form(w http.ResponseWriter, r *http.Request) {
	h.renderForm(w, r, "", koreksiFormInput{})
}

func (h *KoreksiHandler) Simpan(w http.ResponseWriter, r *http.Request) {
	if !sameOriginRequest(r) {
		h.d.Render.Error(w, r, http.StatusForbidden, "Permintaan lintas situs ditolak")
		return
	}
	if err := parseLimitedForm(w, r); err != nil {
		h.d.Render.Error(w, r, http.StatusBadRequest, "Form koreksi tidak valid")
		return
	}

	gajiSumberID, _ := strconv.ParseInt(strings.TrimSpace(r.PostFormValue("gaji_sumber_id")), 10, 64)
	periodeTujuanID, _ := strconv.ParseInt(strings.TrimSpace(r.PostFormValue("periode_gaji_tujuan_id")), 10, 64)

	form := koreksiFormInput{
		GajiSumberID:    gajiSumberID,
		PeriodeTujuanID: periodeTujuanID,
		JenisKoreksi:    strings.TrimSpace(r.PostFormValue("jenis_koreksi")),
		Nominal:         strings.TrimSpace(r.PostFormValue("nominal")),
		Alasan:          strings.TrimSpace(r.PostFormValue("alasan")),
	}

	_, err := h.svc.Buat(r.Context(), service.InputKoreksi{
		GajiSumberID:    form.GajiSumberID,
		PeriodeTujuanID: form.PeriodeTujuanID,
		JenisKoreksi:    form.JenisKoreksi,
		NominalText:     form.Nominal,
		Alasan:          form.Alasan,
	})
	switch {
	case err == nil:
		http.Redirect(w, r, "/admin/koreksi?sukses=1", http.StatusSeeOther)
	case errors.Is(err, service.ErrInputTidakValid):
		h.renderForm(w, r, "Input tidak valid. Pastikan gaji sumber, periode tujuan yang sesuai kategori pekerja, jenis (tambah/kurang), nominal positif (maks 2 desimal), dan alasan terisi.", form)
	case errors.Is(err, service.ErrPeriodeDitutup):
		h.renderForm(w, r, "Periode tujuan sudah ditutup; pilih periode terbuka.", form)
	case errors.Is(err, service.ErrTidakDitemukan):
		h.renderForm(w, r, "Gaji sumber atau periode tujuan tidak ditemukan.", form)
	default:
		log.Printf("simpan koreksi: %v", err)
		h.d.Render.Error(w, r, http.StatusInternalServerError, "Gagal menyimpan koreksi gaji")
	}
}

func (h *KoreksiHandler) renderForm(w http.ResponseWriter, r *http.Request, pageError string, form koreksiFormInput) {
	data, err := h.svc.DataForm(r.Context())
	if err != nil {
		log.Printf("form koreksi: %v", err)
		h.d.Render.Error(w, r, http.StatusInternalServerError, "Gagal memuat data form koreksi gaji")
		return
	}

	h.d.Render.Page(w, r, "koreksi/form.html", PageData{
		Title: "Tambah Koreksi Gaji",
		Menu:  "koreksi",
		Error: pageError,
		Data: koreksiFormData{
			Form:           form,
			OpsiGajiSumber: data.OpsiGajiSumber,
			PeriodeTerbuka: data.PeriodeTerbuka,
		},
	})
}
