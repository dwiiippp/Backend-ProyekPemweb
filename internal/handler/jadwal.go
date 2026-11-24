package handler

import (
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

	"penggajian/internal/middleware"
	"penggajian/internal/repository"
	"penggajian/internal/service"
)

// Pemilik: Anggota 1 | F-05 Jadwal Kerja · UC-03
// Tabel: jadwal_kerja (tulis)
type JadwalHandler struct {
	d   Deps
	svc *service.JadwalService
}

func NewJadwalHandler(d Deps) *JadwalHandler {
	return &JadwalHandler{d: d, svc: service.NewJadwalService(repository.NewJadwalRepository(d.DB))}
}

func (h *JadwalHandler) Register(mux *http.ServeMux, g Guards) {
	mux.Handle("GET /admin/jadwal", g.Admin(http.HandlerFunc(h.Kalender)))
	mux.Handle("GET /pekerja/jadwal", g.Pekerja(http.HandlerFunc(h.Saya)))
	mux.Handle("GET /admin/jadwal/baru", g.Admin(http.HandlerFunc(h.NewForm)))
	mux.Handle("POST /admin/jadwal", g.Admin(http.HandlerFunc(h.Create)))
	mux.Handle("GET /admin/jadwal/{id}/edit", g.Admin(http.HandlerFunc(h.EditForm)))
	mux.Handle("POST /admin/jadwal/{id}/edit", g.Admin(http.HandlerFunc(h.Update)))
}

func (h *JadwalHandler) NewForm(w http.ResponseWriter, r *http.Request) {
	pekerja, err := h.svc.GetPekerjaAktif(r.Context())
	if err != nil {
		log.Printf("daftar pekerja jadwal: %v", err)
		h.d.Render.Error(w, r, http.StatusInternalServerError, "Gagal mengambil daftar pekerja")
		return
	}

	shifts, err := h.svc.GetShiftAktif(r.Context())
	if err != nil {
		log.Printf("daftar shift jadwal: %v", err)
		h.d.Render.Error(w, r, http.StatusInternalServerError, "Gagal mengambil daftar shift")
		return
	}

	h.d.Render.Page(w, r, "jadwal/form.html", PageData{
		Title: "Tambah Jadwal",
		Menu:  "jadwal",
		Data: map[string]interface{}{
			"Pekerja": pekerja,
			"Shifts":  shifts,
			"Jadwal":  repository.JadwalRecord{},
			"Tanggal": "",
			"Action":  "/admin/jadwal",
		},
	})
}

func (h *JadwalHandler) EditForm(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		h.d.Render.Error(w, r, http.StatusBadRequest, "ID jadwal tidak valid")
		return
	}

	jadwal, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrTidakDitemukan) {
			h.d.Render.Error(w, r, http.StatusNotFound, "Jadwal tidak ditemukan")
			return
		}

		log.Printf("edit form jadwal: %v", err)
		h.d.Render.Error(w, r, http.StatusInternalServerError, "Gagal mengambil data jadwal")
		return
	}

	pekerja, err := h.svc.GetPekerjaAktif(r.Context())
	if err != nil {
		log.Printf("daftar shift jadwal: %v", err)
		h.d.Render.Error(w, r, http.StatusInternalServerError, "Gagal mengambil data shift")
		return
	}

	shifts, err := h.svc.GetShiftAktif(r.Context())
	if err != nil {
		log.Printf("daftar shift jadwal: %v", err)
		h.d.Render.Error(w, r, http.StatusInternalServerError, "Gagal mengambil daftar shift")
		return
	}

	h.d.Render.Page(w, r, "jadwal/form.html", PageData{
		Title: "Edit Jadwal",
		Menu:  "jadwal",
		Data: map[string]interface{}{
			"Jadwal":  jadwal,
			"Pekerja": pekerja,
			"Shifts":  shifts,
			"Tanggal": jadwal.TanggalJadwal.Format("2006-01-02"),
			"Action":  "/admin/jadwal/" + strconv.FormatInt(id, 10) + "/edit",
		},
	})
}

func (h *JadwalHandler) Create(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		h.d.Render.Error(w, r, http.StatusBadRequest, "Form tidak valid")
		return
	}

	pekerjaID, errPekerja := strconv.ParseInt(r.FormValue("pekerja_id"), 10, 64)
	shiftID, errShift := strconv.ParseInt(r.FormValue("shift_id"), 10, 64)
	tanggal, errTanggal := time.Parse("2006-01-02", r.FormValue("tanggal_jadwal"))

	if errPekerja != nil || errShift != nil || errTanggal != nil {
		h.d.Render.Error(w, r, http.StatusBadRequest, "Data jadwal tidak valid")
		return
	}

	input := repository.CreateJadwalParams{
		PekerjaID:     pekerjaID,
		ShiftID:       shiftID,
		TanggalJadwal: tanggal,
	}

	err := h.svc.Create(r.Context(), input)
	if err != nil {
		if errors.Is(err, service.ErrInputTidakValid) {
			h.d.Render.Error(w, r, http.StatusBadRequest, "Data jadwal tidak valid. Periksa pekerja, shift, dan tanggal.")
			return
		}

		if errors.Is(err, repository.ErrJadwalDuplikat) {
			h.d.Render.Error(w, r, http.StatusConflict, "Pekerja tersebut sudah memiliki jadwal pada tanggal yang dipilih")
			return
		}

		log.Printf("create jadwal: %v", err)
		h.d.Render.Error(w, r, http.StatusInternalServerError, "Gagal menambahkan jadwal")
		return
	}

	http.Redirect(w, r, "/admin/jadwal?success=created", http.StatusSeeOther)
}

