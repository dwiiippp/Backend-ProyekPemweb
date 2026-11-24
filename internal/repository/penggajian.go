package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"penggajian/internal/model"

	"github.com/shopspring/decimal"
)

type queryRunner interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

// KasbonAktifPenggajian menampung data kasbon aktif untuk proses potongan gaji.
type KasbonAktifPenggajian struct {
	ID               int64
	PekerjaID        int64
	TanggalPengajuan time.Time
	Nominal          decimal.Decimal
	Sisa             decimal.Decimal
}

// PresensiMenungguRingkas menampung ringkasan presensi yang masih menunggu validasi per tanggal.
type PresensiMenungguRingkas struct {
	Tanggal time.Time
	Jumlah  int
}

// HasilSimpanGaji menampung nilai gaji yang akan disimpan ke database.
type HasilSimpanGaji struct {
	GajiPokok               decimal.Decimal
	JumlahHariHadir         int
	PotonganTerlambat       decimal.Decimal
	UpahLembur              decimal.Decimal
	PenyesuaianProporsional decimal.Decimal
	GajiKotor               decimal.Decimal
	PotonganKasbon          decimal.Decimal
	GajiBersih              decimal.Decimal
}

// BarisHasilGaji menampung baris gaji tersimpan beserta identitas pekerja untuk tampilan hasil.
type BarisHasilGaji struct {
	GajiID                  int64
	PekerjaID               int64
	NIK                     string
	NamaPekerja             string
	KategoriKerja           string
	JumlahHariHadir         int
	GajiPokok               decimal.Decimal
	PotonganTerlambat       decimal.Decimal
	UpahLembur              decimal.Decimal
	PenyesuaianProporsional decimal.Decimal
	TotalKoreksi            decimal.Decimal
	GajiKotor               decimal.Decimal
	PotonganKasbon          decimal.Decimal
	GajiBersih              decimal.Decimal
}

// Pemilik: Anggota 3 | F-07 · F-09 · UC-12 Proses Penggajian · UC-13 Tutup Periode
type PenggajianRepository struct {
	db *sql.DB
}

func NewPenggajianRepository(db *sql.DB) *PenggajianRepository {
	return &PenggajianRepository{db: db}
}

// BeginTx membuka transaksi database untuk proses penggajian.
func (r *PenggajianRepository) BeginTx(ctx context.Context) (*sql.Tx, error) {
	return r.db.BeginTx(ctx, nil)
}

func (r *PenggajianRepository) DaftarPeriode(ctx context.Context) ([]model.PeriodeGaji, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, jenis_periode, tanggal_mulai, tanggal_selesai,
		       tanggal_pembayaran, status, created_at
		FROM periode_gaji
		ORDER BY status ASC, tanggal_mulai DESC, id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var hasil []model.PeriodeGaji
	for rows.Next() {
		var p model.PeriodeGaji
		var tglBayar sql.NullTime
		if err := rows.Scan(
			&p.ID, &p.JenisPeriode, &p.TanggalMulai, &p.TanggalSelesai,
			&tglBayar, &p.Status, &p.DibuatPada,
		); err != nil {
			return nil, err
		}
		if tglBayar.Valid {
			t := tglBayar.Time
			p.TanggalPembayaran = &t
		}
		hasil = append(hasil, p)
	}
	return hasil, rows.Err()
}

func (r *PenggajianRepository) AmbilPeriodeByID(ctx context.Context, id int64) (model.PeriodeGaji, error) {
	return ambilPeriode(ctx, r.db, id, false)
}

func (r *PenggajianRepository) AmbilPeriodeByIDForUpdateTx(ctx context.Context, tx *sql.Tx, id int64) (model.PeriodeGaji, error) {
	return ambilPeriode(ctx, tx, id, true)
}

