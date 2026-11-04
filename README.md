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
- **Dua role aplikasi:** `admin` (kelola data, validasi, penggajian) dan `pekerja` (lihat jadwal, input presensi, kasbon, slip).
- **Kebutuhan lengkap** ada di dokumen *Detail Requirement – Sistem Penggajian* (use case UC-xx, fitur F-xx, aturan bisnis BR-xx).

Proyek ini sengaja disusun agar tiap anggota bekerja di **file modul masing-masing**. Semua modul sudah terdaftar di router, jadi tidak ada yang perlu mengedit `main.go` atau `routes.go` untuk pekerjaan modul biasa.

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
git config --global user.name "Nama Anda"
git config --global user.email "email-github-anda@contoh.com"
```

## 3. Setup untuk anggota tim

> Ketua tim sudah membuat repo dan mengundang Anda. Terima undangan lewat email atau di https://github.com/notifications. Belum menerima undangan? Hubungi ketua dengan menyebut username GitHub Anda.

**Langkah 1 — Clone repo**

```bash
git clone https://github.com/dwiiippp/Backend-ProyekPemweb.git
cd Backend-ProyekPemweb
```

**Langkah 2 — Pindah ke branch develop**

```bash
git checkout develop
git pull origin develop
```

Seluruh pengembangan dilakukan di branch `develop`. Jangan mengerjakan fitur langsung di `main`.

**Langkah 3 — Buat file `.env`**

```bash
# Windows (CMD/PowerShell)
copy .env.example .env

# macOS / Linux
cp .env.example .env
```

Buka `.env`, lalu sesuaikan `DB_USER` dan `DB_PASS` dengan konfigurasi MySQL di komputer Anda (lihat [bagian 4](#4-konfigurasi-env)).

File `.env` **tidak boleh di-commit** karena berisi konfigurasi dan password lokal. File ini sudah dimasukkan ke `.gitignore`.

**Langkah 4 — Buat database dan tabel**

Untuk database baru atau database development yang datanya boleh di-reset, jalankan:

```bash
mysql -u root -p < db/schema.sql
```

Perintah tersebut membuat database `penggajian_mitrajaya` beserta tabel-tabel yang dibutuhkan.

> **Penting:** `schema.sql` menghapus dan membuat ulang tabel. Jangan menjalankannya pada database yang berisi data penting tanpa melakukan backup terlebih dahulu.

**Langkah 5 — Konfigurasi dan jalankan Go seeder**

Akun pengguna development dibuat melalui Go seeder, bukan melalui `seed.sql`.

Sebelum menjalankan seeder, buka `.env` dan isi konfigurasi berikut:

```env
APP_ENV=development
SEED_USERS_ENABLED=true

