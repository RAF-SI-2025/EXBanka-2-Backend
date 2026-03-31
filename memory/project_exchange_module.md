---
name: Exchange module implementation
description: Exchanges (berze) modul implementiran u bank-service u martu 2026 — fajlovi, arhitektura, odluke
type: project
---

Implementiran modul za berze (stock exchanges) unutar bank-service.

**Why:** Specifikacija zahteva evidenciju berzi, proveru radnog vremena i admin toggle za testiranje.

**How to apply:** Kodni stil prati isti Clean Architecture pattern kao ostatak bank-service (domain → service → repository → handler).

## Novi fajlovi

- `proto/banka/banka.proto` — dodati `Exchange`, `ListExchangesRequest/Response`, `GetExchangeRequest`, `ToggleMarketTestModeRequest` + 3 RPC-a
- `internal/domain/berza.go` — `Exchange`, `ListExchangesFilter`, `ExchangeRepository`, `MarketModeStore`, `BerzaService` interfejsi
- `internal/repository/berza_repository.go` — GORM repo (`NewBerzaRepository`)
- `internal/service/berza_service.go` — `NewBerzaService`, `IsExchangeOpen` (09:30–16:00 lokalno; after-hours 4h; Redis override)
- `internal/transport/redis_market.go` — `RedisMarketModeStore` + `NoOpMarketModeStore`; Redis ključ: `market:test_mode`
- `internal/handler/berza_handler.go` — handler metode na `BankHandler`: `ListExchanges`, `GetExchange`, `ToggleMarketTestMode`
- `internal/database/migrations/000015_add_exchanges_table.up/down.sql` — tabela `core_banking.exchange`
- `internal/database/migrations/000016_seed_exchanges.up/down.sql` — seeder za 23 berze
- `data/exchanges.csv` — izvorni CSV (name, acronym, mic_code, polity, currency_code, timezone)

## Izmenjeni fajlovi

- `internal/handler/grpc_handler.go` — dodat `berzaService domain.BerzaService` u `BankHandler` + `NewBankHandler`
- `cmd/server/main.go` — wire-up berzaRepo, marketModeStore, berzaService
- `tests/bdd/krediti_steps_test.go` — dodat nil argument za berzaService u `NewBankHandler`

## HTTP endpointi

- `GET /bank/exchanges` — lista berzi (query: polity, search)
- `GET /bank/exchanges/{id}` — po ID
- `GET /bank/exchanges/mic/{mic_code}` — po MIC kodu
- `POST /bank/admin/exchanges/test-mode` — toggle (samo EMPLOYEE)
