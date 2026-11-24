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

// Tabel: riwayat_gaji (tulis)
type UpahHandler struct {
	d   Deps
	svc *service.UpahService
}

func NewUpahHandler(d Deps) *UpahHandler {
	return &UpahHandler{d: d, svc: service.NewUpahService(repository.NewUpahRepository(d.DB))}
}

func (h *UpahHandler) Register(mux *http.ServeMux, g Guards) {
	mux.Handle("GET /admin/pekerja/{id}/upah", g.Admin(http.HandlerFunc(h.List)))
	mux.Handle("GET /admin/pekerja/{id}/upah/baru", g.Admin(http.HandlerFunc(h.NewForm)))
	mux.Handle("POST /admin/pekerja/{id}/upah", g.Admin(http.HandlerFunc(h.Create)))
}

func (h *UpahHandler) List(w http.ResponseWriter, r *http.Request) {
	pekerjaID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || pekerjaID <= 0 {
		h.d.Render.Error(w, r, http.StatusBadRequest, "ID pekerja tidak valid")
		return
	}

	records, err := h.svc.List(r.Context(), pekerjaID)
	if err != nil {
		log.Printf("gagal mengambil riwayat upah; %v", err)
		h.d.Render.Error(w, r, http.StatusInternalServerError, "Terjadi kesalahan saat mengambil riwayat upah")
		return
	}

	h.d.Render.Page(w, r, "upah/list.html", PageData{
		Title: "Riwayat Upah",
		Menu:  "pekerja",
		Data: map[string]interface{}{
			"PekerjaID": pekerjaID,
			"Records":   records,
		},
	})
}

func (h *UpahHandler) NewForm(w http.ResponseWriter, r *http.Request) {
	pekerjaID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || pekerjaID <= 0 {
		h.d.Render.Error(w, r, http.StatusBadRequest, "ID pekerja tidak valid")
		return
	}

	h.d.Render.Page(w, r, "upah/form.html", PageData{
		Title: "Tambah Riwayat Upah",
		Menu:  "pekerja",
		Data: map[string]int64{
			"PekerjaID": pekerjaID,
		},
	})
}

func (h *UpahHandler) Create(w http.ResponseWriter, r *http.Request) {
	pekerjaID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || pekerjaID <= 0 {
		h.d.Render.Error(w, r, http.StatusBadRequest, "ID pekerja tidak valid")
		return
	}

	if err := r.ParseForm(); err != nil {
		h.d.Render.Error(w, r, http.StatusBadRequest, "Form tidak valid")
		return
	}

	nominal, err := strconv.ParseFloat(r.FormValue("nominal_gaji"), 64)
	if err != nil {
		h.d.Render.Error(w, r, http.StatusBadRequest, "Nominal upah tidak valid")
		return
	}

	tanggal, err := time.Parse("2006-01-02", r.FormValue("tanggal_berlaku"))
	if err != nil {
		h.d.Render.Error(w, r, http.StatusBadRequest, "Tanggal berlaku tidak valid")
		return
	}

	err = h.svc.Create(r.Context(), service.UpahInput{
		PekerjaID:      pekerjaID,
		NominalGaji:    nominal,
		TanggalBerlaku: tanggal,
	})

	if err != nil {
		if errors.Is(err, service.ErrInputTidakValid) {
			h.d.Render.Error(w, r, http.StatusBadRequest, "Data upah tidak valid")
			return
		}

		if errors.Is(err, service.ErrTanggalUpahDuplikat) {
			h.d.Render.Error(w, r, http.StatusConflict, "Tanggal berlaku tersebut sudah memiliki riwayat upah. Gunakan tanggal lain.")
			return
		}

		log.Printf("gagal menyimpan riawyat upah: %v", err)
		h.d.Render.Error(w, r, http.StatusInternalServerError, "Gagal menyimpan riwayat upah")
		return
	}

	http.Redirect(w, r, "/admin/pekerja/"+strconv.FormatInt(pekerjaID, 10)+"/upah", http.StatusSeeOther)

}
