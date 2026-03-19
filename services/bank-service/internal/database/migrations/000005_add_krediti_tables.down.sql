-- =============================================================================
-- Migration: 000005_add_krediti_tables (rollback)
-- =============================================================================

-- Brisanje u obrnutom redosledu od kreiranja (rata zavisi od kredit,
-- kredit zavisi od kreditni_zahtev, oba zavise od racun).
DROP TABLE IF EXISTS core_banking.rata;
DROP TABLE IF EXISTS core_banking.kredit;
DROP TABLE IF EXISTS core_banking.kreditni_zahtev;
