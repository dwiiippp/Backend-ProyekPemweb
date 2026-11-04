package model

import "time"

// Shift memetakan tabel `shift`. JamMasuk/JamKeluar berformat "HH:MM:SS".
type Shift struct {
	ID        int64
	NamaShift string
	JamMasuk  string
	JamKeluar string
}

// JadwalKerja memetakan tabel `jadwal_kerja` (UNIQUE pekerja_id + tanggal_jadwal).
type JadwalKerja struct {
	ID            int64
	PekerjaID     int64
	ShiftID       int64
	TanggalJadwal time.Time
	DibuatPada    time.Time
}
