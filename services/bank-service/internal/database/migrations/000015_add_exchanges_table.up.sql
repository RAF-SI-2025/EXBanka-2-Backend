-- =============================================================================
-- Migration: 000015_add_exchanges_table
-- Service:   bank-service
-- Schema:    core_banking
--
-- Table: exchange
--
-- Čuva informacije o berzama hartija od vrednosti.
-- currency_id je FK na core_banking.valuta.
-- =============================================================================

CREATE TABLE core_banking.exchange (
    id          BIGINT       GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name        VARCHAR(255) NOT NULL,
    acronym     VARCHAR(50)  NOT NULL,
    mic_code    VARCHAR(10)  NOT NULL UNIQUE,  -- ISO 10383 MIC kod
    polity      VARCHAR(100) NOT NULL,         -- država/politički entitet
    currency_id BIGINT       NOT NULL REFERENCES core_banking.valuta (id),
    timezone    VARCHAR(100) NOT NULL          -- IANA timezone, e.g. "America/New_York"
);

CREATE INDEX idx_exchange_polity      ON core_banking.exchange (polity);
CREATE INDEX idx_exchange_currency_id ON core_banking.exchange (currency_id);