SEED_ADMIN_PASSWORD=
SEED_BUDI_PASSWORD=
SEED_SITI_PASSWORD=
SEED_ANDI_PASSWORD=
SEED_DEWI_PASSWORD=
```

Isi setiap variabel password dengan password development masing-masing akun. Ketentuan password:

- Semua password wajib diisi.
- Minimal 8 karakter.
- Password akan di-hash menggunakan bcrypt.
- Password tidak disimpan di repository.

Jalankan seeder:

```bash
go run ./cmd/seed-users
```

Jika berhasil, terminal menampilkan:

```text
Seeder akun berhasil dijalankan.
Total akun diproses: 5
```

Setelah seeder berhasil, ubah kembali konfigurasi berikut di `.env`:

```env
SEED_USERS_ENABLED=false
```

Dengan demikian, seeder tidak dapat dijalankan secara tidak sengaja.

**Langkah 6 — Impor data contoh**

Setelah akun pengguna berhasil dibuat, impor data contoh:

```bash
mysql -u root -p < db/seed.sql
```

Urutan ini wajib dipertahankan:

1. `db/schema.sql`
2. `go run ./cmd/seed-users`
3. `db/seed.sql`

`seed.sql` berisi data pekerja dan skenario pengujian yang menggunakan akun hasil Go seeder.

> **Khusus PowerShell:** Jika perintah redirect `<` tidak berjalan, gunakan:
>
> ```powershell
> Get-Content db/schema.sql | mysql -u root -p
> go run ./cmd/seed-users
> Get-Content db/seed.sql | mysql -u root -p
> ```

**Langkah 7 — Unduh dependensi Go**

```bash
go mod download
```

**Langkah 8 — Jalankan aplikasi**

```bash
go run ./cmd/web
```

Buka http://localhost:8080.

Untuk development, Anda dapat mengisi `DEV_AUTO_LOGIN_ROLE=admin` agar langsung masuk sebagai admin tanpa login. Untuk menguji autentikasi sungguhan, kosongkan variabel tersebut.

**Langkah 9 — Pastikan semuanya sehat**

```bash
go test ./...
```

Jika semua pengujian menunjukkan `ok`, lingkungan development Anda siap.

Template HTML dimuat ulang otomatis di mode development, jadi mengubah file `.html` cukup dengan me-refresh browser. Perubahan kode Go perlu menjalankan ulang `go run ./cmd/web`.

### Checklist siap mulai

Centang semuanya sebelum mengambil Issue pertama. Jika ada yang gagal, lihat [Troubleshooting](#13-troubleshooting) atau tanya di grup.

- [ ] `go version` menampilkan 1.22 atau lebih baru, serta `mysql --version` dan `git --version` berjalan.
- [ ] Repo sudah di-clone dan berada di branch `develop`.
- [ ] File `.env` sudah dibuat dari `.env.example`.
- [ ] Database dan tabel berhasil dibuat melalui `db/schema.sql`.
- [ ] Lima akun development berhasil dibuat melalui `go run ./cmd/seed-users`.
- [ ] Data contoh berhasil diimpor melalui `db/seed.sql`.
- [ ] `SEED_USERS_ENABLED` sudah dikembalikan menjadi `false`.
- [ ] Jumlah data pekerja berjumlah 4 (cek: `mysql -u root -p -e "SELECT COUNT(*) FROM penggajian_mitrajaya.pekerja;"`).
- [ ] `go test ./...` semuanya `ok`.
- [ ] http://localhost:8080 dapat diakses.
- [ ] Anda sudah mencoba `DEV_AUTO_LOGIN_ROLE=pekerja` untuk melihat menu pekerja, lalu mengembalikannya sesuai kebutuhan.
- [ ] Sudah memahami cara mengambil perubahan terbaru, commit, dan push ke `develop` sesuai bagian 7.
- [ ] Sudah membaca [`CONTRIBUTING.md`](CONTRIBUTING.md) dan bagian anggota Anda di [`docs/PEMBAGIAN_TUGAS.md`](docs/PEMBAGIAN_TUGAS.md).
- [ ] Ada Issue yang di-assign ke Anda di tab *Issues* repo.

## 4. Konfigurasi `.env`

Semua konfigurasi lokal disimpan di `.env`. Gunakan `.env.example` sebagai template awal.

| Variabel | Default | Keterangan |
|---|---|---|
| `APP_ENV` | `development` | Environment aplikasi. Seeder dan dev auto-login hanya boleh digunakan pada development. |
| `APP_PORT` | `8080` | Port server |
| `DB_HOST` | `127.0.0.1` | Host MySQL |
| `DB_PORT` | `3306` | Port MySQL |
| `DB_USER` | `root` | User MySQL |
| `DB_PASS` | *(kosong)* | Password MySQL |
| `DB_NAME` | `penggajian_mitrajaya` | Nama database |
| `DEV_AUTO_LOGIN_ROLE` | *(kosong)* | `admin` / `pekerja` untuk melewati login saat development. Kosongkan untuk menguji login sungguhan. |
| `SEED_USERS_ENABLED` | `false` | Harus `true` untuk mengaktifkan Go seeder. Khusus development. |
| `SEED_ADMIN_PASSWORD` | *(kosong)* | Password akun `admin` |
| `SEED_BUDI_PASSWORD` | *(kosong)* | Password akun `budi` |
| `SEED_SITI_PASSWORD` | *(kosong)* | Password akun `siti` |
| `SEED_ANDI_PASSWORD` | *(kosong)* | Password akun `andi` |
| `SEED_DEWI_PASSWORD` | *(kosong)* | Password akun `dewi` |

### Tentang dev auto-login

Jika `DEV_AUTO_LOGIN_ROLE` terisi, setiap request dianggap sudah login sebagai user development:

- `admin` = akun admin.
- `pekerja` = akun pekerja Budi.

Fitur ini hanya aktif saat `APP_ENV=development`.

Untuk menguji tampilan pekerja, ganti nilainya menjadi `pekerja`, lalu restart aplikasi. Untuk menguji autentikasi sungguhan, kosongkan `DEV_AUTO_LOGIN_ROLE`.

### Tentang Go seeder

Go seeder digunakan untuk membuat lima akun development.

Ketentuannya:

- Hanya dapat dijalankan ketika `APP_ENV=development`.
- Harus mengaktifkan `SEED_USERS_ENABLED=true`.
- Semua variabel password akun wajib diisi.
- Password minimal 8 karakter.
- Password disimpan dalam bentuk hash bcrypt.

Setelah proses seeding selesai, kembalikan `SEED_USERS_ENABLED=false`.

> Jangan membagikan file `.env` atau memasukkan password akun ke repository. Setiap anggota tim dapat menggunakan password development masing-masing.

## 5. Struktur proyek

```text
Backend-ProyekPemweb/
├── cmd/
│   ├── web/main.go             # entry point aplikasi [bersama/frozen]
│   └── seed-users/main.go      # Go seeder akun development
├── db/
│   ├── schema.sql              # skema MySQL [bersama/frozen]
│   └── seed.sql                # data contoh + skenario uji
├── internal/
│   ├── config/                 # env & koneksi DB [frozen]
│   ├── handler/                # HTTP handler per modul; routes.go & render.go [frozen]
│   ├── middleware/             # session, RBAC, logging, recover [frozen]
│   ├── model/                  # struct tabel [frozen]
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
├── .env.example
├── .gitignore
└── go.mod
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

