// Package service — exchange (berza) business logic.
package service

import (
	"context"
	"fmt"
	"time"

	"banka-backend/services/bank-service/internal/domain"
)

type berzaService struct {
	repo      domain.ExchangeRepository
	modeStore domain.MarketModeStore
}

func NewBerzaService(repo domain.ExchangeRepository, modeStore domain.MarketModeStore) domain.BerzaService {
	return &berzaService{repo: repo, modeStore: modeStore}
}

func (s *berzaService) ListExchanges(ctx context.Context, filter domain.ListExchangesFilter) ([]domain.Exchange, error) {
	return s.repo.List(ctx, filter)
}

func (s *berzaService) GetExchange(ctx context.Context, id int64, micCode string) (*domain.Exchange, error) {
	if micCode != "" {
		return s.repo.GetByMICCode(ctx, micCode)
	}
	return s.repo.GetByID(ctx, id)
}

// IsExchangeOpen proverava da li berza trenutno radi.
// Radno vreme: 09:30–16:00 po lokalnom vremenu berze.
// isAfterHours: true ako je prošlo manje od 4 sata od zatvaranja (16:00–20:00).
// Ako je market:test_mode=true u Redisu, uvek vraća isOpen=true.
func (s *berzaService) IsExchangeOpen(ctx context.Context, exchangeID int64) (isOpen bool, isAfterHours bool, err error) {
	// 1. Proveri Redis override
	testMode, err := s.modeStore.IsTestMode(ctx)
	if err != nil {
		return false, false, fmt.Errorf("IsTestMode: %w", err)
	}
	if testMode {
		return true, false, nil
	}

	// 2. Učitaj berzu
	ex, err := s.repo.GetByID(ctx, exchangeID)
	if err != nil {
		return false, false, err
	}

	// 3. Konvertuj trenutno vreme u vremensku zonu berze
	loc, err := time.LoadLocation(ex.Timezone)
	if err != nil {
		return false, false, fmt.Errorf("nevalidna vremenska zona %q: %w", ex.Timezone, err)
	}
	now := time.Now().In(loc)

	// 4. Definiši granice radnog vremena za tekući dan
	openTime  := time.Date(now.Year(), now.Month(), now.Day(), 9, 30, 0, 0, loc)
	closeTime := time.Date(now.Year(), now.Month(), now.Day(), 16, 0, 0, 0, loc)

	// 5. Berza je otvorena ako je vreme između 09:30 i 16:00
	if !now.Before(openTime) && now.Before(closeTime) {
		return true, false, nil
	}

	// 6. After-hours: od 16:00 do 20:00 (4 sata posle zatvaranja)
	afterHoursEnd := closeTime.Add(4 * time.Hour)
	if !now.Before(closeTime) && now.Before(afterHoursEnd) {
		return false, true, nil
	}

	return false, false, nil
}

func (s *berzaService) ToggleMarketTestMode(ctx context.Context, enabled bool) error {
	return s.modeStore.SetTestMode(ctx, enabled)
}
