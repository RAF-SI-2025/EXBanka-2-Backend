package repository

import (
	"context"
	"fmt"
	"time"

	"banka-backend/services/bank-service/internal/domain"

	"gorm.io/gorm"
)

// ─── GORM modeli ──────────────────────────────────────────────────────────────

// karticaModel je GORM projekcija tabele core_banking.kartica.
//
// HasOne relacija: jedna kartica ima najviše jedno ovlašćeno lice.
// GORM pronalazi child zapis koristeći strani ključ KarticaID na strani
// ovlascenoLiceModel (child je vlasnik FK kolone).
type karticaModel struct {
	ID                        int64               `gorm:"column:id;primaryKey"`
	BrojKartice               string              `gorm:"column:broj_kartice"`
	TipKartice                string              `gorm:"column:tip_kartice"`
	VrstaKartice              string              `gorm:"column:vrsta_kartice"`
	DatumKreiranja            time.Time           `gorm:"column:datum_kreiranja"`
	DatumIsteka               time.Time           `gorm:"column:datum_isteka"`
	RacunID                   int64               `gorm:"column:racun_id"`
	CvvKod                    string              `gorm:"column:cvv_kod"`
	LimitKartice              float64             `gorm:"column:limit_kartice"`
	Status                    string              `gorm:"column:status"`
	ProvizijaProcenat         *float64            `gorm:"column:provizija_procenat"`
	KonverzijaNaknadaProcenat *float64            `gorm:"column:konverziona_naknada_procenat"`
	OvlascenoLice             *ovlascenoLiceModel `gorm:"foreignKey:KarticaID"`
}

func (karticaModel) TableName() string { return "core_banking.kartica" }

func (m karticaModel) toDomain() domain.Kartica {
	k := domain.Kartica{
		ID:             m.ID,
		BrojKartice:    m.BrojKartice,
		TipKartice:     m.TipKartice,
		VrstaKartice:   m.VrstaKartice,
		DatumKreiranja: m.DatumKreiranja,
		DatumIsteka:    m.DatumIsteka,
		RacunID:        m.RacunID,
		LimitKartice:   m.LimitKartice,
		Status:         m.Status,
	}
	if m.OvlascenoLice != nil {
		ol := m.OvlascenoLice.toDomain()
		k.OvlascenoLice = &ol
	}
	return k
}

// ─────────────────────────────────────────────────────────────────────────────

// ovlascenoLiceModel je GORM projekcija tabele core_banking.ovlasceno_lice.
//
// BelongsTo relacija: ovlašćeno lice pripada tačno jednoj kartici.
// FK kolona KarticaID živi u ovoj tabeli (child strana 1:1 veze).
type ovlascenoLiceModel struct {
	ID            int64         `gorm:"column:id;primaryKey"`
	KarticaID     int64         `gorm:"column:kartica_id;not null"`
	Ime           string        `gorm:"column:ime"`
	Prezime       string        `gorm:"column:prezime"`
	Pol           string        `gorm:"column:pol"`
	EmailAdresa   string        `gorm:"column:email_adresa"`
	BrojTelefona  string        `gorm:"column:broj_telefona"`
	Adresa        string        `gorm:"column:adresa"`
	DatumRodjenja int64         `gorm:"column:datum_rodjenja"`
	Kartica       *karticaModel `gorm:"foreignKey:KarticaID"`
}

func (ovlascenoLiceModel) TableName() string { return "core_banking.ovlasceno_lice" }

func (m ovlascenoLiceModel) toDomain() domain.OvlascenoLice {
	return domain.OvlascenoLice{
		ID:            m.ID,
		KarticaID:     m.KarticaID,
		Ime:           m.Ime,
		Prezime:       m.Prezime,
		Pol:           m.Pol,
		EmailAdresa:   m.EmailAdresa,
		BrojTelefona:  m.BrojTelefona,
		Adresa:        m.Adresa,
		DatumRodjenja: m.DatumRodjenja,
	}
}

// ─── Repository implementacija ────────────────────────────────────────────────

