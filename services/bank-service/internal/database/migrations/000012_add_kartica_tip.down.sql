ALTER TABLE core_banking.kartica
    DROP CONSTRAINT IF EXISTS kartica_tip_kartice_check,
    DROP COLUMN IF EXISTS tip_kartice,
    DROP COLUMN IF EXISTS provizija_procenat,
    DROP COLUMN IF EXISTS konverziona_naknada_procenat;

-- Vraćamo naziv_kartice (nullable jer ne znamo originalne vrednosti).
ALTER TABLE core_banking.kartica
    ADD COLUMN IF NOT EXISTS naziv_kartice VARCHAR(100);