- Schema, model, dan seed sudah disiapkan agar semua anggota menggunakan data yang sama sejak hari pertama.
- Modul lain cukup *membaca* tabel milik modul lain lewat repository miliknya sendiri; tidak saling memanggil service.
- Login bisa dilewati lewat `DEV_AUTO_LOGIN_ROLE`.
- Anggota 3 bisa menulis perhitungan gaji sebagai fungsi murni + unit test memakai angka dari seed, tanpa menunggu presensi selesai.

**Langkah pertama setelah setup (semua anggota)**

1. Buka Issue pertama Anda di tab *Issues*.
2. Pastikan berada di branch `develop` dan sudah mengambil perubahan terbaru.
3. Jalankan server dan buka halaman stub modul Anda lewat sidebar (ditandai *STUB · Anggota N*) untuk melihat kondisi awalnya.
4. Baca struct terkait di `internal/model/` dan data di `db/seed.sql` untuk modul Anda.
5. Kerjakan berurutan: repository → service → handler → template (lihat contoh di [bagian 9](#9-contoh-mengerjakan-satu-fitur-dari-awal)).
6. Uji manual dengan data seed, jalankan `gofmt -w .`, `go vet ./...`, dan `go test ./...`.
7. Commit dan push perubahan ke `develop` sesuai [bagian 7](#7-alur-kerja-git-harian).

## 7. Alur kerja Git harian

Repository ini hanya menggunakan dua branch:

- `main` — menyimpan versi stabil aplikasi.
- `develop` — digunakan seluruh anggota untuk mengembangkan dan mengintegrasikan fitur.

### Langkah 1 — Mengambil versi terbaru

Sebelum mulai bekerja, jalankan:

```bash
git checkout develop
git pull origin develop
```

### Langkah 2 — Mengerjakan modul

Setiap anggota mengerjakan modul sesuai pembagian tugas masing-masing langsung pada branch `develop`.

Hindari mengubah file yang sedang dikerjakan anggota lain tanpa koordinasi. Sebelum mengedit file bersama, beri tahu anggota tim terlebih dahulu.

### Langkah 3 — Melakukan commit

Setelah menyelesaikan perubahan:

```bash
git add .
git commit -m "feat(absensi): implementasi input presensi"
```

Gunakan pesan commit yang menjelaskan perubahan.

Contoh tipe commit:

- `feat` — menambahkan fitur.
- `fix` — memperbaiki bug.
- `refactor` — merapikan struktur kode.
- `test` — menambahkan atau memperbaiki pengujian.
- `docs` — mengubah dokumentasi.

### Langkah 4 — Sinkronisasi dan push

Sebelum push, ambil perubahan terbaru:

```bash
git pull --rebase origin develop
```

Kemudian:

```bash
git push origin develop
```

Jika terjadi konflik, selesaikan terlebih dahulu sebelum melanjutkan. Jangan melakukan force push.

Jika proses rebase mengalami konflik:

```bash
# Periksa file yang mengalami konflik
git status

# Setelah konflik diselesaikan
git add .
git rebase --continue
```

Jika ingin membatalkan proses rebase:

```bash
git rebase --abort
```

### Langkah 5 — Rilis ke main

Setelah fitur selesai dan sudah diuji pada `develop`:

1. Pastikan seluruh perubahan sudah terintegrasi.
2. Pastikan CI berhasil.
3. Buat Pull Request dari `develop` menuju `main`.
4. Lakukan pemeriksaan dan review.
5. Merge setelah disetujui.

### Aturan penting

- Hanya gunakan branch `main` dan `develop`.
- Jangan mengembangkan fitur langsung di `main`.
- Semua pekerjaan pengembangan dilakukan di `develop`.
- Jangan membuat branch tambahan untuk pekerjaan fitur, perbaikan, atau dokumentasi.
- Jangan melakukan force push.
- Selalu sinkronkan perubahan sebelum melakukan push.
- Koordinasikan perubahan pada file yang digunakan bersama.
- Jangan melakukan commit terhadap file `.env`, password, atau kredensial.

## 8. Panduan menulis kode

**Alur request**

```text
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

1. **Issue:** ambil Issue "Kasbon – daftar admin".
2. **Sinkronisasi:** pastikan berada di branch `develop`, lalu jalankan:

   ```bash
   git checkout develop
   git pull origin develop
   ```

3. **Repository** (`internal/repository/kasbon.go`):

   ```go
   func (r *KasbonRepository) List(ctx context.Context) ([]model.Kasbon, error) {
       rows, err := r.db.QueryContext(ctx,
           `SELECT id, pekerja_id, tanggal_pengajuan, nominal_kasbon, sisa_kasbon, keterangan, status, dibuat_pada
            FROM kasbon ORDER BY tanggal_pengajuan DESC`)
       if err != nil {
           return nil, err
       }
       defer rows.Close()

       var out []model.Kasbon
       for rows.Next() {
           var k model.Kasbon
           if err := rows.Scan(&k.ID, &k.PekerjaID, &k.TanggalPengajuan, &k.NominalKasbon,
               &k.SisaKasbon, &k.Keterangan, &k.Status, &k.DibuatPada); err != nil {
               return nil, err
           }
           out = append(out, k)
       }

       return out, rows.Err()
   }
   ```

4. **Service** (`internal/service/kasbon.go`): tambahkan `func (s *KasbonService) Daftar(ctx context.Context) ([]model.Kasbon, error) { return s.repo.List(ctx) }`.
5. **Handler** (`internal/handler/kasbon.go`): di method `List`, panggil `h.svc.Daftar(r.Context())`; bila error, `h.d.Render.Error(w, r, 500, "Gagal memuat data")`; bila sukses, render `kasbon/list.html` dengan `Data: daftar`.
6. **Template** (`templates/kasbon/list.html`): tabel dengan `{{range .Data}}...{{rupiah .SisaKasbon}}...{{end}}`.
7. **Uji manual** dengan data seed (login admin), lalu jalankan:

   ```bash
   gofmt -w .
   go vet ./...
   go test ./...
   go build ./...
   ```

8. **Commit dan push:** commit perubahan, sinkronkan dengan `develop`, lalu push ke `develop` sesuai [bagian 7](#7-alur-kerja-git-harian).

## 10. Mengubah database / file bersama

File bersama (jangan diubah sembarangan):

- `cmd/web/main.go`
- `cmd/seed-users/main.go`
- `internal/handler/routes.go`
- `internal/handler/render.go`
- `internal/config/`
- `internal/middleware/`
- `internal/model/`
- `internal/service/errors.go`
- `internal/service/password.go`
- `db/schema.sql`
- `templates/layout.html`
- `templates/header.html`
- `templates/footer.html`
- `templates/sidebar.html`
- `static/`

Jika Anda **perlu** mengubahnya:

1. Beri tahu tim di grup terlebih dahulu.
2. Jelaskan file yang akan diubah dan alasan perubahannya.
3. Koordinasikan waktu pengerjaan agar tidak ada dua anggota yang mengedit file yang sama secara bersamaan.
4. Untuk perubahan schema, koordinasikan perubahan `db/schema.sql`, `internal/model/`, dan `db/seed.sql` agar konsisten.
5. Setelah perubahan masuk ke `develop`, seluruh anggota melakukan:

   ```bash
   git checkout develop
   git pull origin develop
   ```

6. Jika schema berubah dan database perlu dibuat ulang, ikuti urutan setup berikut:

   ```bash
   mysql -u root -p < db/schema.sql
   go run ./cmd/seed-users
   mysql -u root -p < db/seed.sql
   ```

> **Penting:** Menjalankan ulang `schema.sql` akan menghapus dan membuat ulang tabel. Pastikan data lokal yang dibutuhkan sudah dicadangkan sebelum mengimpor ulang. Pastikan konfigurasi password seeder sudah diisi dan `SEED_USERS_ENABLED=true` sebelum menjalankan seeder.

## 11. Data seed & akun uji

Data akun pengguna dibuat menggunakan Go seeder:

```bash
go run ./cmd/seed-users
```

Password setiap akun dikonfigurasi secara terpisah melalui variabel `SEED_*_PASSWORD` di `.env`. Password tidak disimpan di repository.

Lima akun development yang dibuat:

| Username | Role | Keterangan |
|---|---|---|
| `admin` | admin | Administrator |
| `budi` | pekerja | Harian, upah naik 100.000 → 110.000 pada 28 Okt 2026 |
| `siti` | pekerja | Harian, ada lembur |
| `andi` | pekerja | Harian, ada hari tanpa presensi |
| `dewi` | pekerja | Bulanan, masuk 15 Okt 2026 (uji penyesuaian proporsional) |

**Ketentuan seeder:**

- Hanya dapat dijalankan pada `APP_ENV=development`.
- Harus mengatur `SEED_USERS_ENABLED=true`.
- Semua password akun wajib diisi dan minimal 8 karakter.
- Password disimpan dalam bentuk hash bcrypt.
- Setelah seeder berhasil, kembalikan `SEED_USERS_ENABLED=false`.

> Menjalankan seeder kembali akan memperbarui data akun berdasarkan konfigurasi password dan role di `.env`. Untuk mengembalikan seluruh database ke kondisi awal, jalankan ulang `schema.sql`, Go seeder, lalu `seed.sql` sesuai urutan setup.

### Skenario uji penggajian

- **Periode 1** (mingguan 19–24 Okt): semua hari sudah divalidasi → jalur sukses, berisi lembur dan keterlambatan.
- **Periode 2** (mingguan 26–31 Okt) dan **Periode 3** (bulanan Okt): 31 Okt belum divalidasi → proses penggajian harus **diblokir**.

Tabel skenario lengkap (terlambat > 15 menit, lembur 1,5 jam, presensi ditolak, kasbon aktif/sebagian/lunas, dst.) ada di [`docs/PEMBAGIAN_TUGAS.md`](docs/PEMBAGIAN_TUGAS.md).

## 12. Testing & CI

Jalankan pemeriksaan berikut sebelum melakukan commit:

```bash
gofmt -l .
go vet ./...
go build ./...
go test ./...
```

Jika `gofmt -l .` menampilkan file, format terlebih dahulu:

```bash
gofmt -w .
```

- Ada test bawaan yang memastikan **semua template valid** (`internal/handler/render_test.go`). Jika gagal, ada file `.html` dengan sintaks salah atau tanpa `{{define "content"}}`.
- Tambahkan unit test untuk logika bisnis di file `_test.go` di sebelah kode (mis. `service/penggajian_test.go`, table-driven).
- **GitHub Actions** (`.github/workflows/ci.yml`) menjalankan gofmt, vet, build, dan test pada setiap push ke `main` atau `develop`, serta setiap Pull Request yang menargetkan `main` atau `develop`.
- Sebelum rilis dari `develop` ke `main`, pastikan seluruh pemeriksaan CI berhasil.

## 13. Troubleshooting

| Masalah | Penyebab / solusi |
|---|---|
| `gagal koneksi database: ... connection refused` | MySQL belum jalan, atau `DB_HOST`/`DB_PORT` salah. Nyalakan service MySQL. |
| `Access denied for user 'root'` | `DB_USER`/`DB_PASS` di `.env` tidak cocok dengan MySQL Anda. |
| `Unknown database 'penggajian_mitrajaya'` | `schema.sql` belum diimpor (langkah 4). |
| Seeder menolak dijalankan | Pastikan `APP_ENV=development` dan `SEED_USERS_ENABLED=true`. |
| Seeder memberi pesan password belum diisi | Pastikan kelima variabel `SEED_*_PASSWORD` sudah diisi di `.env`. |
| Password seeder tidak valid | Gunakan password minimal 8 karakter untuk setiap akun. |
| `go: command not found` / `mysql: command not found` | Belum terpasang atau belum masuk PATH; buka ulang terminal setelah instalasi. |
| `missing go.sum entry` | Jalankan `go mod download`. Jika masih muncul, `git pull` dulu; bila `go.sum` memang belum ada di repo, minta ketua men-commit-nya. |
| `template tidak ditemukan: ...` atau halaman error 500 saat render | Nama template salah (harus path relatif dari `templates/`, mis. `kasbon/list.html`), atau file tidak punya `{{define "content"}}`. |
| Redirect terus ke `/login` | `DEV_AUTO_LOGIN_ROLE` kosong dan modul Auth belum selesai; isi `admin`/`pekerja` di `.env` lalu restart. |
| `403 Akses ditolak` | Role tidak sesuai halaman (mis. membuka `/admin/...` sebagai pekerja). Ganti `DEV_AUTO_LOGIN_ROLE`. |
| `bind: address already in use` | Port 8080 terpakai; ubah `APP_PORT` di `.env`. |
| Perubahan kode Go tidak terlihat | Hentikan (`Ctrl+C`) lalu jalankan ulang `go run ./cmd/web`. Template tidak perlu restart. |
| CI gagal di langkah gofmt | Jalankan `gofmt -w .` lalu commit ulang. |
| Tidak bisa push ke `main` | `main` digunakan untuk versi stabil. Lakukan pengembangan di `develop`, lalu ajukan PR dari `develop` ke `main` saat rilis. |
| Push ke `develop` ditolak | Pastikan sudah menjalankan `git pull --rebase origin develop`, periksa konflik, dan pastikan tidak ada perubahan yang belum diselesaikan. |
| `.env` ikut ter-commit | Hapus dari Git: `git rm --cached .env`, commit, dan ganti password yang terlanjur terekspos. |

---

**Catatan workflow:** Repository ini menggunakan dua branch saja: `main` dan `develop`. Seluruh pekerjaan anggota tim dilakukan langsung di `develop`. Pull Request hanya digunakan untuk proses rilis dari `develop` menuju `main`.