type karticaRepository struct {
	db *gorm.DB
}

func NewKarticaRepository(db *gorm.DB) domain.KarticaRepository {
	return &karticaRepository{db: db}
}

// CreateKartica upisuje novu karticu u bazu i vraća njen ID.
// CvvKodHash iz inputa se direktno upisuje u cvv_kod kolonu (CHAR 64).
// DatumKreiranja mora biti eksplicitno postavljen u inputu — GORM ne postavlja
// default vrednost za TIMESTAMPTZ kolone sa DEFAULT NOW() ako polje nije nula.
func (r *karticaRepository) CreateKartica(ctx context.Context, input domain.CreateKarticaInput) (int64, error) {
	m := &karticaModel{
		RacunID:                   input.RacunID,
		BrojKartice:               input.BrojKartice,
		TipKartice:                input.TipKartice,
		VrstaKartice:              input.VrstaKartice,
		CvvKod:                    input.CvvKodHash,
		DatumKreiranja:            input.DatumKreiranja,
		DatumIsteka:               input.DatumIsteka,
		LimitKartice:              input.LimitKartice,
		Status:                    input.Status,
		ProvizijaProcenat:         input.ProvizijaProcenat,
		KonverzijaNaknadaProcenat: input.KonverzijaNaknadaProcenat,
	}
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return 0, err
	}
	return m.ID, nil
}

// CountKarticeZaRacun vraća ukupan broj kartica na datom računu.
// Koristi se za proveru limita na LICNI računima (max 2).
func (r *karticaRepository) CountKarticeZaRacun(ctx context.Context, racunID int64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&karticaModel{}).
		Where("racun_id = ?", racunID).
		Count(&count).Error
	return count, err
}

