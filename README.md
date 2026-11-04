# go-penggajian

Sistem Penggajian Pekerja Harian dan Shift — Konveksi Mitra Jaya
Project Akhir Web Programming (Topik 3) · Go + MySQL · `net/http` + `html/template`

## Struktur
```
go-penggajian/
├── cmd/web/main.go           # entry point (frozen)
├── db/
│   ├── schema.sql            # skema MySQL (frozen)
│   └── seed.sql              # data contoh + skenario uji
├── internal/
│   ├── config/               # env & koneksi DB (frozen)
│   ├── handler/              # HTTP handler per modul + routes.go
│   ├── middleware/           # session, RBAC, logging (frozen)
│   ├── model/                # struct tabel (frozen)
│   ├── repository/           # query SQL per modul
│   └── service/              # logika bisnis per modul
├── templates/                # layout, header, footer, sidebar, home, profil, login + folder per modul
├── static/css/style.css
├── docs/PEMBAGIAN_TUGAS.md   # siapa mengerjakan apa
├── CONTRIBUTING.md           # aturan Git
├── .env.example  .gitignore  go.mod
```

## Menjalankan
```bash
cp .env.example .env                 # Windows: copy .env.example .env
mysql -u root -p < db/schema.sql
mysql -u root -p < db/seed.sql
go mod tidy                          # sekali saja; commit go.sum
go run ./cmd/web                     # http://localhost:8080
go test ./...
```
Saat development, `DEV_AUTO_LOGIN_ROLE=admin` (atau `pekerja`) di `.env` melewati login. Kosongkan untuk uji login sungguhan. Akun seed: `admin`, `budi`, `siti`, `andi`, `dewi` — password `password123`.

## Setup awal oleh ketua tim (sekali)
1. Buat repo GitHub, push isi folder ini ke `main`.
2. Jalankan `go mod tidy`, commit `go.sum`.
3. Ganti `@anggota1/2/3` dan `@ketua-tim` di `.github/CODEOWNERS` dengan username GitHub asli.
4. Settings → Branches → protect `main` (wajib PR, 1 approval, status check CI, require review dari code owners).
5. Undang 3 anggota, buat Issue per checklist di `docs/PEMBAGIAN_TUGAS.md` (label `anggota-1/2/3`).
