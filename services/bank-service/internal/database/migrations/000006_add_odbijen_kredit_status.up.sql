-- =============================================================================
-- Migration: 000006_add_odbijen_kredit_status
-- Service:   bank-service
-- Schema:    core_banking
--
-- Svrha:
--   Dozvoljava status 'ODBIJEN' u tabeli kredit kako bi odbijeni zahtevi
--   bili vidljivi u zaposlenom portalu ("Svi krediti").
--
--   Pri odbijanju zahteva, repozitorijum atomski upisuje ledger zapis u
--   kredit tabelu sa status = 'ODBIJEN' i nultim finansijskim veličinama.
--   GetAllCredits query prirodno vraća i te zapise.
-- =============================================================================

ALTER TABLE core_banking.kredit
    DROP CONSTRAINT k_status_check;

ALTER TABLE core_banking.kredit
    ADD CONSTRAINT k_status_check
        CHECK (status IN ('ODOBREN', 'OTPLACEN', 'U_KASNJENJU', 'ODBIJEN'));