// HasVlasnikovaKarticaPostoji proverava da li vlasnik već ima karticu na računu.
//
// Vlasnikova kartica = kartica BEZ pridruženog ovlasceno_lice zapisa.
// Kartica zaposlenog (Flow 2) uvek ima ovlasceno_lice — LEFT JOIN + IS NULL
// razlikuje ova dva slučaja bez dodatne kolone u kartica tabeli.
func (r *karticaRepository) HasVlasnikovaKarticaPostoji(ctx context.Context, racunID int64) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Raw(`
		SELECT COUNT(*)
		FROM core_banking.kartica k
		LEFT JOIN core_banking.ovlasceno_lice ol ON ol.kartica_id = k.id
		WHERE k.racun_id = ?
		  AND ol.kartica_id IS NULL
	`, racunID).Scan(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// GetKarticaByID vraća karticu sa opcionalno učitanim ovlašćenim licem.
func (r *karticaRepository) GetKarticaByID(ctx context.Context, karticaID int64) (*domain.Kartica, error) {
	var m karticaModel
	err := r.db.WithContext(ctx).
		Preload("OvlascenoLice").
		First(&m, karticaID).Error
	if err != nil {
		return nil, err
	}
	k := m.toDomain()
	return &k, nil
}

// GetKarticeByRacun vraća sve kartice jednog računa sa učitanim ovlašćenim licima.
func (r *karticaRepository) GetKarticeByRacun(ctx context.Context, racunID int64) ([]domain.Kartica, error) {
	var models []karticaModel
	err := r.db.WithContext(ctx).
		Preload("OvlascenoLice").
		Where("racun_id = ?", racunID).
		Find(&models).Error
	if err != nil {
		return nil, err
	}
	kartice := make([]domain.Kartica, 0, len(models))
	for _, m := range models {
		kartice = append(kartice, m.toDomain())
	}
	return kartice, nil
}

// GetRacunInfo dohvata vrsta_racuna i mesecni_limit za dati račun.
// Koristi se pre kreiranja kartice da bi servis znao koji limit da proveri
// i koliki početni limit da dodeli kartici.
func (r *karticaRepository) GetRacunInfo(ctx context.Context, racunID int64) (*domain.RacunInfo, error) {
	var row struct {
		VrstaRacuna  string  `gorm:"column:vrsta_racuna"`
		MesecniLimit float64 `gorm:"column:mesecni_limit"`
		ValutaOznaka string  `gorm:"column:valuta_oznaka"`
	}
	err := r.db.WithContext(ctx).
		Table("core_banking.racun r").
		Joins("JOIN core_banking.valuta v ON v.id = r.id_valute").
		Select("r.vrsta_racuna, r.mesecni_limit, v.oznaka AS valuta_oznaka").
		Where("r.id = ?", racunID).
		Take(&row).Error
	if err != nil {
		return nil, err
	}
	return &domain.RacunInfo{
		VrstaRacuna:  row.VrstaRacuna,
		MesecniLimit: row.MesecniLimit,
		ValutaOznaka: row.ValutaOznaka,
	}, nil
}

// GetRacunVlasnikInfo dohvata vlasnik_id, vrsta_racuna, status i mesecni_limit za dati račun.
// Koristi se u Flow 2 za security proveru vlasništva i validaciju stanja računa.
func (r *karticaRepository) GetRacunVlasnikInfo(ctx context.Context, racunID int64) (*domain.RacunVlasnikInfo, error) {
	var row struct {
		IDVlasnika   int64   `gorm:"column:id_vlasnika"`
		VrstaRacuna  string  `gorm:"column:vrsta_racuna"`
		Status       string  `gorm:"column:status"`
		MesecniLimit float64 `gorm:"column:mesecni_limit"`
	}
	err := r.db.WithContext(ctx).
		Table("core_banking.racun").
		Select("id_vlasnika, vrsta_racuna, status, mesecni_limit").
		Where("id = ?", racunID).
		Take(&row).Error
	if err != nil {
		return nil, err
	}
	return &domain.RacunVlasnikInfo{
		VlasnikID:    row.IDVlasnika,
		VrstaRacuna:  row.VrstaRacuna,
		Status:       row.Status,
		MesecniLimit: row.MesecniLimit,
	}, nil
}

// CreateKarticaSaOvlascenoLicem kreira karticu i ovlašćeno lice atomično u jednoj
// PostgreSQL transakciji. Ako kreiranje ovlašćenog lica ne uspe, transakcija se
// poništava (rollback) i kartica ne ostaje u bazi u nedovršenom stanju.
func (r *karticaRepository) CreateKarticaSaOvlascenoLicem(
	ctx context.Context,
	karticaInput domain.CreateKarticaInput,
	olInput domain.OvlascenoLiceInput,
) (int64, error) {
	var karticaID int64
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		km := &karticaModel{
			RacunID:                   karticaInput.RacunID,
			BrojKartice:               karticaInput.BrojKartice,
			TipKartice:                karticaInput.TipKartice,
			VrstaKartice:              karticaInput.VrstaKartice,
			CvvKod:                    karticaInput.CvvKodHash,
			DatumKreiranja:            karticaInput.DatumKreiranja,
			DatumIsteka:               karticaInput.DatumIsteka,
			LimitKartice:              karticaInput.LimitKartice,
			Status:                    karticaInput.Status,
			ProvizijaProcenat:         karticaInput.ProvizijaProcenat,
			KonverzijaNaknadaProcenat: karticaInput.KonverzijaNaknadaProcenat,
		}
		if err := tx.Create(km).Error; err != nil {
			return fmt.Errorf("kreiranje kartice: %w", err)
		}
		karticaID = km.ID

		olm := &ovlascenoLiceModel{
			KarticaID:     karticaID,
			Ime:           olInput.Ime,
			Prezime:       olInput.Prezime,
			Pol:           olInput.Pol,
			EmailAdresa:   olInput.EmailAdresa,
			BrojTelefona:  olInput.BrojTelefona,
			Adresa:        olInput.Adresa,
			DatumRodjenja: olInput.DatumRodjenja,
		}
		if err := tx.Create(olm).Error; err != nil {
			return fmt.Errorf("kreiranje ovlašćenog lica: %w", err)
		}
		return nil
	})
	return karticaID, err
}

