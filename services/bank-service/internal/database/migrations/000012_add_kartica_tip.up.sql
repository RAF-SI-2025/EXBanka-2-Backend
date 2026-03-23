-- =============================================================================
-- Migration: 000012_add_kartica_tip
-- Service:   bank-service
-- Schema:    core_banking
--
-- Dodaje podršku za više vrsta kartica (Visa, Mastercard, DinaCard, Amex) i
-- kolone za provizije koje se primenjuju kod Mastercard kartica na RSD računima.
-- =============================================================================

ALTER TABLE core_banking.kartica
    ADD COLUMN IF NOT EXISTS tip_kartice VARCHAR(20) NOT NULL DEFAULT 'VISA',
    ADD COLUMN IF NOT EXISTS provizija_procenat          NUMERIC(5,4) NULL,
    ADD COLUMN IF NOT EXISTS konverziona_naknada_procenat NUMERIC(5,4) NULL;

ALTER TABLE core_banking.kartica
    ADD CONSTRAINT kartica_tip_kartice_check
        CHECK (tip_kartice IN ('VISA', 'MASTERCARD', 'DINACARD', 'AMEX'));

-- naziv_kartice je redundantno uz tip_kartice — uklanjamo ga.
ALTER TABLE core_banking.kartica
    DROP COLUMN IF EXISTS naziv_kartice;

-- broj_kartice je VARCHAR(16) što pokriva i Amex (15 cifara < 16).
