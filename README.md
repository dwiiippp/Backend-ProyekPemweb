# go-penggajian

Sistem Penggajian Pekerja Harian dan Shift — **Konveksi Mitra Jaya**
Project Akhir Web Programming (Topik 3) · Go + MySQL · `net/http` + `html/template`

> **Panduan setup untuk anggota tim backend.** Ikuti bagian 1–3 berurutan sampai *Checklist siap mulai* tercentang semua, baru ambil Issue pertama Anda. Bagian lainnya adalah acuan yang dibuka saat dibutuhkan.

**Daftar isi**
1. [Gambaran singkat](#1-gambaran-singkat)
2. [Prasyarat](#2-prasyarat)
3. [Setup untuk anggota tim](#3-setup-untuk-anggota-tim)
4. [Konfigurasi `.env`](#4-konfigurasi-env)
5. [Struktur proyek](#5-struktur-proyek)
6. [Pembagian tugas](#6-pembagian-tugas)
7. [Alur kerja Git harian](#7-alur-kerja-git-harian)
8. [Panduan menulis kode](#8-panduan-menulis-kode)
9. [Contoh: mengerjakan satu fitur dari awal](#9-contoh-mengerjakan-satu-fitur-dari-awal)
10. [Mengubah database / file bersama](#10-mengubah-database--file-bersama)
11. [Data seed & akun uji](#11-data-seed--akun-uji)
12. [Testing & CI](#12-testing--ci)
13. [Troubleshooting](#13-troubleshooting)

---

## 1. Gambaran singkat

- **Bahasa/stack:** Go 1.22+, MySQL 8, `net/http` (ServeMux dengan pola `METHOD /path`), `html/template`.
- **Arsitektur berlapis:** `handler` (HTTP) → `service` (logika bisnis) → `repository` (SQL) → MySQL. Struct tabel ada di `model`.
- **Dua role:** `admin` (kelola data, validasi, penggajian) dan `pekerja` (lihat jadwal, input presensi, kasbon, slip).
- **Kebutuhan lengkap** ada di dokumen *Detail Requirement – Sistem Penggajian* (use case UC-xx, fitur F-xx, aturan bisnis BR-xx).

Proyek ini sengaja disusun agar tiap anggota bekerja di **file miliknya sendiri**. Semua modul sudah terdaftar di router, jadi tidak ada yang perlu mengedit `main.go` atau `routes.go`.

## 2. Prasyarat

| Alat | Versi | Cek | Unduh |
|---|---|---|---|
| Go | 1.22 atau lebih baru | `go version` | https://go.dev/dl/ |
| MySQL | 8.x | `mysql --version` | https://dev.mysql.com/downloads/ (atau XAMPP/Laragon yang menyertakan MySQL/MariaDB) |
| Git | terbaru | `git --version` | https://git-scm.com/downloads |
| Editor | VS Code + ekstensi **Go** | – | https://code.visualstudio.com/ |
| Akun GitHub | – | – | https://github.com/ |

Pastikan perintah `go`, `git`, dan `mysql` bisa dipanggil dari terminal. Jika `mysql` tidak dikenali di Windows, tambahkan folder `bin` MySQL ke PATH atau jalankan perintah dari folder tersebut.

Atur identitas Git (sekali per komputer):
```bash
git config --global user.name  "Nama Anda"
git config --global user.email "email-github-anda@contoh.com"
```

## 3. Setup untuk anggota tim

> Ketua tim sudah membuat repo dan mengundang Anda. Terima undangan lewat email atau di https://github.com/notifications. Belum menerima undangan? Hubungi ketua dengan menyebut username GitHub Anda.

**Langkah 1 — Clone repo**
```bash
git clone https://github.com/<username-ketua>/go-penggajian.git
cd go-penggajian
```

**Langkah 2 — Buat file `.env`**
```bash
# Windows (CMD/PowerShell)
copy .env.example .env
# macOS / Linux
cp .env.example .env
```
Buka `.env`, isi `DB_USER` / `DB_PASS` sesuai MySQL di komputer Anda (lihat [bagian 4](#4-konfigurasi-env)). File `.env` **tidak boleh di-commit** (sudah ada di `.gitignore`).

**Langkah 3 — Buat database dan isi data contoh**
```bash
mysql -u root -p < db/schema.sql
mysql -u root -p < db/seed.sql
```
(Tambahkan `-p` hanya jika MySQL memakai password; Anda akan diminta mengetik.) Perintah ini membuat database `penggajian_mitrajaya`. Setiap kali `schema.sql` dijalankan, semua tabel dihapus dan dibuat ulang — berguna untuk mereset data.

> Di PowerShell tanda `<` tidak didukung untuk redirect. Gunakan CMD, atau:
> `Get-Content db/schema.sql | mysql -u root -p` dan `Get-Content db/seed.sql | mysql -u root -p`

**Langkah 4 — Unduh dependensi Go**
```bash
go mod download
```

**Langkah 5 — Jalankan**
```bash
go run ./cmd/web
```
Buka http://localhost:8080. Dengan `DEV_AUTO_LOGIN_ROLE=admin` di `.env`, Anda langsung masuk sebagai admin dan sidebar menampilkan semua halaman (saat ini masih stub bertanda *STUB · Anggota N*).

**Langkah 6 — Pastikan semuanya sehat**
```bash
go test ./...
```
Jika semua `ok`, lingkungan Anda siap. Lanjut ke *Checklist siap mulai* di bawah.

Template HTML dimuat ulang otomatis di mode development, jadi mengubah file `.html` cukup di-refresh browser. Perubahan kode Go perlu menjalankan ulang `go run ./cmd/web`.

### Checklist siap mulai

Centang semuanya sebelum mengambil Issue pertama. Jika ada yang gagal, lihat [Troubleshooting](#13-troubleshooting) atau tanya di grup.

- [ ] `go version` menampilkan 1.22 atau lebih baru, dan `mysql --version` serta `git --version` berjalan
- [ ] Repo sudah di-clone dan file `.env` sudah dibuat dari `.env.example`
- [ ] `schema.sql` dan `seed.sql` sudah diimpor (cek: `mysql -u root -p -e "SELECT COUNT(*) FROM penggajian_mitrajaya.pekerja;"` menghasilkan 4)
- [ ] `go test ./...` semuanya `ok`
- [ ] http://localhost:8080 menampilkan dashboard dengan sidebar lengkap
- [ ] Anda sudah mengganti `DEV_AUTO_LOGIN_ROLE` ke `pekerja`, restart, dan melihat menu pekerja (lalu kembalikan ke `admin`)
- [ ] Latihan push cabang berhasil: `git checkout -b chore/setup-<nama-anda>`, lalu `git push -u origin chore/setup-<nama-anda>` (tidak perlu membuka PR; hapus cabangnya setelah itu)
- [ ] Sudah membaca [`CONTRIBUTING.md`](CONTRIBUTING.md) dan bagian anggota Anda di [`docs/PEMBAGIAN_TUGAS.md`](docs/PEMBAGIAN_TUGAS.md)
- [ ] Ada Issue yang di-assign ke Anda di tab *Issues* repo
## 4. Konfigurasi `.env`

| Variabel | Default | Keterangan |
|---|---|---|
| `APP_ENV` | `development` | `development` mengaktifkan reload template dan dev auto-login |
| `APP_PORT` | `8080` | Port server |
| `DB_HOST` | `127.0.0.1` | Host MySQL |
| `DB_PORT` | `3306` | Port MySQL |
| `DB_USER` | `root` | User MySQL |
| `DB_PASS` | *(kosong)* | Password MySQL |
| `DB_NAME` | `penggajian_mitrajaya` | Nama database |
| `DEV_AUTO_LOGIN_ROLE` | `admin` | `admin` / `pekerja` untuk melewati login saat development; kosongkan untuk uji login sungguhan |

**Tentang dev auto-login:** Jika terisi, setiap request dianggap sudah login sebagai user seed (`admin` = user id 1, `pekerja` = Budi, user id 2 / pekerja id 1). Fitur ini otomatis mati bila `APP_ENV` bukan `development`. Untuk menguji tampilan pekerja, ganti ke `pekerja` lalu restart server. Setelah modul Auth selesai, uji juga dengan mengosongkannya.

## 5. Struktur proyek

```
go-penggajian/
├── cmd/web/main.go             # entry point  [bersama/frozen]
├── db/
│   ├── schema.sql              # skema MySQL  [bersama/frozen]
│   └── seed.sql                # data contoh + skenario uji
├── internal/
│   ├── config/                 # env & koneksi DB  [frozen]
│   ├── handler/                # HTTP handler per modul; routes.go & render.go [frozen]
│   ├── middleware/             # session, RBAC, logging, recover  [frozen]
│   ├── model/                  # struct tabel  [frozen]
│   ├── repository/             # query SQL, satu file per modul
│   └── service/                # logika bisnis, satu file per modul (+ errors.go, password.go [frozen])
├── templates/
│   ├── layout.html header.html footer.html sidebar.html   # [frozen]
│   ├── home.html profil.html login.html error.html
│   └── <modul>/ list.html form.html detail.html ...       # satu folder per modul
├── static/css/style.css        # [frozen]
├── docs/PEMBAGIAN_TUGAS.md     # siapa mengerjakan apa + matriks tabel + skenario seed
├── CONTRIBUTING.md             # aturan Git & kolaborasi
├── .github/                    # CODEOWNERS, template PR, CI
├── .env.example  .gitignore  go.mod
```

Aturan emasnya: **satu modul = satu file di handler, service, repository, dan folder di templates.** Contoh modul `kasbon`: `handler/kasbon.go`, `service/kasbon.go`, `repository/kasbon.go`, `templates/kasbon/*.html`.

## 6. Pembagian tugas

| Anggota | Fokus | Modul |
|---|---|---|
| **1** | Fondasi & Data Master | `auth`, `home` (dashboard & profil), `pekerja`, `upah`, `shift`, `jadwal` |
| **2** | Presensi & Kasbon | `absensi`, `validasi`, `kasbon`, `laporan_presensi_kasbon` |
| **3** | Penggajian & Laporan | `periode`, `penggajian`, `koreksi`, `slip`, `laporan_gaji` |

Detail checklist per anggota, matriks tabel (siapa menulis/membaca), dan skenario data uji ada di [`docs/PEMBAGIAN_TUGAS.md`](docs/PEMBAGIAN_TUGAS.md). Kepemilikan file juga tercatat di `.github/CODEOWNERS`.

**Mengapa tidak saling menunggu?**
- Schema, model, dan seed sudah jadi — semua memakai data yang sama sejak hari pertama.
- Modul lain cukup *membaca* tabel milik modul lain lewat repository miliknya sendiri; tidak saling memanggil service.
- Login bisa dilewati lewat `DEV_AUTO_LOGIN_ROLE`.
- Anggota 3 bisa menulis perhitungan gaji sebagai fungsi murni + unit test memakai angka dari seed, tanpa menunggu presensi selesai.

**Langkah pertama setelah setup (semua anggota)**
1. Buka Issue pertama Anda di tab *Issues*, lalu buat cabang `feat/<modul>-<deskripsi>`.
2. Jalankan server dan buka halaman stub modul Anda lewat sidebar (ditandai *STUB · Anggota N*) untuk melihat kondisi awalnya.
3. Baca struct terkait di `internal/model/` dan data di `db/seed.sql` untuk modul Anda.
4. Kerjakan berurutan: repository → service → handler → template (lihat contoh di [bagian 9](#9-contoh-mengerjakan-satu-fitur-dari-awal)).
5. Uji manual dengan data seed, jalankan `gofmt -l .`, `go vet ./...`, `go test ./...`, lalu buka PR.
## 7. Alur kerja Git harian

```bash
git checkout main
git pull                                   # selalu mulai dari main terbaru
git checkout -b feat/kasbon-form           # buat cabang kerja

# ... kerjakan, uji, lalu commit kecil-kecil ...
git add internal/handler/kasbon.go templates/kasbon/form.html
git commit -m "feat(kasbon): form pencatatan kasbon"

git fetch origin
git rebase origin/main                     # sinkronkan sebelum push
git push -u origin feat/kasbon-form
```
Lalu buka **Pull Request** di GitHub ke `main`, isi template PR, minta review ke anggota lain. Setelah di-merge: `git checkout main && git pull` dan hapus cabang lama.

**Penamaan cabang:** `feat/<modul>-<deskripsi>`, `fix/<modul>-<deskripsi>`.
**Pesan commit:** `<tipe>(<modul>): <ringkas>` — tipe `feat`, `fix`, `refactor`, `test`, `docs`.
**Ukuran PR:** kecil (± < 400 baris) agar cepat direview.

Jika `git rebase` memunculkan konflik: buka file yang ditandai, pilih/gabungkan isi, lalu `git add <file>` dan `git rebase --continue`. Jika bingung, `git rebase --abort` dan tanyakan ke tim. Konflik seharusnya jarang karena tiap orang mengedit file berbeda; bila sering terjadi, berarti ada yang mengedit file milik orang lain.

Aturan lengkap ada di [`CONTRIBUTING.md`](CONTRIBUTING.md).

## 8. Panduan menulis kode

**Alur request**
```
Browser → routes.go → Handler → Service → Repository → MySQL
                         └── render template (templates/<modul>/*.html)
```
- **Handler** membaca input (`r.PathValue("id")`, `r.FormValue(...)`), memanggil service, lalu merender template atau redirect. Tanpa SQL, tanpa aturan bisnis.
- **Service** berisi aturan bisnis (BR-xx) dan validasi. Mengembalikan error domain dari `service/errors.go` (`ErrHariDitutup`, `ErrPeriodeDitutup`, `ErrPresensiGanda`, `ErrBukanMilikSendiri`, dst.).
- **Repository** berisi semua SQL modul tersebut.

**Aturan wajib**
1. **Uang = `decimal.Decimal`** (`github.com/shopspring/decimal`), jangan `float64`. Kolom DB `DECIMAL(12,2)`.
2. **SQL selalu pakai placeholder `?`**, tidak boleh menyambung string dengan input pengguna (cegah SQL injection).
3. **Operasi multi-tabel = satu transaksi** (`db.BeginTx` … `Commit`/`Rollback`), misalnya validasi harian dan tutup periode.
4. **Cek akses di server:** role (guard `g.Admin` / `g.Pekerja`), kepemilikan data (pekerja hanya melihat datanya, ambil `PekerjaID` dari session, **bukan** dari form), dan status hari/periode ditutup.
5. **Jangan pernah percaya input dari form**, termasuk id tersembunyi: validasi ulang di service.
6. Hash password dengan `service.HashPassword` / `service.CheckPassword` (bcrypt); jangan simpan password polos.
7. Template otomatis di-escape oleh `html/template`; jangan memakai `template.HTML` untuk input pengguna.

**Konvensi routing**
- Admin di `/admin/<modul>`, pekerja di `/pekerja/<modul>`. Daftarkan route hanya di bawah prefix modul Anda, di fungsi `Register` pada handler modul Anda.
- Pola Go 1.22: `mux.Handle("GET /admin/kasbon/{id}", g.Admin(http.HandlerFunc(h.Detail)))`; ambil parameter dengan `r.PathValue("id")`.
- Form yang mengubah data memakai `POST` (jangan `GET`). Setelah sukses, redirect (`http.Redirect(..., http.StatusSeeOther)`).

**Template**
- Setiap file halaman wajib berisi `{{define "content"}} ... {{end}}`; layout, header, sidebar, footer ditambahkan otomatis.
- Render dengan: `h.d.Render.Page(w, r, "kasbon/list.html", PageData{Title: "Kasbon", Menu: "kasbon", Data: data})`.
- `PageData` punya `Title`, `Menu` (menu aktif sidebar), `Flash` (pesan sukses), `Error` (pesan gagal), `User`, dan `Data` (isi bebas).
- Fungsi bantu di template: `{{rupiah .Nominal}}` dan `{{tanggal .Tanggal}}`.
- Akses error: `h.d.Render.Error(w, r, http.StatusNotFound, "Data tidak ditemukan")`.

## 9. Contoh: mengerjakan satu fitur dari awal

Misal Anggota 2 mengerjakan **daftar kasbon (admin)**.

1. **Issue & cabang:** ambil Issue "Kasbon – daftar admin", lalu `git checkout -b feat/kasbon-list`.
2. **Repository** (`internal/repository/kasbon.go`):
   ```go
   func (r *KasbonRepository) List(ctx context.Context) ([]model.Kasbon, error) {
       rows, err := r.db.QueryContext(ctx,
           `SELECT id, pekerja_id, tanggal_pengajuan, nominal_kasbon, sisa_kasbon, keterangan, status, dibuat_pada
            FROM kasbon ORDER BY tanggal_pengajuan DESC`)
       if err != nil { return nil, err }
       defer rows.Close()
       var out []model.Kasbon
       for rows.Next() {
           var k model.Kasbon
           if err := rows.Scan(&k.ID, &k.PekerjaID, &k.TanggalPengajuan, &k.NominalKasbon,
               &k.SisaKasbon, &k.Keterangan, &k.Status, &k.DibuatPada); err != nil { return nil, err }
           out = append(out, k)
       }
       return out, rows.Err()
   }
   ```
3. **Service** (`internal/service/kasbon.go`): tambahkan `func (s *KasbonService) Daftar(ctx context.Context) ([]model.Kasbon, error) { return s.repo.List(ctx) }`.
4. **Handler** (`internal/handler/kasbon.go`): di method `List`, panggil `h.svc.Daftar(r.Context())`; bila error, `h.d.Render.Error(w, r, 500, "Gagal memuat data")`; bila sukses, render `kasbon/list.html` dengan `Data: daftar`.
5. **Template** (`templates/kasbon/list.html`): tabel dengan `{{range .Data}}...{{rupiah .SisaKasbon}}...{{end}}`.
6. **Uji manual** dengan data seed (login admin), lalu `go test ./...`, `gofmt -l .`, `go vet ./...`.
7. **Commit, push, buka PR**, minta review.

## 10. Mengubah database / file bersama

File bersama (jangan diubah sembarangan): `cmd/web/main.go`, `internal/handler/routes.go`, `internal/handler/render.go`, `internal/config/`, `internal/middleware/`, `internal/model/`, `internal/service/errors.go`, `internal/service/password.go`, `db/schema.sql`, `templates/layout|header|footer|sidebar.html`, `static/`.

Jika Anda **perlu** mengubahnya:
1. Beri tahu tim di grup terlebih dahulu.
2. Buat PR **terpisah** berjudul `[shared] ...`, jangan dicampur dengan fitur.
3. Untuk perubahan schema: ubah `db/schema.sql`, `internal/model/`, dan `db/seed.sql` dalam **satu** PR berjudul `[shared][schema] ...`.
4. Setelah di-merge, semua anggota `git pull` lalu mengimpor ulang database:
   ```bash
   mysql -u root -p < db/schema.sql
   mysql -u root -p < db/seed.sql
   ```

## 11. Data seed & akun uji

Semua akun memakai password **`password123`** (khusus development):

| Username | Role | Keterangan |
|---|---|---|
| `admin` | admin | Administrator |
| `budi` | pekerja | Harian, upah naik 100.000 → 110.000 pada 28 Okt 2026 |
| `siti` | pekerja | Harian, ada lembur |
| `andi` | pekerja | Harian, ada hari tanpa presensi |
| `dewi` | pekerja | Bulanan, masuk 15 Okt 2026 (uji penyesuaian proporsional) |

Skenario uji penggajian:
- **Periode 1** (mingguan 19–24 Okt): semua hari sudah divalidasi → jalur sukses, berisi lembur dan keterlambatan.
- **Periode 2** (mingguan 26–31 Okt) dan **Periode 3** (bulanan Okt): 31 Okt belum divalidasi → proses penggajian harus **diblokir**.

Tabel skenario lengkap (terlambat > 15 menit, lembur 1,5 jam, presensi ditolak, kasbon aktif/sebagian/lunas, dst.) ada di [`docs/PEMBAGIAN_TUGAS.md`](docs/PEMBAGIAN_TUGAS.md). Untuk mereset data ke kondisi awal, jalankan ulang `schema.sql` lalu `seed.sql`.

## 12. Testing & CI

```bash
gofmt -l .          # harus tidak mengeluarkan apa pun (format -> gofmt -w .)
go vet ./...
go build ./...
go test ./...
```
- Ada test bawaan yang memastikan **semua template valid** (`internal/handler/render_test.go`). Jika gagal, ada file `.html` dengan sintaks salah atau tanpa `{{define "content"}}`.
- Tambahkan unit test untuk logika bisnis di file `_test.go` di sebelah kode (mis. `service/penggajian_test.go`, table-driven).
- **GitHub Actions** (`.github/workflows/ci.yml`) menjalankan gofmt, vet, build, dan test pada setiap push/PR ke `main`. PR tidak boleh di-merge bila CI merah.

## 13. Troubleshooting

| Masalah | Penyebab / solusi |
|---|---|
| `gagal koneksi database: ... connection refused` | MySQL belum jalan, atau `DB_HOST`/`DB_PORT` salah. Nyalakan service MySQL. |
| `Access denied for user 'root'` | `DB_USER`/`DB_PASS` di `.env` tidak cocok dengan MySQL Anda. |
| `Unknown database 'penggajian_mitrajaya'` | `schema.sql` belum diimpor (langkah 3). |
| `go: command not found` / `mysql: command not found` | Belum terpasang atau belum masuk PATH; buka ulang terminal setelah instalasi. |
| `missing go.sum entry` | Jalankan `go mod download`. Jika masih muncul, `git pull` dulu; bila `go.sum` memang belum ada di repo, minta ketua men-commit-nya. |
| `template tidak ditemukan: ...` atau halaman error 500 saat render | Nama template salah (harus path relatif dari `templates/`, mis. `kasbon/list.html`), atau file tidak punya `{{define "content"}}`. |
| Redirect terus ke `/login` | `DEV_AUTO_LOGIN_ROLE` kosong dan modul Auth belum selesai; isi `admin`/`pekerja` di `.env` lalu restart. |
| `403 Akses ditolak` | Role tidak sesuai halaman (mis. membuka `/admin/...` sebagai pekerja). Ganti `DEV_AUTO_LOGIN_ROLE`. |
| `bind: address already in use` | Port 8080 terpakai; ubah `APP_PORT` di `.env`. |
| Perubahan kode Go tidak terlihat | Hentikan (`Ctrl+C`) lalu jalankan ulang `go run ./cmd/web`. Template tidak perlu restart. |
| CI gagal di langkah gofmt | Jalankan `gofmt -w .` lalu commit ulang. |
| Tidak bisa push ke `main` | Memang diproteksi; push ke cabang `feat/...` dan buka PR. |
| `.env` ikut ter-commit | Hapus dari Git: `git rm --cached .env`, commit, dan ganti password yang terlanjur terekspos. |

Masih macet? Salin pesan error lengkap ke grup tim beserta langkah yang sudah dicoba.