// GetKarticeKorisnika vraća sve kartice na računima čiji je vlasnik korisnikID,
// uključujući naziv i broj računa.
func (r *karticaRepository) GetKarticeKorisnika(ctx context.Context, korisnikID int64) ([]domain.KarticaSaRacunom, error) {
	type karticaRacunRow struct {
		ID             int64     `gorm:"column:id"`
		BrojKartice    string    `gorm:"column:broj_kartice"`
		TipKartice     string    `gorm:"column:tip_kartice"`
		VrstaKartice   string    `gorm:"column:vrsta_kartice"`
		DatumKreiranja time.Time `gorm:"column:datum_kreiranja"`
		DatumIsteka    time.Time `gorm:"column:datum_isteka"`
		RacunID        int64     `gorm:"column:racun_id"`
		LimitKartice   float64   `gorm:"column:limit_kartice"`
		Status         string    `gorm:"column:status"`
		NazivRacuna    string    `gorm:"column:naziv_racuna"`
		BrojRacuna     string    `gorm:"column:broj_racuna"`
	}
	var rows []karticaRacunRow
	err := r.db.WithContext(ctx).Raw(`
		SELECT k.id, k.broj_kartice, k.tip_kartice, k.vrsta_kartice,
		       k.datum_kreiranja, k.datum_isteka, k.racun_id,
		       k.limit_kartice, k.status,
		       ra.naziv_racuna, ra.broj_racuna
		FROM core_banking.kartica k
		JOIN core_banking.racun ra ON ra.id = k.racun_id
		WHERE ra.id_vlasnika = ?
		ORDER BY ra.id, k.id
	`, korisnikID).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	result := make([]domain.KarticaSaRacunom, 0, len(rows))
	for _, row := range rows {
		result = append(result, domain.KarticaSaRacunom{
			Kartica: domain.Kartica{
				ID:             row.ID,
				BrojKartice:    row.BrojKartice,
				TipKartice:     row.TipKartice,
				VrstaKartice:   row.VrstaKartice,
				DatumKreiranja: row.DatumKreiranja,
				DatumIsteka:    row.DatumIsteka,
				RacunID:        row.RacunID,
				LimitKartice:   row.LimitKartice,
				Status:         row.Status,
			},
			NazivRacuna: row.NazivRacuna,
			BrojRacuna:  row.BrojRacuna,
		})
	}
	return result, nil
}

// GetKarticaOwnerInfo vraća status kartice i ID vlasnika njenog računa.
// Vraća ErrKarticaNotFound ako kartica sa datim ID-om ne postoji.
func (r *karticaRepository) GetKarticaOwnerInfo(ctx context.Context, karticaID int64) (*domain.KarticaOwnerInfo, error) {
	var rows []struct {
		Status    string `gorm:"column:status"`
		VlasnikID int64  `gorm:"column:id_vlasnika"`
	}
	err := r.db.WithContext(ctx).Raw(`
		SELECT k.status, ra.id_vlasnika
		FROM core_banking.kartica k
		JOIN core_banking.racun ra ON ra.id = k.racun_id
		WHERE k.id = ?
	`, karticaID).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, domain.ErrKarticaNotFound
	}
	return &domain.KarticaOwnerInfo{
		Status:    rows[0].Status,
		VlasnikID: rows[0].VlasnikID,
	}, nil
}

// SetKarticaStatus ažurira status kartice po ID-u.
// Servis je odgovoran da pozove ovu metodu tek nakon validacije vlasništva i prelaza statusa.
func (r *karticaRepository) SetKarticaStatus(ctx context.Context, karticaID int64, noviStatus string) error {
	result := r.db.WithContext(ctx).
		Model(&karticaModel{}).
		Where("id = ?", karticaID).
		Update("status", noviStatus)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrKarticaNotFound
	}
	return nil
}

// HasOvlascenoLiceKarticu proverava limit za Flow 2:
// da li dato lice (po emailu) već ima karticu na bilo kom računu.
func (r *karticaRepository) HasOvlascenoLiceKarticu(ctx context.Context, emailAdresa string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&ovlascenoLiceModel{}).
		Where("email_adresa = ?", emailAdresa).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// karticaStatusChangeRow je GORM projekcija za dohvat kartice pri promeni statusa.
