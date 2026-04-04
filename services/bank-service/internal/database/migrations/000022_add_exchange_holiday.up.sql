-- =============================================================================
-- Migration: 000022_add_exchange_holiday
-- Service:   bank-service
-- Schema:    core_banking
--
-- Table: exchange_holiday
--
-- Čuva datume praznika po državama (polity).
-- IsExchangeOpen u berza_service.go proverava ovu tabelu pre nego što
-- odluči da li je berza otvorena.
--
-- Kompozitni UNIQUE indeks (polity, date) osigurava jedinstvenost i ubrzava
-- lookup po paru (polity, lokalni datum) koji se izvršava pri svakom pozivu
-- IsExchangeOpen.
--
-- Podatke unositi naknadno kroz posebnu seed skriptu.
-- =============================================================================

CREATE TABLE core_banking.exchange_holiday (
    id     BIGINT       GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    polity VARCHAR(100) NOT NULL,  -- e.g. "United States", "Germany"
    date   DATE         NOT NULL,  -- lokalni datum praznika za datu državu
    CONSTRAINT uq_exchange_holiday UNIQUE (polity, date)
);

-- Primarni indeks za brzo čitanje: IsExchangeOpen filtrira po (polity, date).
CREATE INDEX idx_exchange_holiday_polity_date
    ON core_banking.exchange_holiday (polity, date);
