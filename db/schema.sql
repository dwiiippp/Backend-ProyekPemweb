-- Schema Sistem Penggajian Konveksi Mitra Jaya (MySQL 8)
-- FILE BERSAMA (frozen): perubahan schema harus lewat PR tersendiri yang direview semua anggota.
-- Sumber: Detail Requirement bab 6 (ERD, Kamus Data, Catatan Implementasi).
-- Nilai uang memakai DECIMAL(12,2).

CREATE DATABASE IF NOT EXISTS penggajian_mitrajaya CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE penggajian_mitrajaya;

SET FOREIGN_KEY_CHECKS = 0;
DROP TABLE IF EXISTS koreksi_gaji, potongan_kasbon_gaji, gaji, periode_gaji, kasbon,
  validasi_presensi_harian, absensi, jadwal_kerja, shift, riwayat_gaji, pekerja, users;
SET FOREIGN_KEY_CHECKS = 1;

CREATE TABLE users (
  id            INT AUTO_INCREMENT PRIMARY KEY,
  username      VARCHAR(50)  NOT NULL UNIQUE,
  password_hash VARCHAR(255) NOT NULL,
  role          ENUM('admin','pekerja') NOT NULL,
  status        ENUM('aktif','nonaktif') NOT NULL DEFAULT 'aktif',
  dibuat_pada   TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  diubah_pada   TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB;

CREATE TABLE pekerja (
  id             INT AUTO_INCREMENT PRIMARY KEY,
  user_id        INT NULL UNIQUE,
  nik            VARCHAR(20)  NOT NULL UNIQUE,
  nama           VARCHAR(100) NOT NULL,
  kategori_kerja ENUM('harian','bulanan') NOT NULL,
  tanggal_masuk  DATE NOT NULL,
  tanggal_keluar DATE NULL,
  status         ENUM('aktif','nonaktif') NOT NULL DEFAULT 'aktif',
  dibuat_pada    TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  diubah_pada    TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  FOREIGN KEY (user_id) REFERENCES users(id)
) ENGINE=InnoDB;

CREATE TABLE riwayat_gaji (
  id              INT AUTO_INCREMENT PRIMARY KEY,
  pekerja_id      INT NOT NULL,
  nominal_gaji    DECIMAL(12,2) NOT NULL,
  tanggal_berlaku DATE NOT NULL,
  dibuat_pada     TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (pekerja_id) REFERENCES pekerja(id)
) ENGINE=InnoDB;

CREATE TABLE shift (
  id         INT AUTO_INCREMENT PRIMARY KEY,
  nama_shift VARCHAR(50) NOT NULL,
  jam_masuk  TIME NOT NULL,
  jam_keluar TIME NOT NULL
) ENGINE=InnoDB;

CREATE TABLE jadwal_kerja (
  id             INT AUTO_INCREMENT PRIMARY KEY,
  pekerja_id     INT NOT NULL,
  shift_id       INT NOT NULL,
  tanggal_jadwal DATE NOT NULL,
  dibuat_pada    TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE KEY uq_jadwal_pekerja_tanggal (pekerja_id, tanggal_jadwal),
  FOREIGN KEY (pekerja_id) REFERENCES pekerja(id),
  FOREIGN KEY (shift_id) REFERENCES shift(id)
) ENGINE=InnoDB;

CREATE TABLE absensi (
  id                INT AUTO_INCREMENT PRIMARY KEY,
  pekerja_id        INT NOT NULL,
  jadwal_kerja_id   INT NOT NULL UNIQUE,           -- satu absensi per jadwal (BR-13)
  tanggal_absensi   DATE NOT NULL,
  jam_masuk_aktual  DATETIME NULL,
  jam_keluar_aktual DATETIME NULL,
  menit_terlambat   INT NULL,                       -- diisi saat disetujui
  jam_lembur        DECIMAL(4,2) NULL,              -- diisi saat disetujui
  status_validasi   ENUM('menunggu','disetujui','ditolak') NOT NULL DEFAULT 'menunggu',
  catatan_validasi  TEXT NULL,
  divalidasi_oleh   INT NULL,
  divalidasi_pada   DATETIME NULL,
  dibuat_pada       TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  INDEX idx_absensi_tanggal (tanggal_absensi),
  FOREIGN KEY (pekerja_id) REFERENCES pekerja(id),
  FOREIGN KEY (jadwal_kerja_id) REFERENCES jadwal_kerja(id),
  FOREIGN KEY (divalidasi_oleh) REFERENCES users(id)
) ENGINE=InnoDB;

-- Ada baris untuk suatu tanggal  =>  presensi pada tanggal itu DITUTUP (BR-15).
CREATE TABLE validasi_presensi_harian (
  id               INT AUTO_INCREMENT PRIMARY KEY,
  tanggal_presensi DATE NOT NULL UNIQUE,
  divalidasi_oleh  INT NOT NULL,
  divalidasi_pada  DATETIME NOT NULL,
  dibuat_pada      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (divalidasi_oleh) REFERENCES users(id)
) ENGINE=InnoDB;

CREATE TABLE kasbon (
  id                INT AUTO_INCREMENT PRIMARY KEY,
  pekerja_id        INT NOT NULL,
  tanggal_pengajuan DATE NOT NULL,
  nominal_kasbon    DECIMAL(12,2) NOT NULL,
  sisa_kasbon       DECIMAL(12,2) NOT NULL,
  keterangan        TEXT NULL,
  status            ENUM('aktif','lunas') NOT NULL DEFAULT 'aktif',
  dibuat_pada       TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (pekerja_id) REFERENCES pekerja(id)
) ENGINE=InnoDB;

CREATE TABLE periode_gaji (
  id                 INT AUTO_INCREMENT PRIMARY KEY,
  jenis_periode      ENUM('mingguan','bulanan') NOT NULL,
  tanggal_mulai      DATE NOT NULL,
  tanggal_selesai    DATE NOT NULL,
  tanggal_pembayaran DATE NULL,
  status             ENUM('terbuka','ditutup') NOT NULL DEFAULT 'terbuka',
  dibuat_pada        TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB;

CREATE TABLE gaji (
  id                       INT AUTO_INCREMENT PRIMARY KEY,
  periode_gaji_id          INT NOT NULL,
  pekerja_id               INT NOT NULL,
  gaji_pokok               DECIMAL(12,2) NOT NULL DEFAULT 0,
  jumlah_hari_hadir        INT NOT NULL DEFAULT 0,
  potongan_terlambat       DECIMAL(12,2) NOT NULL DEFAULT 0,
  upah_lembur              DECIMAL(12,2) NOT NULL DEFAULT 0,
  penyesuaian_proporsional DECIMAL(12,2) NOT NULL DEFAULT 0,
  gaji_kotor               DECIMAL(12,2) NOT NULL DEFAULT 0,
  potongan_kasbon          DECIMAL(12,2) NOT NULL DEFAULT 0,
  gaji_bersih              DECIMAL(12,2) NOT NULL DEFAULT 0,
  dibuat_pada              TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE KEY uq_gaji_periode_pekerja (periode_gaji_id, pekerja_id),
  FOREIGN KEY (periode_gaji_id) REFERENCES periode_gaji(id),
  FOREIGN KEY (pekerja_id) REFERENCES pekerja(id)
) ENGINE=InnoDB;

CREATE TABLE potongan_kasbon_gaji (
  id               INT AUTO_INCREMENT PRIMARY KEY,
  gaji_id          INT NOT NULL,
  kasbon_id        INT NOT NULL,
  nominal_potongan DECIMAL(12,2) NOT NULL,
  dibuat_pada      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (gaji_id) REFERENCES gaji(id),
  FOREIGN KEY (kasbon_id) REFERENCES kasbon(id)
) ENGINE=InnoDB;

CREATE TABLE koreksi_gaji (
  id                    INT AUTO_INCREMENT PRIMARY KEY,
  periode_gaji_tujuan_id INT NOT NULL,
  gaji_sumber_id        INT NOT NULL,
  pekerja_id            INT NOT NULL,
  jenis_koreksi         ENUM('tambah','kurang') NOT NULL,
  nominal               DECIMAL(12,2) NOT NULL,
  alasan                TEXT NULL,
  dibuat_pada           TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (periode_gaji_tujuan_id) REFERENCES periode_gaji(id),
  FOREIGN KEY (gaji_sumber_id) REFERENCES gaji(id),
  FOREIGN KEY (pekerja_id) REFERENCES pekerja(id)
) ENGINE=InnoDB;
