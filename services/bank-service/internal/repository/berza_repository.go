package repository

import (
	"context"
	"errors"
	"strings"
	"time"

	"banka-backend/services/bank-service/internal/domain"

	"gorm.io/gorm"
)

// ─── Models ───────────────────────────────────────────────────────────────────

// exchangeWithCurrency je projekcija JOIN-a core_banking.exchange ⟵ core_banking.valuta.
// GORM-ov Scan() popunjava ovu strukturu po imenu kolone iz SELECT klauzule.
// OpenTime i CloseTime su string jer pgx vraća PostgreSQL TIME kao "HH:MM:SS" string.
type exchangeWithCurrency struct {
	ID           int64  `gorm:"column:id"`
	Name         string `gorm:"column:name"`
	Acronym      string `gorm:"column:acronym"`
	MICCode      string `gorm:"column:mic_code"`
	Polity       string `gorm:"column:polity"`
	CurrencyID   int64  `gorm:"column:currency_id"`
	CurrencyName string `gorm:"column:currency_name"` // iz core_banking.valuta.naziv
	Timezone     string `gorm:"column:timezone"`
	OpenTime     string `gorm:"column:open_time"`  // "HH:MM:SS" — pgx vraća TIME kao string
	CloseTime    string `gorm:"column:close_time"` // "HH:MM:SS" — pgx vraća TIME kao string
}

// parseTimeCol parsira PostgreSQL TIME string ("HH:MM:SS" ili "HH:MM") u time.Time.
// U slučaju greške vraća nulti time.Time (što odgovara 00:00).
func parseTimeCol(s string) time.Time {
	for _, layout := range []string{"15:04:05", "15:04"} {
		if t, err := time.Parse(layout, s); err == nil {
			return t
		}
	}
	return time.Time{}
}

func (r exchangeWithCurrency) toDomain() domain.Exchange {
	return domain.Exchange{
		ID:           r.ID,
		Name:         r.Name,
		Acronym:      r.Acronym,
		MICCode:      r.MICCode,
		Polity:       r.Polity,
		CurrencyID:   r.CurrencyID,
		CurrencyName: r.CurrencyName,
		Timezone:     r.Timezone,
		OpenTime:     parseTimeCol(r.OpenTime),
		CloseTime:    parseTimeCol(r.CloseTime),
	}
}

// exchangeHolidayModel je GORM projekcija tabele core_banking.exchange_holiday.
type exchangeHolidayModel struct {
	ID     int64     `gorm:"column:id;primaryKey"`
	Polity string    `gorm:"column:polity"`
	Date   time.Time `gorm:"column:date"`
}

func (exchangeHolidayModel) TableName() string {
	return "core_banking.exchange_holiday"
}

// ─── Repository ───────────────────────────────────────────────────────────────

type berzaRepository struct {
	db *gorm.DB
}

func NewBerzaRepository(db *gorm.DB) domain.ExchangeRepository {
	return &berzaRepository{db: db}
}

// joinQuery gradi zajednički SELECT sa JOIN-om koji popunjava currency_name.
// Sve tri metode (List, GetByID, GetByMICCode) koriste isti JOIN da bi vratile
// CurrencyName bez duplikovanja SQL-a.
func (r *berzaRepository) joinQuery(ctx context.Context) *gorm.DB {
	return r.db.WithContext(ctx).
		Table("core_banking.exchange e").
		Select("e.id, e.name, e.acronym, e.mic_code, e.polity, e.currency_id, e.timezone, e.open_time, e.close_time, v.naziv AS currency_name").
		Joins("JOIN core_banking.valuta v ON v.id = e.currency_id")
}

func (r *berzaRepository) List(ctx context.Context, filter domain.ListExchangesFilter) ([]domain.Exchange, error) {
	q := r.joinQuery(ctx)
	if filter.Polity != "" {
		like := "%" + strings.ToLower(filter.Polity) + "%"
		q = q.Where("LOWER(e.polity) LIKE ?", like)
	}
	if filter.Search != "" {
		like := "%" + strings.ToLower(filter.Search) + "%"
		q = q.Where("LOWER(e.name) LIKE ? OR LOWER(e.acronym) LIKE ?", like, like)
	}

	var rows []exchangeWithCurrency
	if err := q.Order("e.name ASC").Scan(&rows).Error; err != nil {
		return nil, err
	}

	exchanges := make([]domain.Exchange, 0, len(rows))
	for _, row := range rows {
		exchanges = append(exchanges, row.toDomain())
	}
	return exchanges, nil
}

func (r *berzaRepository) GetByID(ctx context.Context, id int64) (*domain.Exchange, error) {
	var row exchangeWithCurrency
	err := r.joinQuery(ctx).Where("e.id = ?", id).Scan(&row).Error
	if err != nil {
		return nil, err
	}
	if row.ID == 0 {
		return nil, domain.ErrExchangeNotFound
	}
	e := row.toDomain()
	return &e, nil
}

func (r *berzaRepository) GetByMICCode(ctx context.Context, micCode string) (*domain.Exchange, error) {
	var row exchangeWithCurrency
	err := r.joinQuery(ctx).Where("e.mic_code = ?", micCode).Scan(&row).Error
	if err != nil {
		return nil, err
	}
	if row.ID == 0 {
		return nil, domain.ErrExchangeNotFound
	}
	e := row.toDomain()
	return &e, nil
}

// IsHoliday proverava da li je dati datum praznik za datu državu.
// date bi trebalo da bude lokalni datum berze (vreme se ignoriše, samo datum).
func (r *berzaRepository) IsHoliday(ctx context.Context, polity string, date time.Time) (bool, error) {
	// Normalizujemo na početak dana (00:00:00) da bismo izbjegli sat/minut mismatch.
	dateOnly := date.Truncate(24 * time.Hour)

	var count int64
	err := r.db.WithContext(ctx).
		Model(&exchangeHolidayModel{}).
		Where("polity = ? AND date = ?", polity, dateOnly).
		Count(&count).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	return count > 0, nil
}