func ambilPeriode(ctx context.Context, q queryRunner, id int64, forUpdate bool) (model.PeriodeGaji, error) {
	query := `
		SELECT id, jenis_periode, tanggal_mulai, tanggal_selesai,
		       tanggal_pembayaran, status, created_at
		FROM periode_gaji
		WHERE id = ?`
	if forUpdate {
		query += " FOR UPDATE"
	}

	var p model.PeriodeGaji
	var tglBayar sql.NullTime
	err := q.QueryRowContext(ctx, query, id).Scan(
		&p.ID, &p.JenisPeriode, &p.TanggalMulai, &p.TanggalSelesai,
		&tglBayar, &p.Status, &p.DibuatPada,
	)
	if err != nil {
		return model.PeriodeGaji{}, err
	}
	if tglBayar.Valid {
		t := tglBayar.Time
		p.TanggalPembayaran = &t
	}
	return p, nil
}

func (r *PenggajianRepository) TanggalBelumDivalidasi(ctx context.Context, mulai, selesai time.Time) ([]time.Time, error) {
	return tanggalBelumDivalidasi(ctx, r.db, mulai, selesai)
}

func (r *PenggajianRepository) TanggalBelumDivalidasiTx(ctx context.Context, tx *sql.Tx, mulai, selesai time.Time) ([]time.Time, error) {
	return tanggalBelumDivalidasi(ctx, tx, mulai, selesai)
}

