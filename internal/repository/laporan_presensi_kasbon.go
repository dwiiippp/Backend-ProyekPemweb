package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/shopspring/decimal"
)

type LaporanPresensiKasbonRepository struct {
	db *sql.DB
}

func NewLaporanPresensiKasbonRepository(db *sql.DB) *LaporanPresensiKasbonRepository {
	return &LaporanPresensiKasbonRepository{db: db}
}

type PekerjaFilter struct {
	ID   int64
	Nama string
}

type BarisLaporanPresensi struct {
	Tanggal         time.Time
	PekerjaID       int64
	NamaPekerja     string
	NIK             string
	NamaShift       string
	JamMasukAktual  time.Time
	JamKeluarAktual time.Time
	MenitTerlambat  sql.NullInt64
	JamLembur       decimal.NullDecimal
	Status          string
	Catatan         sql.NullString
}

type RekapKasbon struct {
	PekerjaID   int64
	NamaPekerja string
	NIK         string
	Jumlah      int64
	Total       decimal.Decimal
	Sisa        decimal.Decimal
}

func (r *LaporanPresensiKasbonRepository) Pekerja(ctx context.Context) ([]PekerjaFilter, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT p.id, u.nama
		FROM pekerja p
		JOIN users u ON u.id = p.user_id AND u.deleted_at IS NULL
		WHERE p.deleted_at IS NULL
		ORDER BY u.nama`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []PekerjaFilter
	for rows.Next() {
		var item PekerjaFilter
		if err := rows.Scan(&item.ID, &item.Nama); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

func (r *LaporanPresensiKasbonRepository) Presensi(ctx context.Context, mulai, selesai time.Time, pekerjaID int64) ([]BarisLaporanPresensi, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT a.tanggal_absensi, p.id, u.nama, p.nik, s.nama_shift,
		       a.jam_masuk_aktual, a.jam_keluar_aktual, a.menit_terlambat,
		       a.jam_lembur, a.status_validasi, a.catatan_validasi
		FROM absensi a
		JOIN pekerja p ON p.id = a.pekerja_id AND p.deleted_at IS NULL
		JOIN users u ON u.id = p.user_id AND u.deleted_at IS NULL
		JOIN jadwal_kerja j ON j.id = a.jadwal_kerja_id
		JOIN shift s ON s.id = j.shift_id
		WHERE a.tanggal_absensi BETWEEN ? AND ?
		  AND (? = 0 OR p.id = ?)
		ORDER BY a.tanggal_absensi DESC, u.nama`,
		mulai.Format("2006-01-02"), selesai.Format("2006-01-02"), pekerjaID, pekerjaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []BarisLaporanPresensi
	for rows.Next() {
		var item BarisLaporanPresensi
		if err := rows.Scan(&item.Tanggal, &item.PekerjaID, &item.NamaPekerja, &item.NIK,
			&item.NamaShift, &item.JamMasukAktual, &item.JamKeluarAktual,
			&item.MenitTerlambat, &item.JamLembur, &item.Status, &item.Catatan); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

func (r *LaporanPresensiKasbonRepository) RekapKasbon(ctx context.Context) ([]RekapKasbon, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT p.id, u.nama, p.nik, COUNT(k.id),
		       COALESCE(SUM(k.nominal_kasbon), 0),
		       COALESCE(SUM(CASE WHEN k.status = 'aktif' THEN k.sisa_kasbon ELSE 0 END), 0)
		FROM pekerja p
		JOIN users u ON u.id = p.user_id AND u.deleted_at IS NULL
		LEFT JOIN kasbon k ON k.pekerja_id = p.id
		WHERE p.deleted_at IS NULL
		GROUP BY p.id, u.nama, p.nik
		HAVING COUNT(k.id) > 0
		ORDER BY u.nama`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []RekapKasbon
	for rows.Next() {
		var item RekapKasbon
		if err := rows.Scan(&item.PekerjaID, &item.NamaPekerja, &item.NIK, &item.Jumlah, &item.Total, &item.Sisa); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, rows.Err()
}
