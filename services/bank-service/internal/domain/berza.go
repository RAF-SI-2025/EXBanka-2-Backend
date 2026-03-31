package domain

import (
	"context"
	"errors"
)

// ─── Greške ───────────────────────────────────────────────────────────────────

var (
	ErrExchangeNotFound = errors.New("berza nije pronađena")
)

// ─── Exchange ─────────────────────────────────────────────────────────────────

// Exchange je čisti domenski objekat za berzu — ne zna za GORM niti za gRPC.
type Exchange struct {
	ID         int64
	Name       string
	Acronym    string
	MICCode    string
	Polity     string
	CurrencyID int64
	Timezone   string // IANA timezone, e.g. "America/New_York"
}

// ListExchangesFilter parametri za filtriranje liste berzi.
type ListExchangesFilter struct {
	Polity string // tačno poklapanje; "" = bez filtera
	Search string // parcijalni match na name ili acronym; "" = bez filtera
}

// ExchangeRepository definiše ugovor prema sloju podataka.
type ExchangeRepository interface {
	List(ctx context.Context, filter ListExchangesFilter) ([]Exchange, error)
	GetByID(ctx context.Context, id int64) (*Exchange, error)
	GetByMICCode(ctx context.Context, micCode string) (*Exchange, error)
}

// MarketModeStore apstrahuje čuvanje zastavice za bypass radnog vremena berzi.
// Implementacija živi u transport paketu (Redis), NoOp fallback kada Redis nije dostupan.
type MarketModeStore interface {
	SetTestMode(ctx context.Context, enabled bool) error
	IsTestMode(ctx context.Context) (bool, error)
}

// BerzaService definiše ugovor prema sloju poslovne logike.
type BerzaService interface {
	ListExchanges(ctx context.Context, filter ListExchangesFilter) ([]Exchange, error)
	GetExchange(ctx context.Context, id int64, micCode string) (*Exchange, error)
	// IsExchangeOpen proverava da li je berza trenutno otvorena.
	// Berza radi od 09:30 do 16:00 po lokalnom vremenu (timezone polje).
	// isAfterHours je true ako je prošlo manje od 4 sata od zatvaranja (16:00).
	// Ako je uključen market test mode (Redis: market:test_mode=true),
	// uvek vraća isOpen=true bez obzira na vreme.
	IsExchangeOpen(ctx context.Context, exchangeID int64) (isOpen bool, isAfterHours bool, err error)
	ToggleMarketTestMode(ctx context.Context, enabled bool) error
}
