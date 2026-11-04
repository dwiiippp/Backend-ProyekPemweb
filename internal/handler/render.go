// Package handler berisi HTTP handler per modul + routing.
// render.go, routes.go: FILE BERSAMA (frozen).
package handler

import (
	"bytes"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/shopspring/decimal"

	"penggajian/internal/middleware"
	"penggajian/internal/model"
)

// PageData adalah data standar yang diterima semua template.
type PageData struct {
	Title string // judul halaman
	Menu  string // kunci menu aktif di sidebar
	Flash string // pesan sukses
	Error string // pesan error
	User  *model.SessionUser
	Data  any // data khusus halaman (isi bebas oleh pemilik modul)
}

var sharedTemplates = map[string]bool{
	"layout.html": true, "header.html": true, "sidebar.html": true, "footer.html": true,
}

var funcMap = template.FuncMap{
	"rupiah":  rupiah,
	"tanggal": tanggal,
}

func rupiah(d decimal.Decimal) string {
	s := d.Abs().StringFixed(0)
	var b strings.Builder
	for i, c := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			b.WriteByte('.')
		}
		b.WriteRune(c)
	}
	out := "Rp" + b.String()
	if d.IsNegative() {
		out = "-" + out
	}
	return out
}

func tanggal(t time.Time) string {
	if t.IsZero() {
		return "-"
	}
	return t.Format("02-01-2006")
}

// Renderer memuat template: setiap halaman = layout + header + sidebar + footer + file halaman.
// File halaman wajib mendefinisikan {{define "content"}}...{{end}}.
type Renderer struct {
	dir    string
	reload bool // development: muat ulang template tiap request (tanpa restart)
	mu     sync.RWMutex
	pages  map[string]*template.Template
}

func NewRenderer(dir string, reload bool) (*Renderer, error) {
	r := &Renderer{dir: dir, reload: reload}
	if err := r.load(); err != nil {
		return nil, err
	}
	return r, nil
}

func (r *Renderer) load() error {
	var shared []string
	for name := range sharedTemplates {
		shared = append(shared, filepath.Join(r.dir, name))
	}

	pages := make(map[string]*template.Template)
	err := filepath.WalkDir(r.dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".html") {
			return nil
		}
		rel, err := filepath.Rel(r.dir, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		if sharedTemplates[rel] {
			return nil
		}
		files := append(append([]string{}, shared...), path)
		t, err := template.New(rel).Funcs(funcMap).ParseFiles(files...)
		if err != nil {
			return err
		}
		pages[rel] = t
		return nil
	})
	if err != nil {
		return err
	}
	r.mu.Lock()
	r.pages = pages
	r.mu.Unlock()
	return nil
}

// Page merender template `name` (path relatif dari folder templates, mis. "pekerja/list.html").
func (r *Renderer) Page(w http.ResponseWriter, req *http.Request, name string, pd PageData) {
	r.render(w, req, http.StatusOK, name, pd)
}

// Error merender halaman error dengan status HTTP tertentu.
func (r *Renderer) Error(w http.ResponseWriter, req *http.Request, status int, msg string) {
	r.render(w, req, status, "error.html", PageData{Title: http.StatusText(status), Error: msg})
}

func (r *Renderer) render(w http.ResponseWriter, req *http.Request, status int, name string, pd PageData) {
	if r.reload {
		if err := r.load(); err != nil {
			log.Printf("template: %v", err)
			http.Error(w, "gagal memuat template: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}
	r.mu.RLock()
	t, ok := r.pages[name]
	r.mu.RUnlock()
	if !ok {
		http.Error(w, "template tidak ditemukan: "+name, http.StatusInternalServerError)
		return
	}
	if pd.User == nil {
		pd.User = middleware.UserFrom(req.Context())
	}

	var buf bytes.Buffer
	if err := t.ExecuteTemplate(&buf, "layout", pd); err != nil {
		log.Printf("template %s: %v", name, err)
		http.Error(w, "gagal merender halaman", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	_, _ = buf.WriteTo(w)
}
