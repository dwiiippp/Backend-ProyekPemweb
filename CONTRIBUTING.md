# Panduan Kolaborasi Git

Dokumen ini menjadi pedoman kolaborasi seluruh anggota tim backend dalam mengembangkan Sistem Penggajian Konveksi Mitra Jaya.

## 1. Strategi Branch

Repository ini hanya menggunakan dua branch:

- `main` — branch utama yang berisi versi stabil aplikasi. Perubahan masuk melalui Pull Request dari `develop`.
- `develop` — branch pengembangan bersama yang digunakan seluruh anggota untuk mengerjakan dan mengintegrasikan fitur.

**Aturan utama:**
- Hanya gunakan branch `main` dan `develop`.
- Seluruh pekerjaan fitur, perbaikan bug, dan dokumentasi dilakukan langsung di `develop`.
- Tidak perlu membuat branch tambahan untuk setiap pekerjaan.
- Jangan melakukan pengembangan langsung di `main`.

## 2. Alur Kerja Harian

### 1. Mengambil versi terbaru

Sebelum mulai bekerja, pastikan berada di branch `develop`:

```bash
git checkout develop
git pull origin develop
```

Tujuannya agar pekerjaan dimulai dari kode terbaru yang sudah diintegrasikan anggota tim lainnya.

### 2. Mengerjakan modul

Setiap anggota mengerjakan modul sesuai pembagian tugas masing-masing.

| Anggota | Fokus | Modul |
|---|---|---|
| Anggota 1 | Fondasi & Data Master | Auth, Home, Pekerja, Upah, Shift, Jadwal |
| Anggota 2 | Presensi & Kasbon | Absensi, Validasi, Kasbon, Laporan Presensi/Kasbon |
| Anggota 3 | Penggajian & Laporan | Periode, Penggajian, Koreksi, Slip, Laporan Gaji |

Aturan pengerjaan:

- Kerjakan file yang menjadi tanggung jawab modul Anda.
- Hindari mengedit file yang sedang dikerjakan anggota lain tanpa koordinasi.
- Beri tahu anggota tim jika membutuhkan perubahan pada file bersama.
- Jangan mengubah implementasi modul anggota lain tanpa persetujuan.

### 3. Melakukan commit

Setelah menyelesaikan perubahan, lakukan commit:

```bash
git add .
git commit -m "feat(kasbon): tambah form pengajuan kasbon"
```

Lakukan commit secara bertahap berdasarkan perubahan yang sudah selesai.

Gunakan format pesan commit:

```text
<tipe>(<modul>): <deskripsi singkat>
```

Jenis commit:

- `feat` — menambahkan fitur.
- `fix` — memperbaiki bug.
- `refactor` — merapikan struktur kode.
- `test` — menambahkan atau memperbaiki pengujian.
- `docs` — mengubah dokumentasi.

Contoh:

```bash
git commit -m "feat(absensi): tambah fitur input presensi"
git commit -m "fix(kasbon): perbaiki validasi nominal"
git commit -m "docs(readme): perbarui panduan setup"
```

### 4. Sinkronisasi sebelum push

Sebelum melakukan push, ambil perubahan terbaru dari `develop`:

```bash
git pull --rebase origin develop
```

Jika tidak ada konflik, lanjutkan dengan push:

```bash
git push origin develop
```

Jika terjadi konflik saat rebase:

```bash
# Periksa file yang mengalami konflik
git status
```

Buka file yang mengalami konflik, selesaikan bagian yang bertentangan, kemudian:

```bash
git add .
git rebase --continue
```

Jika ingin membatalkan proses rebase:

```bash
git rebase --abort
```

**Penting:**
- Jangan melakukan force push.
- Jangan menghapus commit anggota lain.
- Jika push ditolak karena ada perubahan terbaru, lakukan sinkronisasi kembali sebelum push.
- Pastikan perubahan lokal sudah disimpan melalui commit sebelum melakukan rebase.

## 3. Alur Rilis ke Main

Pull Request hanya digunakan ketika akan merilis perubahan dari `develop` ke `main`.

Tahapan rilis:

1. Pastikan seluruh pekerjaan anggota sudah terintegrasi di `develop`.
2. Pastikan aplikasi sudah diuji.
3. Jalankan pemeriksaan kode dan pastikan CI berhasil.
4. Buat Pull Request dengan ketentuan:
   - **Base:** `main`
   - **Compare:** `develop`
5. Lakukan pemeriksaan dan review terhadap perubahan.
6. Merge setelah disetujui.

**Ketentuan rilis:**

