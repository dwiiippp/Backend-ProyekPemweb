package repository

import (
	"context"
	"database/sql"
)

// Tabel: pekerja (baca)
// Semua query SQL modul ini ditulis di sini (pakai placeholder ?, jangan concat string).
type HomeRepository struct {
	db *sql.DB
}

func NewHomeRepository(db *sql.DB) *HomeRepository {
	return &HomeRepository{db: db}
}

type HomeAdminSummary struct {
	PresensiMenunggu int
	PeriodeTerbuka   int
}

type HomeTodaySchedule struct {
	NamaShift string
	JamMasuk  string
	JamKeluar string
}

type HomeProfile struct {
	NIK           string
	Nama          string
	KategoriKerja string
}

// Mengambil ringkasan untuk dashboard
func (r *HomeRepository) GetAdminSummary(ctx context.Context) (*HomeAdminSummary, error) {
	var data HomeAdminSummary

	query := `
		SELECT (SELECT COUNT(*) FROM absensi WHERE status_validasi = 'menunggu'), (SELECT COUNT(*) FROM periode_gaji WHERE status = 'terbuka')
	`
	err := r.db.QueryRowContext(ctx, query).Scan(
		&data.PresensiMenunggu,
		&data.PeriodeTerbuka,
	)

	if err != nil {
		return nil, err
	}

	return &data, nil
}

// Mengambil jadwal pekerja berdasarkan id pekerja
func (r *HomeRepository) GetTodaySchedule(ctx context.Context, pekerjaID int64) (*HomeTodaySchedule, error) {
	var data HomeTodaySchedule

	query := `
		SELECT
			s.nama_shift, TIME_FORMAT(s.jam_masuk, '%H:%i'), TIME_FORMAT(s.jam_keluar, '%H:%i')
		FROM jadwal_kerja j JOIN shift s ON s.id = j.shift_id
		WHERE j.pekerja_id = ? AND j.tanggal_jadwal = CURDATE() AND s.deleted_at IS NULL
		LIMIT 1
	`

	err := r.db.QueryRowContext(ctx, query, pekerjaID).Scan(
		&data.NamaShift,
		&data.JamMasuk,
		&data.JamKeluar,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &data, nil
}

// Mengambil profile pekerja berdasarkan id pekerja
func (r *HomeRepository) GetProfile(ctx context.Context, pekerjaID int64) (*HomeProfile, error) {
	var data HomeProfile

	query := `
		SELECT p.nik, u.nama, p.kategori_kerja FROM pekerja p
		JOIN users u ON u.id = p.user_id WHERE p.id = ? AND p.deleted_at IS NULL AND u.deleted_at IS NULL
		LIMIT 1
	`
	err := r.db.QueryRowContext(ctx, query, pekerjaID).Scan(
		&data.NIK,
		&data.Nama,
		&data.KategoriKerja,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &data, nil
}
