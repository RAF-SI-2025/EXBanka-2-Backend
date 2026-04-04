-- =============================================================================
-- Migration: 000023_add_exchange_times
-- Service:   bank-service
-- Schema:    core_banking
--
-- Dodaje kolone open_time i close_time u tabelu exchange.
-- Popunjava vrednosti za 23 vec postojece berze.
-- =============================================================================

ALTER TABLE core_banking.exchange
    ADD COLUMN open_time  TIME NOT NULL DEFAULT '09:30:00',
    ADD COLUMN close_time TIME NOT NULL DEFAULT '16:00:00';

ALTER TABLE core_banking.exchange
    ALTER COLUMN open_time  DROP DEFAULT,
    ALTER COLUMN close_time DROP DEFAULT;

-- Azuriraj postojecih 23 berze tacnim radnim vremenom
UPDATE core_banking.exchange SET open_time = '09:30', close_time = '16:00' WHERE mic_code = 'XNYS'; -- New York Stock Exchange
UPDATE core_banking.exchange SET open_time = '09:30', close_time = '16:00' WHERE mic_code = 'XNAS'; -- NASDAQ
UPDATE core_banking.exchange SET open_time = '09:30', close_time = '16:00' WHERE mic_code = 'XASE'; -- NYSE American
UPDATE core_banking.exchange SET open_time = '08:30', close_time = '15:15' WHERE mic_code = 'XCBO'; -- Chicago Board Options Exchange
UPDATE core_banking.exchange SET open_time = '08:00', close_time = '16:30' WHERE mic_code = 'XLON'; -- London Stock Exchange
UPDATE core_banking.exchange SET open_time = '09:00', close_time = '17:30' WHERE mic_code = 'XETR'; -- Deutsche Boerse XETRA
UPDATE core_banking.exchange SET open_time = '09:00', close_time = '17:30' WHERE mic_code = 'XPAR'; -- Euronext Paris
UPDATE core_banking.exchange SET open_time = '09:00', close_time = '17:25' WHERE mic_code = 'XMIL'; -- Borsa Italiana
UPDATE core_banking.exchange SET open_time = '09:00', close_time = '17:30' WHERE mic_code = 'XAMS'; -- Euronext Amsterdam
UPDATE core_banking.exchange SET open_time = '09:00', close_time = '17:30' WHERE mic_code = 'XBRU'; -- Euronext Brussels
UPDATE core_banking.exchange SET open_time = '09:00', close_time = '17:30' WHERE mic_code = 'XMAD'; -- Bolsa de Madrid
UPDATE core_banking.exchange SET open_time = '09:00', close_time = '17:30' WHERE mic_code = 'XVIE'; -- Vienna Stock Exchange
UPDATE core_banking.exchange SET open_time = '08:00', close_time = '16:30' WHERE mic_code = 'XLIS'; -- Euronext Lisbon
UPDATE core_banking.exchange SET open_time = '08:00', close_time = '16:30' WHERE mic_code = 'XDUB'; -- Irish Stock Exchange
UPDATE core_banking.exchange SET open_time = '10:15', close_time = '17:20' WHERE mic_code = 'ASEX'; -- Athens Exchange
UPDATE core_banking.exchange SET open_time = '10:00', close_time = '18:30' WHERE mic_code = 'XHEL'; -- Helsinki Stock Exchange
UPDATE core_banking.exchange SET open_time = '09:00', close_time = '15:30' WHERE mic_code = 'XTKS'; -- Tokyo Stock Exchange
UPDATE core_banking.exchange SET open_time = '09:00', close_time = '15:30' WHERE mic_code = 'XOSE'; -- Osaka Exchange
UPDATE core_banking.exchange SET open_time = '09:30', close_time = '16:00' WHERE mic_code = 'XTSE'; -- Toronto Stock Exchange
UPDATE core_banking.exchange SET open_time = '09:30', close_time = '16:00' WHERE mic_code = 'XTSX'; -- TSX Venture Exchange
UPDATE core_banking.exchange SET open_time = '10:00', close_time = '16:00' WHERE mic_code = 'XASX'; -- Australian Securities Exchange
UPDATE core_banking.exchange SET open_time = '09:00', close_time = '17:30' WHERE mic_code = 'XSWX'; -- SIX Swiss Exchange
UPDATE core_banking.exchange SET open_time = '10:00', close_time = '14:00' WHERE mic_code = 'XBEL'; -- Belgrade Stock Exchange
