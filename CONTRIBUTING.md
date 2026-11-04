# Panduan Kolaborasi Git

## Cabang
- `main` — selalu bisa dijalankan; **dilindungi** (tanpa push langsung, wajib PR + 1 approval + CI hijau).
- Cabang kerja: `feat/<modul>-<deskripsi>`, contoh `feat/validasi-hitung-terlambat`, `fix/absensi-presensi-ganda`.

## Alur harian
```bash
git checkout main && git pull
git checkout -b feat/kasbon-form
# ... kerjakan, commit kecil-kecil ...
git fetch origin && git rebase origin/main   # sebelum push
git push -u origin feat/kasbon-form
# buka Pull Request ke main
```

## Aturan agar tidak bentrok
1. **Hanya edit file modul milikmu** (lihat `docs/PEMBAGIAN_TUGAS.md` dan `.github/CODEOWNERS`).
2. **File bersama (frozen)**: `cmd/web/main.go`, `internal/handler/routes.go`, `internal/handler/render.go`, `internal/config/`, `internal/middleware/`, `internal/model/`, `internal/service/errors.go`, `internal/service/password.go`, `db/schema.sql`, `templates/layout|header|footer|sidebar.html`, `static/`. Butuh perubahan? Buat PR **terpisah** dengan judul `[shared] ...`, beri tahu tim di grup, dan merge cepat.
3. **Perubahan schema**: ubah `db/schema.sql` + `internal/model/` + `db/seed.sql` dalam satu PR `[shared][schema]`; semua anggota wajib re-import database setelah merge.
4. Route baru hanya di bawah prefix modulmu (`/admin/<modul>` atau `/pekerja/<modul>`).
5. Jangan commit `.env` atau kredensial.

## Commit & PR
- Pesan commit: `<tipe>(<modul>): <ringkas>` — tipe: `feat`, `fix`, `refactor`, `test`, `docs`. Contoh: `feat(validasi): hitung menit terlambat saat disetujui`.
- PR kecil (< ~400 baris), isi template PR, hubungkan ke Issue.
- Reviewer memeriksa: aturan bisnis, cek akses di server, transaksi untuk operasi multi-tabel, tidak ada float untuk uang.
