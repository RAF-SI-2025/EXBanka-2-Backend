-- =============================================================================
-- Migration: 000009_add_kartica_table
-- Service:   bank-service
-- Schema:    core_banking
--
-- Dodaje tabelu za platne kartice vezane za račune.
--
-- Finansijske napomene:
--   • broj_kartice    — u produkciji treba da bude tokenizovan (PCI-DSS); ovde
--                       se čuva samo poslednje 4 cifre ili PAN token.
--   • cvv_kod         — PCI-DSS zahteva da se CVV NIKAD ne čuva nakon
--                       autorizacije. Ukoliko je čuvanje neophodno, koristiti
--                       HMAC-SHA256 sa tajnim ključem iz env varijable (pepper).
--                       bcrypt NIJE prikladan — CVV ima samo 1000 kombinacija
--                       (000–999) pa je brute-force trivijalan bez tajnog ključa.
--                       HMAC-SHA256 output je uvek tačno 64 hex karaktera.
--   • limit_kartice   — NUMERIC garantuje egzaktnu decimalnu aritmetiku bez
--                       floating-point grešaka, kritično za finansijske iznose.
-- =============================================================================

-- ─── 1. kartica ──────────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS core_banking.kartica (
    id              BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,

    broj_kartice    VARCHAR(16)    NOT NULL,
    vrsta_kartice   VARCHAR(50)    NOT NULL
        CONSTRAINT kartica_vrsta_check     CHECK (vrsta_kartice IN ('DEBIT', 'CREDIT')),

    naziv_kartice   VARCHAR(100)   NOT NULL,   -- npr. Visa, Mastercard

    datum_kreiranja TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    datum_isteka    DATE           NOT NULL,

    racun_id        BIGINT         NOT NULL
        REFERENCES core_banking.racun(id),

    -- PCI-DSS: HMAC-SHA256(cvv, pepper) — tačno 64 hex karaktera, nikad plain-text.
    cvv_kod         CHAR(64)       NOT NULL,

    limit_kartice   NUMERIC(15,2)  NOT NULL DEFAULT 0.00,

    status          VARCHAR(20)    NOT NULL DEFAULT 'AKTIVNA'
        CONSTRAINT kartica_status_check    CHECK (status IN ('AKTIVNA', 'BLOKIRANA', 'DEAKTIVIRANA')),

    CONSTRAINT kartica_broj_kartice_unique UNIQUE (broj_kartice)
);

-- ─── 2. Indeksi ───────────────────────────────────────────────────────────────

-- Brzo pronalaženje svih kartica jednog računa.
CREATE INDEX IF NOT EXISTS idx_kartica_racun_id
    ON core_banking.kartica (racun_id);

-- Filtriranje po statusu (npr. sve aktivne kartice).
CREATE INDEX IF NOT EXISTS idx_kartica_status
    ON core_banking.kartica (status);

-- Brza provera isteklih kartica (batch job za deaktivaciju).
CREATE INDEX IF NOT EXISTS idx_kartica_datum_isteka
    ON core_banking.kartica (datum_isteka);
