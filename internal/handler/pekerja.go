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

// Pemilik: Anggota 1 | F-02 · UC-08 Kelola Pekerja
// Tabel: pekerja, users (tulis)
type PekerjaHandler struct {
	d   Deps
	svc *service.PekerjaService
}

func NewPekerjaHandler(d Deps) *PekerjaHandler {
	return &PekerjaHandler{d: d, svc: service.NewPekerjaService(repository.NewPekerjaRepository(d.DB))}
}

func (h *PekerjaHandler) Register(mux *http.ServeMux, g Guards) {
	mux.Handle("GET /admin/pekerja", g.Admin(http.HandlerFunc(h.List)))
	mux.Handle("GET /admin/pekerja/baru", g.Admin(http.HandlerFunc(h.NewForm)))
	mux.Handle("POST /admin/pekerja", g.Admin(http.HandlerFunc(h.Create)))
	mux.Handle("GET /admin/pekerja/{id}", g.Admin(http.HandlerFunc(h.Detail)))
	mux.Handle("GET /admin/pekerja/{id}/edit", g.Admin(http.HandlerFunc(h.EditForm)))
	mux.Handle("POST /admin/pekerja/{id}/edit", g.Admin(http.HandlerFunc(h.Update)))
	mux.Handle("POST /admin/pekerja/{id}/nonaktifkan", g.Admin(http.HandlerFunc(h.Deactivate)))
}

func (h *PekerjaHandler) List(w http.ResponseWriter, r *http.Request) {
	pekerja, err := h.svc.List(r.Context())
	if err != nil {
		log.Printf("list pekerja: %v", err)
		h.d.Render.Error(w, r, http.StatusInternalServerError, "Gagal mengambil data pekerja")
		return
	}

	flash := ""
	if r.URL.Query().Get("success") == "1" {
		flash = "Pekerja berhasil ditambahkan"
	}

	h.d.Render.Page(w, r, "pekerja/list.html", PageData{
		Title: "Data Pekerja",
		Menu:  "pekerja",
		Flash: flash,
		Data:  pekerja,
	})
}

func (h *PekerjaHandler) NewForm(w http.ResponseWriter, r *http.Request) {
	h.d.Render.Page(w, r, "pekerja/form.html", PageData{
		Title: "Tambah Pekerja",
		Menu:  "pekerja",
	})
}

// Create memproses penambahan pekerja sekaligus akun login.
func (h *PekerjaHandler) Create(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		h.d.Render.Error(w, r, http.StatusBadRequest, "Form tidak valid")
		return
	}

	tanggalMasuk, err := time.Parse("2006-01-02", r.FormValue("tanggal_masuk"))
	if err != nil {
		h.d.Render.Page(w, r, "pekerja/form.html", PageData{
			Title: "Tambah Pekerja",
			Menu:  "pekerja",
			Error: "Format tanggal masuk tidak valid",
		})
		return
	}

	input := service.PekerjaInput{
		NIK:           r.FormValue("nik"),
		Nama:          r.FormValue("nama"),
		Email:         r.FormValue("email"),
		Password:      r.FormValue("password"),
		KategoriKerja: r.FormValue("kategori_kerja"),
		TanggalMasuk:  tanggalMasuk,
	}

	_, err = h.svc.Create(r.Context(), input)
	if err != nil {
		msg := "Gagal menambahkan pekerja"

		if errors.Is(err, service.ErrInputTidakValid) {
			msg = "Periksa kembali data. Semua kolom wajib diisi, email harus valid, dan password minimal 8 karakter."
		} else {
			log.Printf("create pekerja: %v", err)
		}

		h.d.Render.Page(w, r, "pekerja/form.html", PageData{
			Title: "Tambah Pekerja",
			Menu:  "pekerja",
			Error: msg,
		})
		return
	}

	http.Redirect(w, r, "/admin/pekerja?success=1", http.StatusSeeOther)
}

func (h *PekerjaHandler) Detail(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		h.d.Render.Error(w, r, http.StatusBadRequest, "ID pekerja tidak valid")
		return
	}

	pekerja, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrTidakDitemukan) {
			h.d.Render.Error(w, r, http.StatusNotFound, "Data pekerja tidak ditemukan")
			return
		}

		log.Printf("detail pekerja: %v", err)
		h.d.Render.Error(w, r, http.StatusInternalServerError, "Gagal mengambil detail pekerja")
		return
	}

	h.d.Render.Page(w, r, "pekerja/detail.html", PageData{
		Title: "Detail Pekerja",
		Menu:  "pekerja",
		Data:  pekerja,
	})
}

func (h *PekerjaHandler) EditForm(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		h.d.Render.Error(w, r, http.StatusBadRequest, "ID pekerja tidak valid")
		return
	}

	pekerja, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrTidakDitemukan) {
			h.d.Render.Error(w, r, http.StatusNotFound, "Data pekerja tidak ditemukan")
			return
		}
		log.Printf("edit  form pekerja: %v", err)
		h.d.Render.Error(w, r, http.StatusInternalServerError, "Gagal mengambil data pekerja")
		return
	}

	h.d.Render.Page(w, r, "pekerja/form.html", PageData{
		Title: "Edit Pekerja",
		Menu:  "pekerja",
		Data:  pekerja,
	})
}

func (h *PekerjaHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		h.d.Render.Error(w, r, http.StatusBadRequest, "ID pekerja tidak valid")
		return
	}

	if err := r.ParseForm(); err != nil {
		h.d.Render.Error(w, r, http.StatusBadRequest, "Form tidak valid")
		return
	}

	tanggalMasuk, err := time.Parse("2006-01-02", r.FormValue("tanggal_masuk"))
	if err != nil {
		h.d.Render.Error(w, r, http.StatusBadRequest, "Format tanggal masuk tidak valid")
		return
	}

	input := service.PekerjaInput{
		NIK:           r.FormValue("nik"),
		Nama:          r.FormValue("nama"),
		Email:         r.FormValue("email"),
		KategoriKerja: r.FormValue("kategori_kerja"),
		TanggalMasuk:  tanggalMasuk,
	}

	if err := h.svc.Update(r.Context(), id, input); err != nil {
		msg := "Gagal memperbarui data pekerja"

		if errors.Is(err, service.ErrInputTidakValid) {
			msg = "Data pekerja tidak valid. Periksa kembali seluruh kolom"
		} else if errors.Is(err, service.ErrTidakDitemukan) {
			h.d.Render.Error(w, r, http.StatusNotFound, "Data pekerja tidak ditemukan")
			return
		} else {
			log.Printf("Update pekerja: %v", err)
		}

		h.d.Render.Page(w, r, "pekerja/form.html", PageData{
			Title: "Edit Pekerja",
			Menu:  "pekerja",
			Error: msg,
		})
		return
	}
	http.Redirect(w, r, "/admin/pekerja/"+strconv.FormatInt(id, 10), http.StatusSeeOther)
}

func (h *PekerjaHandler) Deactivate(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		http.Error(w, "ID pekerja tidak valid", http.StatusBadRequest)
		return
	}

	err = h.svc.Deactivate(r.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrTidakDitemukan) {
			http.Error(w, "Pekerja tidak ditemukan", http.StatusNotFound)
			return
		}

		log.Printf("gagal menonaktifkan pekerja: %v", err)
		http.Error(w, "Terja kesalahan pada server", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin/pekerja", http.StatusSeeOther)
}
