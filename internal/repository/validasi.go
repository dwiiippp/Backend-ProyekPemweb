package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

type ValidasiRepository struct {
	db *sql.DB
}

func NewValidasiRepository(db *sql.DB) *ValidasiRepository {
	return &ValidasiRepository{db: db}
}

type HariValidasi struct {
	Tanggal  time.Time
	Jumlah   int
	Menunggu int
	Ditutup  bool
}

type PresensiValidasi struct {
	ID              int64
	NamaPekerja     string
	NIK             string
	NamaShift       string
	JamMasukShift   string
	JamKeluarShift  string
	JamMasukAktual  time.Time
	JamKeluarAktual time.Time
	Status          string
	Catatan         sql.NullString
	HariDitutup     bool
}

type KeputusanValidasi struct {
	AbsensiID      int64
	Status         string
	Catatan        string
	MenitTerlambat sql.NullInt64
	JamLembur      sql.NullString
}

var (
	ErrTanggalValidasiTidakDitemukan = errors.New("tanggal validasi tidak ditemukan")
	ErrValidasiSudahDitutup          = errors.New("tanggal sudah divalidasi")
	ErrValidasiPeriodeDitutup        = errors.New("periode gaji sudah ditutup")
	ErrDataValidasiBerubah           = errors.New("data presensi berubah, muat ulang halaman")
)

func (r *ValidasiRepository) DaftarHari(ctx context.Context) ([]HariValidasi, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT j.tanggal_jadwal, COUNT(a.id) AS jumlah,
		       COALESCE(SUM(a.status_validasi = 'menunggu'), 0) AS menunggu,
		       MAX(v.tanggal_presensi IS NOT NULL) AS ditutup
		FROM jadwal_kerja j
		LEFT JOIN absensi a ON a.jadwal_kerja_id = j.id AND a.tanggal_absensi = j.tanggal_jadwal
		LEFT JOIN validasi_presensi_harian v ON v.tanggal_presensi = j.tanggal_jadwal
		WHERE j.tanggal_jadwal <= CURDATE()
		GROUP BY j.tanggal_jadwal
		ORDER BY j.tanggal_jadwal DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []HariValidasi
	for rows.Next() {
		var item HariValidasi
		if err := rows.Scan(&item.Tanggal, &item.Jumlah, &item.Menunggu, &item.Ditutup); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

func (r *ValidasiRepository) InfoTanggal(ctx context.Context, tanggal time.Time) (bool, bool, error) {
	var hasSchedule, closed bool
	err := r.db.QueryRowContext(ctx, `
		SELECT EXISTS(SELECT 1 FROM jadwal_kerja WHERE tanggal_jadwal = ?),
		       EXISTS(SELECT 1 FROM validasi_presensi_harian WHERE tanggal_presensi = ?)`,
		tanggal.Format("2006-01-02"), tanggal.Format("2006-01-02")).Scan(&hasSchedule, &closed)
	return hasSchedule, closed, err
}

func (r *ValidasiRepository) Detail(ctx context.Context, tanggal time.Time) ([]PresensiValidasi, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT a.id, u.nama, p.nik, s.nama_shift, s.jam_masuk, s.jam_keluar,
		       a.jam_masuk_aktual, a.jam_keluar_aktual, a.status_validasi, a.catatan_validasi,
		       EXISTS(SELECT 1 FROM validasi_presensi_harian v WHERE v.tanggal_presensi = a.tanggal_absensi)
		FROM absensi a
		JOIN pekerja p ON p.id = a.pekerja_id
		JOIN users u ON u.id = p.user_id
		JOIN jadwal_kerja j ON j.id = a.jadwal_kerja_id
		JOIN shift s ON s.id = j.shift_id
		WHERE a.tanggal_absensi = ?
		ORDER BY u.nama, a.id`, tanggal.Format("2006-01-02"))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []PresensiValidasi
	for rows.Next() {
		var item PresensiValidasi
		if err := rows.Scan(&item.ID, &item.NamaPekerja, &item.NIK, &item.NamaShift,
			&item.JamMasukShift, &item.JamKeluarShift, &item.JamMasukAktual,
			&item.JamKeluarAktual, &item.Status, &item.Catatan, &item.HariDitutup); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

func (r *ValidasiRepository) Simpan(ctx context.Context, tanggal time.Time, validatorID int64, decisions []KeputusanValidasi) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	rows, err := tx.QueryContext(ctx, `
		SELECT id FROM jadwal_kerja WHERE tanggal_jadwal = ? ORDER BY id FOR UPDATE`,
		tanggal.Format("2006-01-02"))
	if err != nil {
		return err
	}
	var scheduleIDs []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return err
		}
		scheduleIDs = append(scheduleIDs, id)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return err
	}
	if err := rows.Close(); err != nil {
		return err
	}
	if len(scheduleIDs) == 0 {
		return ErrTanggalValidasiTidakDitemukan
	}

	var closed bool
	err = tx.QueryRowContext(ctx, `
		SELECT EXISTS(SELECT 1 FROM validasi_presensi_harian WHERE tanggal_presensi = ?)`,
		tanggal.Format("2006-01-02")).Scan(&closed)
	if err != nil {
		return err
	}
	if closed {
		return ErrValidasiSudahDitutup
	}
	closed, err = adaPeriodeDitutup(ctx, tx, tanggal.Format("2006-01-02"))
	if err != nil {
		return err
	}
	if closed {
		return ErrValidasiPeriodeDitutup
	}

	pendingRows, err := tx.QueryContext(ctx, `
		SELECT id FROM absensi WHERE tanggal_absensi = ? AND status_validasi = 'menunggu' FOR UPDATE`,
		tanggal.Format("2006-01-02"))
	if err != nil {
		return err
	}
	pendingIDs := make(map[int64]bool)
	for pendingRows.Next() {
		var id int64
		if err := pendingRows.Scan(&id); err != nil {
			pendingRows.Close()
			return err
		}
		pendingIDs[id] = true
	}
	if err := pendingRows.Err(); err != nil {
		pendingRows.Close()
		return err
	}
	if err := pendingRows.Close(); err != nil {
		return err
	}
	if len(pendingIDs) != len(decisions) {
		return ErrDataValidasiBerubah
	}

	for _, decision := range decisions {
		if !pendingIDs[decision.AbsensiID] {
			return ErrDataValidasiBerubah
		}
		var late, overtime any
		if decision.Status == "disetujui" {
			late = decision.MenitTerlambat.Int64
			if decision.JamLembur.Valid {
				overtime = decision.JamLembur.String
			}
		}
		var note any
		if decision.Status == "ditolak" {
			note = decision.Catatan
		}
		result, err := tx.ExecContext(ctx, `
			UPDATE absensi
			SET status_validasi = ?, catatan_validasi = ?, menit_terlambat = ?, jam_lembur = ?,
			    divalidasi_oleh = ?, divalidasi_pada = NOW()
			WHERE id = ? AND tanggal_absensi = ? AND status_validasi = 'menunggu'`,
			decision.Status, note, late, overtime, validatorID, decision.AbsensiID, tanggal.Format("2006-01-02"))
		if err != nil {
			return err
		}
		affected, err := result.RowsAffected()
		if err != nil {
			return err
		}
		if affected != 1 {
			return ErrDataValidasiBerubah
		}
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO validasi_presensi_harian (tanggal_presensi, divalidasi_oleh, divalidasi_pada)
		VALUES (?, ?, NOW())`, tanggal.Format("2006-01-02"), validatorID); err != nil {
		return err
	}
	return tx.Commit()
}
