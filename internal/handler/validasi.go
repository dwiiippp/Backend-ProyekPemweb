package handler

import (
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"penggajian/internal/middleware"
	"penggajian/internal/repository"
	"penggajian/internal/service"
)

type ValidasiHandler struct {
	d   Deps
	svc *service.ValidasiService
}

func NewValidasiHandler(d Deps) *ValidasiHandler {
	return &ValidasiHandler{d: d, svc: service.NewValidasiService(repository.NewValidasiRepository(d.DB))}
}

func (h *ValidasiHandler) Register(mux *http.ServeMux, g Guards) {
	mux.Handle("GET /admin/validasi", g.Admin(http.HandlerFunc(h.List)))
	mux.Handle("GET /admin/validasi/{tanggal}", g.Admin(http.HandlerFunc(h.Detail)))
	mux.Handle("POST /admin/validasi/{tanggal}", g.Admin(http.HandlerFunc(h.Simpan)))
}

func (h *ValidasiHandler) List(w http.ResponseWriter, r *http.Request) {
	days, err := h.svc.DaftarHari(r.Context())
	if err != nil {
		log.Printf("daftar validasi: %v", err)
		h.d.Render.Error(w, r, http.StatusInternalServerError, "Gagal memuat daftar tanggal presensi")
		return
	}
	flash := ""
	if r.URL.Query().Get("sukses") == "1" {
		flash = "Validasi presensi berhasil disimpan dan tanggal ditutup."
	}
	h.d.Render.Page(w, r, "validasi/list.html", PageData{
		Title: "Validasi Presensi Harian", Menu: "validasi",
		Flash: flash,
		Data: struct {
			Hari []repository.HariValidasi
		}{Hari: days},
	})
}

func (h *ValidasiHandler) Detail(w http.ResponseWriter, r *http.Request) {
	tanggal, items, closed, err := h.svc.Detail(r.Context(), r.PathValue("tanggal"))
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInputTidakValid):
			h.d.Render.Error(w, r, http.StatusBadRequest, "Format tanggal tidak valid")
		case errors.Is(err, service.ErrTidakDitemukan):
			h.d.Render.Error(w, r, http.StatusNotFound, "Tidak ada presensi pada tanggal tersebut")
		default:
			log.Printf("detail validasi: %v", err)
			h.d.Render.Error(w, r, http.StatusInternalServerError, "Gagal memuat presensi")
		}
		return
	}
	h.renderDetail(w, r, tanggal, items, closed, "")
}

func (h *ValidasiHandler) Simpan(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		h.d.Render.Error(w, r, http.StatusBadRequest, "Form validasi tidak valid")
		return
	}
	tanggalText := strings.TrimSpace(r.PathValue("tanggal"))
	tanggal, items, closed, err := h.svc.Detail(r.Context(), tanggalText)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInputTidakValid):
			h.d.Render.Error(w, r, http.StatusBadRequest, "Tanggal validasi tidak valid")
		case errors.Is(err, service.ErrTidakDitemukan):
			h.d.Render.Error(w, r, http.StatusNotFound, "Tidak ada presensi pada tanggal tersebut")
		default:
			log.Printf("detail validasi untuk simpan: %v", err)
			h.d.Render.Error(w, r, http.StatusInternalServerError, "Gagal memuat presensi")
		}
		return
	}

	decisions := make(map[int64]service.InputKeputusanValidasi, len(items))
	for _, item := range items {
		if _, exists := r.Form["status_"+intString(item.ID)]; !exists {
			h.renderDetail(w, r, tanggal, items, closed, "Tentukan keputusan untuk setiap presensi.")
			return
		}
		decisions[item.ID] = service.InputKeputusanValidasi{
			Status:  r.FormValue("status_" + intString(item.ID)),
			Catatan: r.FormValue("catatan_" + intString(item.ID)),
		}
	}

	user := middleware.UserFrom(r.Context())
	err = h.svc.Simpan(r.Context(), tanggalText, user.UserID, decisions)
	switch {
	case err == nil:
		http.Redirect(w, r, "/admin/validasi?sukses=1", http.StatusSeeOther)
	case errors.Is(err, service.ErrInputTidakValid):
		h.renderDetail(w, r, tanggal, items, closed, "Input tidak valid. Pilih setujui/tolak untuk setiap presensi dan isi catatan jika menolak.")
	case errors.Is(err, service.ErrHariDitutup):
		h.renderDetail(w, r, tanggal, items, true, "Tanggal ini sudah divalidasi dan ditutup.")
	case errors.Is(err, service.ErrPeriodeDitutup):
		h.renderDetail(w, r, tanggal, items, closed, "Periode gaji sudah ditutup; presensi tidak dapat diubah.")
	default:
		log.Printf("simpan validasi: %v", err)
		h.d.Render.Error(w, r, http.StatusInternalServerError, "Gagal menyimpan validasi presensi")
	}
}

func (h *ValidasiHandler) renderDetail(w http.ResponseWriter, r *http.Request, tanggal time.Time, items []repository.PresensiValidasi, closed bool, pageError string) {
	h.d.Render.Page(w, r, "validasi/detail.html", PageData{
		Title: "Validasi Presensi", Menu: "validasi", Error: pageError,
		Data: struct {
			Tanggal     time.Time
			Hari        []repository.PresensiValidasi
			Ditutup     bool
			AdaPresensi bool
		}{Tanggal: tanggal, Hari: items, Ditutup: closed, AdaPresensi: len(items) > 0},
	})
}

func intString(value int64) string {
	return strconv.FormatInt(value, 10)
}
