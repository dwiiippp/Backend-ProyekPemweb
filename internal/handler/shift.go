package handler

import (
	"errors"
	"log"
	"net/http"
	"strconv"

	"penggajian/internal/model"
	"penggajian/internal/repository"
	"penggajian/internal/service"
)

// Pemilik: Anggota 1 | F-04 Kelola Shift (BR-05)
// Tabel: shift (tulis)
type ShiftHandler struct {
	d   Deps
	svc *service.ShiftService
}

func NewShiftHandler(d Deps) *ShiftHandler {
	return &ShiftHandler{d: d, svc: service.NewShiftService(repository.NewShiftRepository(d.DB))}
}

func (h *ShiftHandler) Register(mux *http.ServeMux, g Guards) {
	mux.Handle("GET /admin/shift", g.Admin(http.HandlerFunc(h.List)))
	mux.Handle("GET /admin/shift/baru", g.Admin(http.HandlerFunc(h.NewForm)))
	mux.Handle("POST /admin/shift", g.Admin(http.HandlerFunc(h.Create)))
	mux.Handle("GET /admin/shift/{id}/edit", g.Admin(http.HandlerFunc(h.EditForm)))
	mux.Handle("POST /admin/shift/{id}/edit", g.Admin(http.HandlerFunc(h.Update)))
	mux.Handle("POST /admin/shift/{id}/hapus", g.Admin(http.HandlerFunc(h.Delete)))
}

func (h *ShiftHandler) List(w http.ResponseWriter, r *http.Request) {
	shifts, err := h.svc.List(r.Context())
	if err != nil {
		log.Printf("list shift: %v", err)
		h.d.Render.Error(w, r, http.StatusInternalServerError, "Gagal mengambil data shift")
		return
	}

	flash := ""
	switch r.URL.Query().Get("success") {
	case "created":
		flash = "Shift berhasil ditambahkan"
	case "updated":
		flash = "Shift berhasil diperbarui"
	case "deleted":
		flash = "Shift berhasil dihapus"
	}

	h.d.Render.Page(w, r, "shift/list.html", PageData{
		Title: "Master Shift",
		Menu:  "shift",
		Flash: flash,
		Data: map[string]interface{}{
			"Shifts": shifts,
		},
	})
}

func (h *ShiftHandler) NewForm(w http.ResponseWriter, r *http.Request) {
	h.renderForm(w, r, "Tambah Shift", model.Shift{}, "/admin/shift", "")
}

func (h *ShiftHandler) EditForm(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		h.d.Render.Error(w, r, http.StatusBadRequest, "ID shift tidak valid")
		return
	}

	shift, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrShiftNotFound) {
			h.d.Render.Error(w, r, http.StatusNotFound, "Shift tidak ditemukan")
			return
		}

		log.Printf("edit form shift: %v", err)
		h.d.Render.Error(w, r, http.StatusInternalServerError, "Gagal mengambil data shift")
		return
	}

	action := "/admin/shift/" + strconv.FormatInt(id, 10) + "/edit"
	h.renderForm(w, r, "Edit Shift", shift, action, "")

}

func (h *ShiftHandler) Create(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		h.d.Render.Error(w, r, http.StatusBadRequest, "Form tidak valid")
		return
	}

	shift := model.Shift{
		NamaShift: r.FormValue("nama_shift"),
		JamMasuk:  r.FormValue("jam_masuk"),
		JamKeluar: r.FormValue("jam_keluar"),
	}

	err := h.svc.Create(r.Context(), service.ShiftInput{
		NamaShift: shift.NamaShift,
		JamMasuk:  shift.JamMasuk,
		JamKeluar: shift.JamKeluar,
	})

	if err != nil {
		if errors.Is(err, service.ErrShiftTidakValid) {
			h.renderForm(w, r, "Tambah Shift", shift, "/admin/shift", "Data shift tidak valid. Periksa nama dan jam shift")
			return
		}

		log.Printf("create shift: %v", err)
		h.d.Render.Error(w, r, http.StatusInternalServerError, "Gagal menambahkan shift")
		return
	}

	http.Redirect(w, r, "/admin/shift?success=created", http.StatusSeeOther)
}

func (h *ShiftHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		h.d.Render.Error(w, r, http.StatusBadRequest, "ID Shift tidak valid")
		return
	}

	if err := r.ParseForm(); err != nil {
		h.d.Render.Error(w, r, http.StatusBadRequest, "Form tidak valid")
		return
	}

	shift := model.Shift{
		ID:        id,
		NamaShift: r.FormValue("nama_shift"),
		JamMasuk:  r.FormValue("jam_masuk"),
		JamKeluar: r.FormValue("jam_keluar"),
	}

	err = h.svc.Update(r.Context(), id, service.ShiftInput{
		NamaShift: shift.NamaShift,
		JamMasuk:  shift.JamMasuk,
		JamKeluar: shift.JamKeluar,
	})

	if err != nil {
		if errors.Is(err, service.ErrShiftTidakValid) {
			action := "/admin/shift/" + strconv.FormatInt(id, 10) + "/edit"
			h.renderForm(w, r, "Edit Shift", shift, action, "Data shift tidak valid. Periksa nama dan jam shift.")
			return
		}

		if errors.Is(err, repository.ErrShiftNotFound) {
			h.d.Render.Error(w, r, http.StatusNotFound, "Shift tidak ditemukan")
			return
		}

		log.Printf("update shift: %v", err)
		h.d.Render.Error(w, r, http.StatusInternalServerError, "Gagal memperbarui shift")
		return

	}
	http.Redirect(w, r, "/admin/shift?success=updated", http.StatusSeeOther)
}

func (h *ShiftHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		h.d.Render.Error(w, r, http.StatusBadRequest, "ID shift tidak valid")
		return
	}

	err = h.svc.Delete(r.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrShiftNotFound) {
			h.d.Render.Error(w, r, http.StatusNotFound, "Shift tidak ditemukan")
			return
		}

		log.Printf("delete shift: %v", err)
		h.d.Render.Error(w, r, http.StatusInternalServerError, "Gagal menghapus shift")
		return
	}
	http.Redirect(w, r, "/admin/shift?success=deleted", http.StatusSeeOther)
}

func (h *ShiftHandler) renderForm(w http.ResponseWriter, r *http.Request, title string, shift model.Shift, action string, errMsg string) {
	h.d.Render.Page(w, r, "shift/form.html", PageData{
		Title: title,
		Menu:  "shift",
		Error: errMsg,
		Data: map[string]interface{}{
			"Shift":  shift,
			"Action": action,
		},
	})
}
