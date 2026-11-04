// Package service berisi logika bisnis per modul.
// errors.go, password.go: FILE BERSAMA (frozen).
package service

import "errors"

// Error domain bersama — pakai ini agar handler semua modul menampilkan pesan yang seragam.
var (
	ErrTidakDitemukan    = errors.New("data tidak ditemukan")
	ErrInputTidakValid   = errors.New("input tidak valid")
	ErrBukanMilikSendiri = errors.New("data bukan milik Anda")
	ErrHariDitutup       = errors.New("presensi pada tanggal ini sudah divalidasi dan ditutup") // BR-15
	ErrPeriodeDitutup    = errors.New("periode gaji sudah ditutup, data tidak dapat diubah")    // BR-10
	ErrPresensiGanda     = errors.New("presensi untuk jadwal ini sudah ada")                    // BR-13
	ErrAkunNonaktif      = errors.New("akun tidak aktif")
	ErrKredensialSalah   = errors.New("username atau password salah")
)
