package handler

import (
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"penggajian/internal/middleware"
	"penggajian/internal/repository"
	"penggajian/internal/service"
)

type KasbonHandler struct {
	d   Deps
	svc *service.KasbonService
}

func NewKasbonHandler(d Deps) *KasbonHandler {
	return &KasbonHandler{d: d, svc: service.NewKasbonService(repository.NewKasbonRepository(d.DB))}
}

func (h *KasbonHandler) Register(mux *http.ServeMux, g Guards) {
	mux.Handle("GET /admin/kasbon", g.Admin(http.HandlerFunc(h.List)))
	mux.Handle("GET /pekerja/kasbon", g.Pekerja(http.HandlerFunc(h.Saya)))
	mux.Handle("GET /pekerja/kasbon/baru", g.Pekerja(http.HandlerFunc(h.Form)))
	mux.Handle("POST /pekerja/kasbon", g.Pekerja(http.HandlerFunc(h.Simpan)))
}

func (h *KasbonHandler) List(w http.ResponseWriter, r *http.Request) {
	items, err := h.svc.Daftar(r.Context())
	if err != nil {
		log.Printf("daftar kasbon: %v", err)
		h.d.Render.Error(w, r, http.StatusInternalServerError, "Gagal memuat data kasbon")
		return
	}
	h.d.Render.Page(w, r, "kasbon/list.html", PageData{
		Title: "Kasbon", Menu: "kasbon",
		Data: struct {
			Kasbon []repository.KasbonRecord
		}{Kasbon: items},
	})
}

func (h *KasbonHandler) Saya(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFrom(r.Context())
	items, err := h.svc.MilikPekerja(r.Context(), user.PekerjaID)
	if err != nil {
		if errors.Is(err, service.ErrBukanMilikSendiri) {
			h.d.Render.Error(w, r, http.StatusForbidden, "Akun pekerja tidak valid")
			return
		}
		log.Printf("kasbon pekerja: %v", err)
		h.d.Render.Error(w, r, http.StatusInternalServerError, "Gagal memuat riwayat kasbon")
		return
	}
	flash := ""
	if r.URL.Query().Get("sukses") == "1" {
		flash = "Pengajuan kasbon berhasil dicatat."
	}
	h.d.Render.Page(w, r, "kasbon/saya.html", PageData{
		Title: "Kasbon Saya", Menu: "kasbon", Flash: flash,
		Data: struct {
			Kasbon []repository.KasbonRecord
		}{Kasbon: items},
	})
}

func (h *KasbonHandler) Form(w http.ResponseWriter, r *http.Request) {
	h.renderForm(w, r, "", "", "")
}

func (h *KasbonHandler) Simpan(w http.ResponseWriter, r *http.Request) {
	if !sameOriginRequest(r) {
		h.d.Render.Error(w, r, http.StatusForbidden, "Permintaan lintas situs ditolak")
		return
	}
	if err := parseLimitedForm(w, r); err != nil {
		h.d.Render.Error(w, r, http.StatusBadRequest, "Form kasbon tidak valid")
		return
	}
	nominal := strings.TrimSpace(r.PostFormValue("nominal"))
	keterangan := strings.TrimSpace(r.PostFormValue("keterangan"))
	user := middleware.UserFrom(r.Context())
	err := h.svc.Buat(r.Context(), user.PekerjaID, nominal, keterangan, time.Now())
	switch {
	case err == nil:
		http.Redirect(w, r, "/pekerja/kasbon?sukses=1", http.StatusSeeOther)
	case errors.Is(err, service.ErrInputTidakValid):
		h.renderForm(w, r, "Nominal harus lebih dari nol, maksimal dua angka desimal, dan keterangan maksimal 2.000 karakter.", nominal, keterangan)
	case errors.Is(err, service.ErrBukanMilikSendiri):
		h.d.Render.Error(w, r, http.StatusForbidden, "Akun pekerja tidak valid")
	default:
		log.Printf("simpan kasbon: %v", err)
		h.d.Render.Error(w, r, http.StatusInternalServerError, "Gagal mencatat pengajuan kasbon")
	}
}

func (h *KasbonHandler) renderForm(w http.ResponseWriter, r *http.Request, pageError, nominal, keterangan string) {
	h.d.Render.Page(w, r, "kasbon/form.html", PageData{
		Title: "Ajukan Kasbon", Menu: "kasbon", Error: pageError,
		Data: struct {
			Nominal    string
			Keterangan string
		}{Nominal: nominal, Keterangan: keterangan},
	})
}
