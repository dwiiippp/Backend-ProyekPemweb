package handler

import (
	"errors"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"penggajian/internal/middleware"
	"penggajian/internal/repository"
	"penggajian/internal/service"
)

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
}

func (h *AbsensiHandler) Form(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFrom(r.Context())
	items, err := h.svc.JadwalTerbuka(r.Context(), user.PekerjaID, time.Now())
	if err != nil {
		log.Printf("absensi jadwal: %v", err)
		h.d.Render.Error(w, r, http.StatusInternalServerError, "Gagal memuat jadwal presensi")
		return
	}
	flash := ""
	if r.URL.Query().Get("sukses") == "1" {
		flash = "Presensi berhasil dikirim dan menunggu validasi admin."
	}
	h.d.Render.Page(w, r, "absensi/form.html", PageData{
		Title: "Input Presensi", Menu: "absensi", Flash: flash,
		Data: struct {
			Jadwal []repository.JadwalPresensi
		}{Jadwal: items},
	})
}

func (h *AbsensiHandler) Simpan(w http.ResponseWriter, r *http.Request) {
	if !sameOriginRequest(r) {
		h.d.Render.Error(w, r, http.StatusForbidden, "Permintaan lintas situs ditolak")
		return
	}
	if err := parseLimitedForm(w, r); err != nil {
		h.d.Render.Error(w, r, http.StatusBadRequest, "Form presensi tidak valid")
		return
	}
	jadwalID, err := strconv.ParseInt(strings.TrimSpace(r.PostFormValue("jadwal_id")), 10, 64)
	if err != nil || jadwalID <= 0 {
		h.d.Render.Error(w, r, http.StatusBadRequest, "Pilih jadwal yang valid")
		return
	}
	user := middleware.UserFrom(r.Context())
	err = h.svc.Simpan(r.Context(), user.PekerjaID, jadwalID, r.PostFormValue("jam_masuk"), r.PostFormValue("jam_keluar"))
	switch {
	case err == nil:
		http.Redirect(w, r, "/pekerja/absensi?sukses=1", http.StatusSeeOther)
	case errors.Is(err, service.ErrInputTidakValid):
		h.d.Render.Error(w, r, http.StatusBadRequest, "Jadwal atau jam presensi tidak valid. Jam keluar harus sama dengan atau sesudah jam selesai shift.")
	case errors.Is(err, service.ErrBukanMilikSendiri):
		h.d.Render.Error(w, r, http.StatusForbidden, "Jadwal tidak ditemukan atau bukan milik Anda")
	case errors.Is(err, service.ErrHariDitutup), errors.Is(err, service.ErrPeriodeDitutup), errors.Is(err, service.ErrPresensiGanda):
		h.d.Render.Error(w, r, http.StatusConflict, err.Error())
	default:
		log.Printf("simpan absensi: %v", err)
		h.d.Render.Error(w, r, http.StatusInternalServerError, "Gagal menyimpan presensi")
	}
}

func (h *AbsensiHandler) Riwayat(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFrom(r.Context())
	items, err := h.svc.Riwayat(r.Context(), user.PekerjaID)
	if err != nil {
		log.Printf("riwayat absensi: %v", err)
		h.d.Render.Error(w, r, http.StatusInternalServerError, "Gagal memuat riwayat presensi")
		return
	}
	h.d.Render.Page(w, r, "absensi/riwayat.html", PageData{
		Title: "Riwayat Presensi", Menu: "absensi",
		Data: struct {
			Riwayat []repository.RiwayatPresensi
		}{Riwayat: items},
	})
}

func sameOriginRequest(r *http.Request) bool {
	source := strings.TrimSpace(r.Header.Get("Origin"))
	isOrigin := source != ""
	if !isOrigin {
		source = strings.TrimSpace(r.Referer())
	}
	if source == "" {
		return true
	}

	u, err := url.Parse(source)
	if err != nil || u.Host == "" || u.User != nil {
		return false
	}
	if isOrigin && (u.Path != "" && u.Path != "/" || u.RawQuery != "" || u.Fragment != "") {
		return false
	}
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	return strings.EqualFold(u.Scheme, scheme) && strings.EqualFold(u.Host, r.Host)
}

func parseLimitedForm(w http.ResponseWriter, r *http.Request) error {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	return r.ParseForm()
}
