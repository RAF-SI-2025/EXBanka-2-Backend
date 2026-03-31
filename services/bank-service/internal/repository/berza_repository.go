package repository

import (
	"context"
	"errors"
	"strings"

	"banka-backend/services/bank-service/internal/domain"

	"gorm.io/gorm"
)

// exchangeModel je GORM projekcija tabele core_banking.exchange.
type exchangeModel struct {
	ID         int64  `gorm:"column:id;primaryKey"`
	Name       string `gorm:"column:name"`
	Acronym    string `gorm:"column:acronym"`
	MICCode    string `gorm:"column:mic_code"`
	Polity     string `gorm:"column:polity"`
	CurrencyID int64  `gorm:"column:currency_id"`
	Timezone   string `gorm:"column:timezone"`
}

func (exchangeModel) TableName() string {
	return "core_banking.exchange"
}

func (m exchangeModel) toDomain() domain.Exchange {
	return domain.Exchange{
		ID:         m.ID,
		Name:       m.Name,
		Acronym:    m.Acronym,
		MICCode:    m.MICCode,
		Polity:     m.Polity,
		CurrencyID: m.CurrencyID,
		Timezone:   m.Timezone,
	}
}

type berzaRepository struct {
	db *gorm.DB
}

func NewBerzaRepository(db *gorm.DB) domain.ExchangeRepository {
	return &berzaRepository{db: db}
}

func (r *berzaRepository) List(ctx context.Context, filter domain.ListExchangesFilter) ([]domain.Exchange, error) {
	q := r.db.WithContext(ctx).Model(&exchangeModel{})
	if filter.Polity != "" {
		q = q.Where("polity = ?", filter.Polity)
	}
	if filter.Search != "" {
		like := "%" + strings.ToLower(filter.Search) + "%"
		q = q.Where("LOWER(name) LIKE ? OR LOWER(acronym) LIKE ?", like, like)
	}
	var models []exchangeModel
	if err := q.Order("name ASC").Find(&models).Error; err != nil {
		return nil, err
	}
	exchanges := make([]domain.Exchange, 0, len(models))
	for _, m := range models {
		exchanges = append(exchanges, m.toDomain())
	}
	return exchanges, nil
}

func (r *berzaRepository) GetByID(ctx context.Context, id int64) (*domain.Exchange, error) {
	var m exchangeModel
	if err := r.db.WithContext(ctx).First(&m, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrExchangeNotFound
		}
		return nil, err
	}
	e := m.toDomain()
	return &e, nil
}

func (r *berzaRepository) GetByMICCode(ctx context.Context, micCode string) (*domain.Exchange, error) {
	var m exchangeModel
	if err := r.db.WithContext(ctx).Where("mic_code = ?", micCode).First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrExchangeNotFound
		}
		return nil, err
	}
	e := m.toDomain()
	return &e, nil
}
