package handler

import (
	"database/sql"
	"net/http"

	"penggajian/internal/config"
	"penggajian/internal/middleware"
	"penggajian/internal/model"
)

// Deps adalah dependensi bersama yang diterima setiap handler.
type Deps struct {
	DB       *sql.DB
	Render   *Renderer
	Sessions *middleware.Manager
	Cfg      config.Config
}

// Guards membungkus handler dengan pembatasan akses (RBAC).
type Guards struct {
	Login   func(http.Handler) http.Handler // harus login
	Admin   func(http.Handler) http.Handler // role admin
	Pekerja func(http.Handler) http.Handler // role pekerja
}

// Module diimplementasikan setiap modul handler.
type Module interface {
	Register(mux *http.ServeMux, g Guards)
}

// Routes merakit seluruh aplikasi. SEMUA modul sudah terdaftar di bawah ini,
// sehingga anggota tim tidak perlu mengubah file ini (menghindari merge conflict).
//
// Konvensi prefix URL: admin => /admin/<modul>, pekerja => /pekerja/<modul>.
func Routes(d Deps) http.Handler {
	mux := http.NewServeMux()
	g := Guards{
		Login:   middleware.RequireLogin,
		Admin:   middleware.RequireRole(model.RoleAdmin),
		Pekerja: middleware.RequireRole(model.RolePekerja),
	}

	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	modules := []Module{
		// ===== Anggota 1 =====
		NewAuthHandler(d),
		NewHomeHandler(d),
		NewPekerjaHandler(d),
		NewUpahHandler(d),
		NewShiftHandler(d),
		NewJadwalHandler(d),
		// ===== Anggota 2 =====
		NewAbsensiHandler(d),
		NewValidasiHandler(d),
		NewKasbonHandler(d),
		NewLaporanPresensiKasbonHandler(d),
		// ===== Anggota 3 =====
		NewPeriodeHandler(d),
		NewPenggajianHandler(d),
		NewKoreksiHandler(d),
		NewSlipHandler(d),
		NewLaporanGajiHandler(d),
	}
	for _, m := range modules {
		m.Register(mux, g)
	}

	var h http.Handler = mux
	h = d.Sessions.Load(d.Cfg.DevAutoLoginRole)(h)
	h = middleware.Recover(middleware.Logging(h))
	return h
}
