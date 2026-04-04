-- =============================================================================
-- Migration: 000025_deactivate_exchange_only_currencies
-- Service:   bank-service
-- Schema:    core_banking
--
-- Deaktivira valute koje su dodate isključivo radi FK veze sa tabelom berze.
-- status = FALSE → ne pojavljuju se u UI-u (menjačnica, kreiranje računa, itd.)
-- ali ostaju u bazi kako bi FK ograničenje na core_banking.exchange funkcionisalo.
-- =============================================================================

UPDATE core_banking.valuta
SET status = FALSE
WHERE oznaka IN ('IDR', 'INR', 'PLN', 'UAH', 'ARS', 'TRY', 'SEK', 'HUF', 'BGN');
