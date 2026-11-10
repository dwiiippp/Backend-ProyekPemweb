package handler

import (
	"log"
	"net/http"

	"penggajian/internal/middleware"
	"penggajian/internal/repository"
	"penggajian/internal/service"
)

// Tabel: pekerja (baca)
type HomeHandler struct {
	d   Deps
	svc *service.HomeService
}

func NewHomeHandler(d Deps) *HomeHandler {
	return &HomeHandler{d: d, svc: service.NewHomeService(repository.NewHomeRepository(d.DB))}
}

func (h *HomeHandler) Register(mux *http.ServeMux, g Guards) {
	mux.Handle("GET /{$}", g.Login(http.HandlerFunc(h.Index)))
	mux.Handle("GET /profil", g.Pekerja(http.HandlerFunc(h.Profil)))
	// TODO(Anggota 1): tambah route lain di sini. Dashboard per role dan halaman profil pekerja (NIK, nama, kategori kerja).
}

func (h *HomeHandler) Index(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFrom(r.Context())

	if user == nil {
		h.d.Render.Error(w, r, http.StatusUnauthorized, "Silakan login terlebih dahulu")
		return
	}

	data := make(map[string]interface{})

	if user.IsAdmin() {
		summary, err := h.svc.GetAdminSummary(r.Context())
		if err != nil {
			log.Printf("dashboard admin: %v", err)
			h.d.Render.Error(w, r, http.StatusInternalServerError, "Gagal mengambil data dashboard")
			return
		}

		data["PresensiMenunggu"] = summary.PresensiMenunggu
		data["PeriodeTerbuka"] = summary.PeriodeTerbuka
	} else if user.IsPekerja() {
		if user.PekerjaID <= 0 {
			h.d.Render.Error(w, r, http.StatusForbidden, "Data pekerja tidak ditemukan")
			return
		}

		jadwal, err := h.svc.GetTodaySchedule(r.Context(), user.PekerjaID)
		if err != nil {
			log.Printf("dashboard pekerja: %v", err)
			h.d.Render.Error(w, r, http.StatusInternalServerError, "Gagal mengambil jadwal hari ini")
			return
		}

		data["JadwalHariIni"] = jadwal
	}

	h.d.Render.Page(w, r, "home.html", PageData{
		Title: "Dashboard",
		Menu:  "home",
		Data:  data,
	})
}

func (h *HomeHandler) Profil(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFrom(r.Context())

	if user == nil || user.PekerjaID <= 0 {
		h.d.Render.Error(w, r, http.StatusForbidden, "Data pekerja tidak ditemukan")
		return
	}

	profil, err := h.svc.GetProfile(r.Context(), user.PekerjaID)
	if err != nil {
		log.Printf("profil pekerja: %v", err)
		h.d.Render.Error(w, r, http.StatusInternalServerError, "Gagal memgambil data profil")
		return
	}

	if profil == nil {
		h.d.Render.Error(w, r, http.StatusNotFound, "Profil pekerja tidak ditemukan")
		return
	}

	h.d.Render.Page(w, r, "profil.html", PageData{
		Title: "Profil Saya",
		Menu:  "profil",
		Data:  profil,
	})
}
