# Pembagian Tugas Tim Backend (3 Anggota)

Proyek: Sistem Penggajian Pekerja Harian dan Shift — Konveksi Mitra Jaya (Go + MySQL).
Referensi kebutuhan: dokumen *Detail Requirement – Sistem Penggajian*.

## Prinsip: bekerja paralel tanpa saling menunggu

1. **Schema & model dibekukan** (`db/schema.sql`, `internal/model/`) — semua orang memakai struktur data yang sama sejak hari pertama.
2. **Data contoh siap pakai** (`db/seed.sql`) — Anggota 3 bisa menguji perhitungan gaji tanpa menunggu Anggota 2 menyelesaikan presensi.
3. **Login bisa dilewati saat development** — set `DEV_AUTO_LOGIN_ROLE=admin` atau `pekerja` di `.env`, jadi tak perlu menunggu modul Auth.
4. **Setiap modul punya file sendiri** (handler / service / repository / templates) dan **sudah terdaftar** di `routes.go`. Tidak ada yang perlu mengedit `main.go` atau `routes.go`, sehingga merge conflict hampir nol.
5. **Kontrak lintas modul = tabel database** (matriks di bawah). Modul lain cukup *membaca* tabel itu lewat repository miliknya sendiri; tidak saling memanggil service.

## Anggota 1 — Fondasi & Data Master
| Modul | Fitur / Use case | File milik |
|---|---|---|
| `auth` | F-01 · UC-01 Login | `handler/auth.go` `service/auth.go` `repository/auth.go` |
| `home` | UC-02 · F-13 Dashboard & profil | `handler/home.go` `service/home.go` `repository/home.go` |
| `pekerja` | F-02 · UC-08 Kelola Pekerja | `handler/pekerja.go` `service/pekerja.go` `repository/pekerja.go` `templates/pekerja/list.html` `templates/pekerja/form.html` `templates/pekerja/detail.html` |
| `upah` | F-03 Riwayat Upah (BR-04) | `handler/upah.go` `service/upah.go` `repository/upah.go` `templates/upah/list.html` `templates/upah/form.html` |
| `shift` | F-04 Kelola Shift (BR-05) | `handler/shift.go` `service/shift.go` `repository/shift.go` `templates/shift/list.html` `templates/shift/form.html` |
| `jadwal` | F-05 Jadwal Kerja · UC-03 | `handler/jadwal.go` `service/jadwal.go` `repository/jadwal.go` `templates/jadwal/kalender.html` `templates/jadwal/form.html` `templates/jadwal/saya.html` |

- [ ] **Auth** — login (bcrypt, tolak akun nonaktif), logout, session, redirect per role (F-01 / UC-01). `templates/login.html` sudah ada form dasarnya.
- [ ] **Pekerja** — CRUD pekerja + akun login dalam satu transaksi (F-02 / UC-08). Hash password pakai `service.HashPassword`.
- [ ] **Upah** — tambah upah baru + tanggal berlaku, jangan menimpa baris lama (F-03 / BR-04).
- [ ] **Shift** — CRUD shift (F-04 / BR-05).
- [ ] **Jadwal** — kalender mingguan admin, satu pekerja satu jadwal per hari (F-05); jadwal milik sendiri untuk pekerja (UC-03).
- [ ] **Home & Profil** — dashboard per role, profil pekerja (UC-02).
- [ ] Rapikan pesan error/flash agar konsisten dengan `PageData.Flash` / `PageData.Error`.

## Anggota 2 — Presensi & Kasbon
| Modul | Fitur / Use case | File milik |
|---|---|---|
| `absensi` | F-06 · UC-04 Input Presensi · UC-05 Riwayat (BR-13, BR-15) | `handler/absensi.go` `service/absensi.go` `repository/absensi.go` `templates/absensi/form.html` `templates/absensi/riwayat.html` |
| `validasi` | F-14 · UC-10 Validasi Presensi Harian (BR-14) | `handler/validasi.go` `service/validasi.go` `repository/validasi.go` `templates/validasi/list.html` `templates/validasi/detail.html` |
| `kasbon` | F-08 · UC-11 Kelola Kasbon · UC-07 Riwayat Kasbon | `handler/kasbon.go` `service/kasbon.go` `repository/kasbon.go` `templates/kasbon/list.html` `templates/kasbon/form.html` `templates/kasbon/saya.html` |
| `laporan_presensi_kasbon` | F-12 Laporan presensi & rekap kasbon | `handler/laporan_presensi_kasbon.go` `service/laporan_presensi_kasbon.go` `repository/laporan_presensi_kasbon.go` `templates/laporan/presensi.html` `templates/laporan/kasbon.html` |

- [ ] **Absensi (pekerja)** — input presensi (UC-04, BR-13, BR-15): tolak bila hari sudah ditutup, periode ditutup, presensi ganda, jam keluar < jam keluar shift, atau bukan jadwal sendiri; status awal `menunggu`.
- [ ] **Riwayat presensi** — status validasi + catatan penolakan (UC-05).
- [ ] **Validasi harian (admin)** — list per tanggal, setujui/tolak (catatan wajib saat tolak), hitung `menit_terlambat` & `jam_lembur`, insert `validasi_presensi_harian` — satu transaksi (F-14 / UC-10 / BR-14).
- [ ] **Kasbon** — catat pengajuan (sisa = nominal, status aktif), daftar admin, riwayat pekerja (F-08 / UC-11 / UC-07).
- [ ] **Laporan** — laporan presensi & rekap kasbon (F-12).

