-- Rollback: uklanja 'ODBIJEN' iz dozvoljenih statusa u tabeli kredit.
-- UPOZORENJE: Ovaj rollback će biti odbijen od strane baze ako postoje
-- redovi sa status = 'ODBIJEN'. Pre rollback-a ručno obrišite te redove.

ALTER TABLE core_banking.kredit
    DROP CONSTRAINT k_status_check;

ALTER TABLE core_banking.kredit
    ADD CONSTRAINT k_status_check
        CHECK (status IN ('ODOBREN', 'OTPLACEN', 'U_KASNJENJU'));
