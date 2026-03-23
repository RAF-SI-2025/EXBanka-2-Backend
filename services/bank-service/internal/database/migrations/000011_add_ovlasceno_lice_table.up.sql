-- =============================================================================
-- Migration: 000011_add_ovlasceno_lice_table
-- Service:   bank-service
-- Schema:    core_banking
--
-- Dodaje tabelu za ovlašćena lica vezana za kartice poslovnih računa.
--
-- Biznis pravila:
--   • Vlasnik poslovnog računa može zatražiti karticu za "ovlašćeno lice"
--     (npr. radnika firme).
--   • Entitet postoji isključivo radi praćenja kome pripada izdata kartica,
--     zbog provere limita "max 1 kartica po osobi".
--   • Relacija je 1:1 — jedna kartica ima najviše jedno ovlašćeno lice
--     (ovlašćeno lice je child entitet).
--   • ON DELETE CASCADE — brisanjem kartice automatski se briše i zapis
--     ovlašćenog lica (kartica je vlasnik relacije).
-- =============================================================================

-- ─── 1. ovlasceno_lice ────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS core_banking.ovlasceno_lice (
    id              BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,

    kartica_id      BIGINT         NOT NULL,

    ime             VARCHAR(100),
    prezime         VARCHAR(100),
    pol             VARCHAR(20),
    email_adresa    VARCHAR(255),
    broj_telefona   VARCHAR(30),
    adresa          VARCHAR(255),

    -- Unix timestamp (sekunde od 1970-01-01 UTC)
    datum_rodjenja  BIGINT,

    CONSTRAINT fk_ovlasceno_lice_kartica
        FOREIGN KEY (kartica_id)
        REFERENCES core_banking.kartica(id)
        ON DELETE CASCADE
);

-- ─── 2. Indeks ────────────────────────────────────────────────────────────────

-- Brza provera limita: "da li dato lice već ima karticu?"
CREATE INDEX IF NOT EXISTS idx_ovlasceno_lice_kartica_id
    ON core_banking.ovlasceno_lice (kartica_id);