## Anggota 3 — Penggajian & Laporan
| Modul | Fitur / Use case | File milik |
|---|---|---|
| `periode` | Periode Gaji (BR-01, BR-02, BR-10) | `handler/periode.go` `service/periode.go` `repository/periode.go` `templates/periode/list.html` `templates/periode/form.html` `templates/periode/detail.html` |
| `penggajian` | F-07 · F-09 · UC-12 Proses Penggajian · UC-13 Tutup Periode | `handler/penggajian.go` `service/penggajian.go` `repository/penggajian.go` `templates/penggajian/index.html` `templates/penggajian/pratinjau.html` `templates/penggajian/hasil.html` |
| `koreksi` | F-10 · UC-14 Koreksi Gaji (BR-11) | `handler/koreksi.go` `service/koreksi.go` `repository/koreksi.go` `templates/koreksi/list.html` `templates/koreksi/form.html` |
| `slip` | F-11 · UC-06 Lihat Slip · UC-15 Cetak Slip (BR-12) | `handler/slip.go` `service/slip.go` `repository/slip.go` `templates/slip/list.html` `templates/slip/detail.html` `templates/slip/saya.html` |
| `laporan_gaji` | F-12 Laporan status periode & rekap biaya gaji/lembur | `handler/laporan_gaji.go` `service/laporan_gaji.go` `repository/laporan_gaji.go` `templates/laporan/periode.html` `templates/laporan/rekap_gaji.html` |

- [ ] **Periode** — buat/lihat periode mingguan (Senin–Sabtu) & bulanan, status terbuka/ditutup (BR-01, BR-02, BR-10).
- [ ] **Mesin perhitungan** — fungsi murni + unit test berdasarkan rumus F-07 & BR-03–BR-09 (lihat komentar `service/penggajian.go`).
- [ ] **Proses & pratinjau** — prasyarat semua hari periode tervalidasi; tampilkan daftar kesalahan bila tidak valid (UC-12).
- [ ] **Konfirmasi & tutup periode** — SATU transaksi: `gaji`, `potongan_kasbon_gaji`, update `kasbon` (sisa/lunas), `periode_gaji` ditutup (UC-13 / F-09).
- [ ] **Koreksi gaji** — tambah/kurang ke periode tujuan yang terbuka (F-10 / BR-11).
- [ ] **Slip** — admin cetak, pekerja hanya slip miliknya (F-11 / UC-06 / UC-15).
- [ ] **Laporan** — status periode, rekap biaya gaji & lembur (F-12 / BR-12).

## Matriks kontrak tabel (W = tulis, R = baca)

| Tabel | Anggota 1 | Anggota 2 | Anggota 3 |
|---|---|---|---|
| users | W | R | R |
| pekerja | W | R | R |
| riwayat_gaji | W | – | R |
| shift, jadwal_kerja | W | R | R |
| absensi | – | W | R (hanya `disetujui`) |
| validasi_presensi_harian | – | W | R |
| kasbon | – | W (buat) | W (kurangi sisa saat tutup periode) |
| periode_gaji | – | R (cek periode ditutup) | W |
| gaji, potongan_kasbon_gaji, koreksi_gaji | – | – | W |

## Skenario data seed untuk pengujian

| Skenario | Data |
|---|---|
| Terlambat > 15 menit | Budi, 27 Okt (20 menit) |
| Terlambat ≤ 15 menit (tidak dipotong) | Andi, 29 Okt (10 menit); Budi, 22 Okt (5 menit) |
| Lembur 1,5 jam / 2 jam / 1 jam | Siti 28 Okt / Siti 21 Okt / Andi 29 Okt |
| Presensi ditolak | Budi, 29 Okt |
| Tidak hadir (ada jadwal, tanpa presensi) | Andi, 30 Okt |
| Upah berubah di tengah periode | Budi: 100.000 → 110.000 mulai 28 Okt |
| Pekerja bulanan masuk tengah bulan | Dewi, masuk 15 Okt |
| Kasbon aktif / sebagian / lunas | Budi 500.000 · Siti sisa 100.000 · Andi lunas |
| Periode jalur sukses | Periode 1 (19–24 Okt), semua hari tervalidasi |
| Periode harus diblokir | Periode 2 dan 3 (31 Okt belum divalidasi) |

Akun: `admin`, `budi`, `siti`, `andi`, `dewi` — password `password123`.

## Urutan kerja yang disarankan
1. Minggu 1 — masing-masing menyelesaikan CRUD/halaman dasar modulnya dengan data seed.
2. Minggu 2 — logika bisnis inti (validasi harian, mesin perhitungan, auth) + unit test.
3. Minggu 3 — integrasi lintas modul (presensi → validasi → penggajian → slip), perbaikan bug, UI.

## Definition of Done tiap fitur
- Sesuai use case & aturan bisnis (BR-xx) di dokumen requirement.
- Pembatasan akses (role, kepemilikan data, hari/periode ditutup) dicek **di server**.
- Nilai uang memakai `decimal.Decimal`; query memakai placeholder `?`.
- `gofmt`, `go vet`, `go test ./...` lolos (dicek CI).
- PR direview minimal satu anggota lain.