type karticaStatusChangeRow struct {
	ID             int64   `gorm:"column:id"`
	TrenutniStatus string  `gorm:"column:trenutni_status"`
	IDVlasnika     int64   `gorm:"column:id_vlasnika"`
	VrstaRacuna    string  `gorm:"column:vrsta_racuna"`
	OLIme          *string `gorm:"column:ol_ime"`
	OLPrezime      *string `gorm:"column:ol_prezime"`
	OLEmail        *string `gorm:"column:ol_email"`
	OLKarticaID    *int64  `gorm:"column:ol_kartica_id"`
}

// GetKarticaZaStatusChange dohvata karticu po broju kartice, zajedno sa podacima o računu
// i ovlašćenom licu (LEFT JOIN) — za portal zaposlenih.
func (r *karticaRepository) GetKarticaZaStatusChange(ctx context.Context, brojKartice string) (*domain.KarticaZaStatusChange, error) {
	var row karticaStatusChangeRow

	err := r.db.WithContext(ctx).Raw(`
		SELECT
			k.id,
			k.status              AS trenutni_status,
			ra.id_vlasnika,
			ra.vrsta_racuna,
			ol.ime                AS ol_ime,
			ol.prezime            AS ol_prezime,
			ol.email_adresa       AS ol_email,
			ol.kartica_id         AS ol_kartica_id
		FROM core_banking.kartica k
		JOIN core_banking.racun ra ON ra.id = k.racun_id
		LEFT JOIN core_banking.ovlasceno_lice ol ON ol.kartica_id = k.id
		WHERE k.broj_kartice = ?
		LIMIT 1
	`, brojKartice).Scan(&row).Error
	if err != nil {
		return nil, err
	}
	if row.ID == 0 {
		return nil, domain.ErrKarticaNotFound
	}

	result := &domain.KarticaZaStatusChange{
		ID:             row.ID,
		TrenutniStatus: row.TrenutniStatus,
		VlasnikID:      row.IDVlasnika,
		VrstaRacuna:    row.VrstaRacuna,
	}
	if row.OLKarticaID != nil {
		result.OvlascenoLice = &domain.OvlascenoLice{
			Ime:         ptrStr(row.OLIme),
			Prezime:     ptrStr(row.OLPrezime),
			EmailAdresa: ptrStr(row.OLEmail),
		}
	}
	return result, nil
}

// ptrStr dereferencira *string i vraća "" ako je nil.
func ptrStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// karticaEmployeeRow je GORM projekcija za upit na portalu zaposlenih.
type karticaEmployeeRow struct {
	ID          int64  `gorm:"column:id"`
	BrojKartice string `gorm:"column:broj_kartice"`
	Status      string `gorm:"column:status"`
	IDVlasnika  int64  `gorm:"column:id_vlasnika"`
}

// GetKarticeZaRacunBroj vraća sve kartice vezane za račun sa datim brojem računa,
// zajedno sa ID-em vlasnika — za portal zaposlenih.
func (r *karticaRepository) GetKarticeZaRacunBroj(ctx context.Context, brojRacuna string) ([]domain.KarticaEmployeeRow, error) {
	var rows []karticaEmployeeRow

	err := r.db.WithContext(ctx).Raw(`
		SELECT
			k.id,
			k.broj_kartice,
			k.status,
			ra.id_vlasnika
		FROM core_banking.kartica k
		JOIN core_banking.racun ra ON ra.id = k.racun_id
		WHERE ra.broj_racuna = ?
	`, brojRacuna).Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	result := make([]domain.KarticaEmployeeRow, 0, len(rows))
	for _, row := range rows {
		result = append(result, domain.KarticaEmployeeRow{
			ID:          row.ID,
			BrojKartice: row.BrojKartice,
			Status:      row.Status,
			VlasnikID:   row.IDVlasnika,
		})
	}
	return result, nil
}