func tanggalBelumDivalidasi(ctx context.Context, q queryRunner, mulai, selesai time.Time) ([]time.Time, error) {
	rows, err := q.QueryContext(ctx, `
		SELECT DISTINCT j.tanggal_jadwal
		FROM jadwal_kerja j
		LEFT JOIN validasi_presensi_harian v ON v.tanggal_presensi = j.tanggal_jadwal
		WHERE j.tanggal_jadwal BETWEEN ? AND ?
		  AND v.id IS NULL
		ORDER BY j.tanggal_jadwal ASC`,
		mulai.Format("2006-01-02"),
		selesai.Format("2006-01-02"),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var hasil []time.Time
	for rows.Next() {
		var t time.Time
		if err := rows.Scan(&t); err != nil {
			return nil, err
		}
		hasil = append(hasil, t)
	}
	return hasil, rows.Err()
}

func (r *PenggajianRepository) PresensiMenunggu(ctx context.Context, mulai, selesai time.Time) ([]PresensiMenungguRingkas, error) {
	return presensiMenunggu(ctx, r.db, mulai, selesai)
}

func (r *PenggajianRepository) PresensiMenungguTx(ctx context.Context, tx *sql.Tx, mulai, selesai time.Time) ([]PresensiMenungguRingkas, error) {
	return presensiMenunggu(ctx, tx, mulai, selesai)
}

func presensiMenunggu(ctx context.Context, q queryRunner, mulai, selesai time.Time) ([]PresensiMenungguRingkas, error) {
	rows, err := q.QueryContext(ctx, `
		SELECT tanggal_absensi, COUNT(*) AS jumlah
		FROM absensi
		WHERE tanggal_absensi BETWEEN ? AND ?
		  AND status_validasi = 'menunggu'
		GROUP BY tanggal_absensi
		ORDER BY tanggal_absensi ASC`,
		mulai.Format("2006-01-02"),
		selesai.Format("2006-01-02"),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var hasil []PresensiMenungguRingkas
	for rows.Next() {
		var item PresensiMenungguRingkas
		if err := rows.Scan(&item.Tanggal, &item.Jumlah); err != nil {
			return nil, err
		}
		hasil = append(hasil, item)
	}
	return hasil, rows.Err()
}

// ListPekerjaEligible mengambil pekerja aktif yang sesuai kategori dan aktif dalam rentang periode.
func (r *PenggajianRepository) ListPekerjaEligible(ctx context.Context, kategori string, mulai, selesai time.Time) ([]model.Pekerja, error) {
	return listPekerjaEligible(ctx, r.db, kategori, mulai, selesai, false)
}

func (r *PenggajianRepository) ListPekerjaEligibleTx(ctx context.Context, tx *sql.Tx, kategori string, mulai, selesai time.Time) ([]model.Pekerja, error) {
	return listPekerjaEligible(ctx, tx, kategori, mulai, selesai, true)
}

func listPekerjaEligible(ctx context.Context, q queryRunner, kategori string, mulai, selesai time.Time, forUpdate bool) ([]model.Pekerja, error) {
	query := `
    SELECT
			p.id, p.user_id, p.nik, u.nama, p.kategori_kerja,
			p.tanggal_masuk, p.tanggal_keluar, p.status,
			p.created_at, p.updated_at
		FROM pekerja p
		JOIN users u ON u.id = p.user_id
		WHERE p.deleted_at IS NULL
			AND p.status = 'aktif'
			AND u.status = 'aktif'
			AND u.deleted_at IS NULL
			AND p.kategori_kerja = ?
			AND p.tanggal_masuk <= ?
			AND (p.tanggal_keluar IS NULL OR p.tanggal_keluar >= ?)
		ORDER BY p.id ASC
	`

	// TAMBAHKAN FOR UPDATE jika query berjalan dalam transaksi.
	if forUpdate {
		query += " FOR UPDATE"
	}

	rows, err := q.QueryContext(
		ctx,
		query,
		kategori,
		selesai.Format("2006-01-02"),
		mulai.Format("2006-01-02"),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []model.Pekerja
	for rows.Next() {
		var item model.Pekerja
		var userID sql.NullInt64
		var tglKeluar, updated sql.NullTime
		if err := rows.Scan(
			&item.ID, &userID, &item.NIK, &item.Nama, &item.KategoriKerja,
			&item.TanggalMasuk, &tglKeluar, &item.Status,
			&item.DibuatPada, &updated,
		); err != nil {
			return nil, err
		}
		if userID.Valid {
			v := userID.Int64
			item.UserID = &v
		}
		if tglKeluar.Valid {
			t := tglKeluar.Time
			item.TanggalKeluar = &t
		}
		if updated.Valid {
			item.DiubahPada = updated.Time
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

// ListAbsensiDisetujui mengambil absensi yang sudah disetujui milik pekerja dalam rentang periode.
func (r *PenggajianRepository) ListAbsensiDisetujui(
	ctx context.Context,
	pekerjaID int64,
	tanggalMulai time.Time,
	tanggalSelesai time.Time,
) ([]model.Absensi, error) {
	return listAbsensiDisetujui(ctx, r.db, pekerjaID, tanggalMulai, tanggalSelesai)
}

// ListAbsensiDisetujuiTx mengambil absensi yang disetujui menggunakan transaksi database.
func (r *PenggajianRepository) ListAbsensiDisetujuiTx(
	ctx context.Context,
	tx *sql.Tx,
	pekerjaID int64,
	tanggalMulai time.Time,
	tanggalSelesai time.Time,
) ([]model.Absensi, error) {
	return listAbsensiDisetujui(ctx, tx, pekerjaID, tanggalMulai, tanggalSelesai)
}

func listAbsensiDisetujui(
	ctx context.Context,
	q queryRunner,
	pekerjaID int64,
	tanggalMulai time.Time,
	tanggalSelesai time.Time,
) ([]model.Absensi, error) {
	rows, err := q.QueryContext(ctx, `
		SELECT
			a.id, a.pekerja_id, a.jadwal_kerja_id, a.tanggal_absensi,
			a.jam_masuk_aktual, a.jam_keluar_aktual, a.menit_terlambat,
			a.jam_lembur, a.status_validasi, a.catatan_validasi,
			a.divalidasi_oleh, a.divalidasi_pada, a.created_at
		FROM absensi a
		WHERE a.pekerja_id = ?
		  AND a.tanggal_absensi BETWEEN ? AND ?
		  AND a.status_validasi = 'disetujui'
		ORDER BY a.tanggal_absensi ASC`,
		pekerjaID,
		tanggalMulai.Format("2006-01-02"),
		tanggalSelesai.Format("2006-01-02"),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []model.Absensi
	for rows.Next() {
		var item model.Absensi
		if err := rows.Scan(
			&item.ID, &item.PekerjaID, &item.JadwalKerjaID, &item.TanggalAbsensi,
			&item.JamMasukAktual, &item.JamKeluarAktual, &item.MenitTerlambat,
			&item.JamLembur, &item.StatusValidasi, &item.CatatanValidasi,
			&item.DivalidasiOleh, &item.DivalidasiPada, &item.DibuatPada,
		); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

func (r *PenggajianRepository) ListRiwayatGajiPekerja(ctx context.Context, pekerjaID int64, sampai time.Time) ([]model.RiwayatGaji, error) {
	return listRiwayatGajiPekerja(ctx, r.db, pekerjaID, sampai)
}

func (r *PenggajianRepository) ListRiwayatGajiPekerjaTx(ctx context.Context, tx *sql.Tx, pekerjaID int64, sampai time.Time) ([]model.RiwayatGaji, error) {
	return listRiwayatGajiPekerja(ctx, tx, pekerjaID, sampai)
}

func listRiwayatGajiPekerja(ctx context.Context, q queryRunner, pekerjaID int64, sampai time.Time) ([]model.RiwayatGaji, error) {
	rows, err := q.QueryContext(ctx, `
		SELECT id, pekerja_id, nominal_gaji, tanggal_berlaku, created_at
		FROM riwayat_gaji
		WHERE pekerja_id = ?
		  AND tanggal_berlaku <= ?
		ORDER BY tanggal_berlaku ASC, id ASC`,
		pekerjaID,
		sampai.Format("2006-01-02"),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []model.RiwayatGaji
	for rows.Next() {
		var item model.RiwayatGaji
		if err := rows.Scan(
			&item.ID, &item.PekerjaID, &item.NominalGaji, &item.TanggalBerlaku, &item.DibuatPada,
		); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

func (r *PenggajianRepository) TotalKoreksiPeriode(ctx context.Context, periodeTujuanID, pekerjaID int64) (decimal.Decimal, error) {
	return totalKoreksiPeriode(ctx, r.db, periodeTujuanID, pekerjaID)
}

func (r *PenggajianRepository) TotalKoreksiPeriodeTx(ctx context.Context, tx *sql.Tx, periodeTujuanID, pekerjaID int64) (decimal.Decimal, error) {
	return totalKoreksiPeriode(ctx, tx, periodeTujuanID, pekerjaID)
}

func totalKoreksiPeriode(ctx context.Context, q queryRunner, periodeTujuanID, pekerjaID int64) (decimal.Decimal, error) {
	var total decimal.Decimal
	err := q.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(CASE WHEN jenis_koreksi = 'tambah' THEN nominal ELSE -nominal END), 0)
		FROM koreksi_gaji
		WHERE periode_gaji_tujuan_id = ?
		  AND pekerja_id = ?`,
		periodeTujuanID,
		pekerjaID,
	).Scan(&total)
	return total, err
}

// ListKasbonAktif mengambil kasbon aktif yang masih memiliki sisa (FIFO).
func (r *PenggajianRepository) ListKasbonAktif(ctx context.Context, pekerjaID int64) ([]KasbonAktifPenggajian, error) {
	return listKasbonAktif(ctx, r.db, pekerjaID, false)
}

// ListKasbonAktifForUpdateTx mengambil dan mengunci kasbon aktif (FIFO) dalam transaksi.
func (r *PenggajianRepository) ListKasbonAktifForUpdateTx(ctx context.Context, tx *sql.Tx, pekerjaID int64) ([]KasbonAktifPenggajian, error) {
	return listKasbonAktif(ctx, tx, pekerjaID, true)
}

func listKasbonAktif(ctx context.Context, q queryRunner, pekerjaID int64, forUpdate bool) ([]KasbonAktifPenggajian, error) {
	query := `
		SELECT id, pekerja_id, tanggal_pengajuan, nominal_kasbon, sisa_kasbon
		FROM kasbon
		WHERE pekerja_id = ?
		  AND status = 'aktif'
		  AND sisa_kasbon > 0
		ORDER BY tanggal_pengajuan ASC, id ASC`
	if forUpdate {
		query += " FOR UPDATE"
	}

	rows, err := q.QueryContext(ctx, query, pekerjaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []KasbonAktifPenggajian
	for rows.Next() {
		var item KasbonAktifPenggajian
		if err := rows.Scan(
			&item.ID, &item.PekerjaID, &item.TanggalPengajuan, &item.Nominal, &item.Sisa,
		); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

// SimpanGaji menyimpan hasil kalkulasi gaji seorang pekerja ke dalam tabel gaji.
func (r *PenggajianRepository) SimpanGaji(
	ctx context.Context,
	tx *sql.Tx,
	periodeGajiID int64,
	pekerjaID int64,
	hasil HasilSimpanGaji,
) (int64, error) {
	result, err := tx.ExecContext(ctx, `
		INSERT INTO gaji (
			periode_gaji_id, pekerja_id, gaji_pokok, jumlah_hari_hadir,
			potongan_terlambat, upah_lembur, penyesuaian_proporsional,
			gaji_kotor, potongan_kasbon, gaji_bersih
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		periodeGajiID, pekerjaID, hasil.GajiPokok, hasil.JumlahHariHadir,
		hasil.PotonganTerlambat, hasil.UpahLembur, hasil.PenyesuaianProporsional,
		hasil.GajiKotor, hasil.PotonganKasbon, hasil.GajiBersih,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// CatatPotonganKasbon mencatat potongan kasbon ke gaji dan memperbarui sisa kasbon dalam transaksi.
func (r *PenggajianRepository) CatatPotonganKasbon(
	ctx context.Context,
	tx *sql.Tx,
	gajiID int64,
	kasbonID int64,
	nominalPotongan decimal.Decimal,
) error {
	if nominalPotongan.LessThanOrEqual(decimal.Zero) {
		return fmt.Errorf("nominal potongan kasbon harus lebih dari 0")
	}

	result, err := tx.ExecContext(ctx, `
		UPDATE kasbon
		SET
			status = CASE
				WHEN sisa_kasbon <= ? THEN 'lunas'
				ELSE 'aktif'
			END,
			sisa_kasbon = sisa_kasbon - ?
		WHERE id = ?
		AND sisa_kasbon >= ?`,
		nominalPotongan,
		nominalPotongan,
		kasbonID,
		nominalPotongan,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected != 1 {
		return fmt.Errorf("kasbon tidak ditemukan atau saldo tidak mencukupi")
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO potongan_kasbon_gaji (gaji_id, kasbon_id, nominal_potongan)
		VALUES (?, ?, ?)`,
		gajiID, kasbonID, nominalPotongan,
	)
	return err
}

func (r *PenggajianRepository) TutupPeriodeTx(ctx context.Context, tx *sql.Tx, periodeID int64, tanggalPembayaran time.Time) error {
	res, err := tx.ExecContext(ctx, `
		UPDATE periode_gaji
		SET status = ?, tanggal_pembayaran = ?
		WHERE id = ? AND status = ?`,
		model.PeriodeDitutup,
		tanggalPembayaran.Format("2006-01-02"),
		periodeID,
		model.PeriodeTerbuka,
	)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected != 1 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *PenggajianRepository) DaftarHasilGajiPeriode(ctx context.Context, periodeID int64) ([]BarisHasilGaji, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT
			g.id, g.pekerja_id, p.nik, u.nama, p.kategori_kerja,
			g.jumlah_hari_hadir, g.gaji_pokok, g.potongan_terlambat,
			g.upah_lembur, g.penyesuaian_proporsional,
			COALESCE((
				SELECT SUM(CASE WHEN k.jenis_koreksi = 'tambah' THEN k.nominal ELSE -k.nominal END)
				FROM koreksi_gaji k
				WHERE k.periode_gaji_tujuan_id = g.periode_gaji_id
				  AND k.pekerja_id = g.pekerja_id
			), 0) AS total_koreksi,
			g.gaji_kotor, g.potongan_kasbon, g.gaji_bersih
		FROM gaji g
		JOIN pekerja p ON p.id = g.pekerja_id
		JOIN users u ON u.id = p.user_id
		WHERE g.periode_gaji_id = ?
		ORDER BY u.nama ASC, p.id ASC`,
		periodeID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var hasil []BarisHasilGaji
	for rows.Next() {
		var item BarisHasilGaji
		if err := rows.Scan(
			&item.GajiID, &item.PekerjaID, &item.NIK, &item.NamaPekerja, &item.KategoriKerja,
			&item.JumlahHariHadir, &item.GajiPokok, &item.PotonganTerlambat,
			&item.UpahLembur, &item.PenyesuaianProporsional, &item.TotalKoreksi,
			&item.GajiKotor, &item.PotonganKasbon, &item.GajiBersih,
		); err != nil {
			return nil, err
		}
		hasil = append(hasil, item)
	}
	return hasil, rows.Err()
}