func (h *JadwalHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		h.d.Render.Error(w, r, http.StatusBadRequest, "ID jadwal tidak valid")
		return
	}

	if err := r.ParseForm(); err != nil {
		h.d.Render.Error(w, r, http.StatusBadRequest, "Form tidak valid")
		return
	}

	pekerjaID, errPekerja := strconv.ParseInt(r.FormValue("pekerja_id"), 10, 64)
	shiftID, errShift := strconv.ParseInt(r.FormValue("shift_id"), 10, 64)
	tanggal, errTanggal := time.Parse("2006-01-02", r.FormValue("tanggal_jadwal"))

	if errPekerja != nil || errShift != nil || errTanggal != nil {
		h.d.Render.Error(w, r, http.StatusBadRequest, "Data jadwal tidak valid")
		return
	}

	input := repository.CreateJadwalParams{
		PekerjaID:     pekerjaID,
		ShiftID:       shiftID,
		TanggalJadwal: tanggal,
	}

	if err := h.svc.Update(r.Context(), id, input); err != nil {
		if errors.Is(err, service.ErrInputTidakValid) {
			h.d.Render.Error(w, r, http.StatusBadRequest, "Data jadwal tidak valid. Periksa kembali input yang diberikan")
			return
		}

		if errors.Is(err, repository.ErrJadwalDuplikat) {
			h.d.Render.Error(w, r, http.StatusConflict, "Pekerja tersebut sudah memiliki jadwal pada tanggal yang dipilih")
			return
		}

		log.Printf("update jadwal: %v", err)
		h.d.Render.Error(w, r, http.StatusInternalServerError, "Gagal memperbarui jadwal")
		return
	}

	http.Redirect(w, r, "/admin/jadwal?success=updated", http.StatusSeeOther)
}

func (h *JadwalHandler) Kalender(w http.ResponseWriter, r *http.Request) {
	tanggalAcuan := time.Now()

	if tanggalQuery := r.URL.Query().Get("tanggal"); tanggalQuery != "" {
		parsed, err := time.Parse("2006-01-02", tanggalQuery)
		if err != nil {
			h.d.Render.Error(w, r, http.StatusBadRequest, "Format tanggal tidak valid")
			return
		}

		tanggalAcuan = parsed
	}

	selisih := (int(tanggalAcuan.Weekday()) + 6) % 7
	awalMinggu := tanggalAcuan.AddDate(0, 0, -selisih)

	akhirMinggu := awalMinggu.AddDate(0, 0, 6)

	jadwal, err := h.svc.GetKalender(r.Context(), awalMinggu, akhirMinggu)
	if err != nil {
		log.Printf("kalender jadwal: %v", err)
		h.d.Render.Error(w, r, http.StatusInternalServerError, "Gagal mengambil kalender jadwal")
		return
	}

	pekerja, err := h.svc.GetPekerjaAktif(r.Context())
	if err != nil {
		log.Printf("daftar pekerja jadwal: %v", err)
		h.d.Render.Error(w, r, http.StatusInternalServerError, "Gagal mengambil daftar pekerja")
		return
	}

	shifts, err := h.svc.GetShiftAktif(r.Context())
	if err != nil {
		log.Printf("daftar shift jadwal: %v", err)
		h.d.Render.Error(w, r, http.StatusInternalServerError, "Gagal mengambil daftar shift")
		return
	}

	flash := ""
	switch r.URL.Query().Get("success") {
	case "created":
		flash = "Jadwal berhasil ditambahkan"
	case "updated":
		flash = "Jadwal berhasil diperbarui"
	}

	h.d.Render.Page(w, r, "jadwal/kalender.html", PageData{
		Title: "Jadwal Kerja",
		Menu:  "jadwal",
		Flash: flash,
		Data: map[string]interface{}{
			"Jadwal":      jadwal,
			"Pekerja":     pekerja,
			"Shifts":      shifts,
			"AwalMinggu":  awalMinggu,
			"AkhirMinggu": akhirMinggu,
			"Tanggal":     tanggalAcuan.Format("2006-01-02"),
		},
	})
}

func (h *JadwalHandler) Saya(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFrom(r.Context())

	if user == nil || user.PekerjaID <= 0 {
		h.d.Render.Error(w, r, http.StatusForbidden, "Data pekerja tidak ditemukan")
		return
	}

	jadwal, err := h.svc.GetJadwalPekerja(r.Context(), user.PekerjaID)
	if err != nil {
		log.Printf("jadwal pekerja: %v", err)
		h.d.Render.Error(w, r, http.StatusInternalServerError, "Gagal mengambil jadwal kerja")
		return
	}

	h.d.Render.Page(w, r, "jadwal/saya.html", PageData{
		Title: "Jadwal Saya",
		Menu:  "jadwal",
		Data: map[string]interface{}{
			"Jadwal": jadwal,
		},
	})
}
