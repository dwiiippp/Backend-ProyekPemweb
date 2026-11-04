# Panduan Kolaborasi Git

## Strategi Cabang

- `main` — cabang utama yang berisi versi stabil dan siap digunakan. Perubahan hanya masuk melalui Pull Request dari `develop`.
- `develop` — cabang integrasi untuk menggabungkan pekerjaan seluruh anggota tim sebelum masuk ke `main`.
- Cabang kerja:
  - `feat/<modul>-<deskripsi>` — untuk pengembangan fitur baru.
  - `fix/<modul>-<deskripsi>` — untuk perbaikan bug.
  - `docs/<deskripsi>` — untuk perubahan dokumentasi.

Contoh:

- `feat/validasi-hitung-terlambat`
- `feat/kasbon-form`
- `fix/absensi-presensi-ganda`
- `docs/update-readme`

## Alur Kerja Harian

### 1. Mengambil versi terbaru develop

```bash
git checkout develop
git pull origin develop
```

### 2. Membuat branch pekerjaan

```bash
git checkout -b feat/kasbon-form
```

### 3. Mengerjakan fitur dan melakukan commit

```bash
git add .
git commit -m "feat(kasbon): tambah form pengajuan kasbon"
```

Lakukan commit secara bertahap sesuai perubahan yang dikerjakan.

### 4. Sinkronisasi dengan develop sebelum push

```bash
git fetch origin
git rebase origin/develop
```

Jika terdapat konflik, selesaikan terlebih dahulu sebelum melanjutkan.

### 5. Push branch ke GitHub

```bash
git push -u origin feat/kasbon-form
```

### 6. Membuat Pull Request

Buat Pull Request dengan ketentuan:

- **Base:** `develop`
- **Compare:** branch pekerjaan masing-masing.
- Pastikan CI berhasil.
- Isi deskripsi perubahan dan cara pengujian.
- Tunggu review sebelum melakukan merge.

## Alur Rilis

Setelah pekerjaan seluruh anggota berhasil diintegrasikan ke `develop`:

1. Pastikan seluruh fitur berjalan dengan baik.
2. Pastikan pemeriksaan CI berhasil.
3. Buat Pull Request dari `develop` menuju `main`.
4. Lakukan review terhadap perubahan.
5. Merge setelah disetujui.

Ketentuan:

- Tidak diperbolehkan mengembangkan fitur langsung di `main`.
- Tidak diperbolehkan mengembangkan fitur langsung di `develop`.
- Semua perubahan harus melalui Pull Request.

## Aturan Agar Tidak Bentrok

1. **Hanya edit file modul milikmu** (lihat `docs/PEMBAGIAN_TUGAS.md` dan `.github/CODEOWNERS`).
2. **File bersama (frozen)**: `cmd/web/main.go`, `internal/handler/routes.go`, `internal/handler/render.go`, `internal/config/`, `internal/middleware/`, `internal/model/`, `internal/service/errors.go`, `internal/service/password.go`, `db/schema.sql`, `templates/layout|header|footer|sidebar.html`, `static/`. Butuh perubahan? Buat PR **terpisah** dengan judul `[shared] ...`, beri tahu tim di grup, dan koordinasikan sebelum merge.
3. **Perubahan schema**: ubah `db/schema.sql` + `internal/model/` + `db/seed.sql` dalam satu PR `[shared][schema]`; semua anggota wajib re-import database setelah merge.
4. Route baru hanya di bawah prefix modulmu (`/admin/<modul>` atau `/pekerja/<modul>`).
5. Jangan commit `.env` atau kredensial.
6. Jangan melakukan perubahan pada branch anggota lain tanpa koordinasi.
7. Jangan melakukan force push ke `main` atau `develop`.

## Commit & Pull Request

- Pesan commit: `<tipe>(<modul>): <ringkas>` — tipe: `feat`, `fix`, `refactor`, `test`, `docs`.
- Contoh: `feat(validasi): hitung menit terlambat saat disetujui`.
- PR kecil (idealnya kurang dari 400 baris), isi template PR, dan hubungkan ke Issue jika tersedia.
- Reviewer memeriksa: aturan bisnis, pengecekan akses di server, transaksi untuk operasi multi-tabel, dan tidak menggunakan `float` untuk uang.

### Checklist Sebelum Merge

- [ ] Kode sudah diformat menggunakan `gofmt`.
- [ ] `go vet ./...` berhasil.
- [ ] `go build ./...` berhasil.
- [ ] `go test ./...` berhasil.
- [ ] Tidak ada file `.env` atau kredensial yang ikut ter-commit.
- [ ] Perubahan tidak merusak modul anggota lain.