- Jangan mengembangkan fitur langsung di `main`.
- Jangan melakukan merge ke `main` tanpa pemeriksaan.
- Pastikan `develop` dalam kondisi stabil sebelum dirilis.
- Setelah rilis, anggota tetap melanjutkan pengembangan di `develop`.

## 4. Aturan Agar Tidak Bentrok

Karena seluruh anggota bekerja pada branch `develop` yang sama, koordinasi menjadi hal penting.

### 4.1 Pembagian file

Setiap anggota bertanggung jawab terhadap file sesuai modul yang telah dibagikan.

Lihat:
- `docs/PEMBAGIAN_TUGAS.md`
- `.github/CODEOWNERS`

Hindari mengubah file modul milik anggota lain tanpa koordinasi.

### 4.2 File bersama (frozen)

File berikut merupakan fondasi bersama dan tidak boleh diubah sembarangan:

- `cmd/web/main.go`
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

Jika perlu mengubah file bersama:

1. Beri tahu seluruh anggota tim terlebih dahulu.
2. Jelaskan bagian yang akan diubah dan alasannya.
3. Koordinasikan waktu pengerjaan agar tidak ada dua anggota mengedit file yang sama secara bersamaan.
4. Setelah perubahan selesai, lakukan commit dan push ke `develop`.
5. Informasikan kepada anggota lain agar melakukan sinkronisasi.

### 4.3 Perubahan database

Jika terdapat perubahan struktur database:

- Koordinasikan perubahan `db/schema.sql`.
- Sesuaikan `internal/model/` jika struktur tabel berubah.
- Sesuaikan `db/seed.sql` jika data contoh perlu diperbarui.
- Pastikan ketiga bagian tersebut konsisten.

Setelah perubahan masuk ke `develop`, seluruh anggota harus mengambil perubahan terbaru:

```bash
git checkout develop
git pull origin develop
```

Jika schema berubah, anggota perlu mengimpor ulang database:

```bash
mysql -u root -p < db/schema.sql
mysql -u root -p < db/seed.sql
```

> **Perhatian:** Menjalankan `schema.sql` akan menghapus dan membuat ulang tabel. Cadangkan data lokal yang masih dibutuhkan sebelum menjalankannya.

### 4.4 Aturan tambahan

1. Route baru hanya boleh berada di bawah prefix modul masing-masing (`/admin/<modul>` atau `/pekerja/<modul>`).
2. Jangan melakukan commit terhadap `.env` atau kredensial.
3. Jangan mengubah pekerjaan anggota lain tanpa koordinasi.
4. Selalu sinkronkan branch sebelum mulai bekerja dan sebelum push.
5. Jangan melakukan force push ke `main` maupun `develop`.

## 5. Standar Commit dan Review

Gunakan pesan commit yang jelas dan konsisten.

Contoh:

```text
feat(validasi): hitung menit terlambat saat disetujui
feat(penggajian): implementasi perhitungan upah
fix(absensi): perbaiki validasi presensi ganda
test(kasbon): tambah unit test pengajuan kasbon
docs(readme): perbarui panduan instalasi
```

Saat melakukan review rilis dari `develop` ke `main`, perhatikan:

- Kesesuaian implementasi dengan aturan bisnis.
- Validasi input dari pengguna.
- Pengecekan hak akses pada server.
- Penggunaan transaksi untuk operasi multi-tabel.
- Penggunaan `decimal.Decimal` untuk perhitungan uang, bukan `float`.
- Tidak adanya file `.env` atau kredensial di repository.
- Tidak adanya perubahan yang merusak modul anggota lain.

## 6. Checklist Sebelum Rilis

Pastikan seluruh pemeriksaan berikut berhasil sebelum membuat atau menyetujui Pull Request `develop` ke `main`.

- [ ] Kode sudah diformat menggunakan `gofmt`.
- [ ] `go vet ./...` berhasil.
- [ ] `go build ./...` berhasil.
- [ ] `go test ./...` berhasil.
- [ ] Aplikasi sudah diuji secara manual.
- [ ] Tidak ada file `.env` atau kredensial yang ikut ter-commit.
- [ ] Perubahan tidak merusak modul anggota lain.
- [ ] Seluruh pekerjaan anggota sudah terintegrasi di `develop`.
- [ ] GitHub Actions CI berhasil.

---

**Ringkasan workflow tim:**

```text
              main
               ▲
               │
          Pull Request
               │
             develop
          ┌────┼────┐
          │    │    │
       Anggota Anggota Anggota
          1      2      3
          │    │    │
          └────┼────┘
           Integrasi
```

Seluruh anggota bekerja langsung pada `develop`. Branch `main` hanya digunakan untuk menyimpan versi stabil aplikasi.