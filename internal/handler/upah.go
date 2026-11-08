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
		http.Error(w, "ID pekerja tidak valid", http.StatusBadRequest)
		return
	}

	records, err := h.svc.List(r.Context(), pekerjaID)
	if err != nil {
		log.Printf("gagal mengambil riwayat upah; %v", err)
		http.Error(w, "Terjadi kesalahan pada server", http.StatusInternalServerError)
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
		http.Error(w, "ID pekerja tidak valid", http.StatusBadRequest)
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
		http.Error(w, "ID pekerja tidak valid", http.StatusBadRequest)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Form tidak valid", http.StatusBadRequest)
		return
	}

	nominal, err := strconv.ParseFloat(r.FormValue("nominal_gaji"), 64)
	if err != nil {
		http.Error(w, "Nominal upah tidak valid", http.StatusBadRequest)
		return
	}

	tanggal, err := time.Parse("2006-01-02", r.FormValue("tanggal_berlaku"))
	if err != nil {
		http.Error(w, "Tanggal berlaku tidak valid", http.StatusBadRequest)
		return
	}

	err = h.svc.Create(r.Context(), service.UpahInput{
		PekerjaID:      pekerjaID,
		NominalGaji:    nominal,
		TanggalBerlaku: tanggal,
	})

	if err != nil {
		if errors.Is(err, service.ErrInputTidakValid) {
			http.Error(w, "Data upah tidak valid", http.StatusBadRequest)
			return
		}

		if errors.Is(err, service.ErrTanggalUpahDuplikat) {
			http.Error(w, "Tanggal berlaku tersebut sudah memiliki riwayat upah. Gunakan tanggal lain.", http.StatusConflict)
			return
		}

		log.Printf("gagal menyimpan riawyat upah: %v", err)
		http.Error(w, "Gagal menyimpan riwayat upah", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin/pekerja/"+strconv.FormatInt(pekerjaID, 10)+"/upah", http.StatusSeeOther)

}
