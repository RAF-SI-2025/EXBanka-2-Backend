-- =============================================================================
-- Migration: 000010_seed_banka_firma
-- Service:   bank-service
-- Schema:    core_banking
--
-- Seeds EXBanka a.d. as a Firma and opens 8 treasury accounts (1 RSD + 7 FX).
-- These accounts serve as the bank's own pool for commission collection and
-- exchange-rate operations.
--
-- Prerequisites (must already be applied):
--   000007_seed_valute    – RSD, EUR, USD, CHF, GBP, JPY, CAD, AUD
--   000008_seed_delatnosti – sifra '64.19' already seeded there
--   user-service/000002   – trezor@exbanka.rs gets id = 2 (admin=1 in 000001)
--
-- NOTE: vlasnik_id = 2 and id_zaposlenog = 2 are plain BIGINT cross-service
--       references to trezor@exbanka.rs in user-service. No FK constraint.
--
-- Account number derivation (see account_service.go → generateAccountNumber):
--   prefix   = bankCode(666) + branchCode(0001) + typeCode
--   typeCode = 12 for TEKUCI/POSLOVNI, 22 for DEVIZNI/POSLOVNI
--   check    = (11 - digitSum(prefix + random8) % 11) % 11
--
--   RSD  666000112 + 20000000 → digitSum=24, check=9 → 666000112200000009
--   EUR  666000122 + 10000000 → digitSum=24, check=9 → 666000122100000009
--   USD  666000122 + 20000000 → digitSum=25, check=8 → 666000122200000008
--   CHF  666000122 + 30000000 → digitSum=26, check=7 → 666000122300000007
--   GBP  666000122 + 40000000 → digitSum=27, check=6 → 666000122400000006
--   JPY  666000122 + 50000000 → digitSum=28, check=5 → 666000122500000005
--   CAD  666000122 + 60000000 → digitSum=29, check=4 → 666000122600000004
--   AUD  666000122 + 70000000 → digitSum=30, check=3 → 666000122700000003
-- =============================================================================

-- ─── 1. Firma: EXBanka a.d. ──────────────────────────────────────────────────
-- Delatnost 64.19 already exists from 000008_seed_delatnosti; resolved by sifra.

INSERT INTO core_banking.firma (naziv_firme, maticni_broj, poreski_broj, id_delatnosti, adresa, vlasnik_id)
SELECT
    'EXBanka a.d.',
    '12345678',
    '101234567',
    d.id,
    'Bulevar Oslobođenja 1, Beograd',
    2   -- trezor@exbanka.rs in user-service (GENERATED id = 2)
FROM core_banking.delatnost d
WHERE d.sifra = '64.19'
ON CONFLICT (maticni_broj) DO NOTHING;

-- ─── 2. Računi trezora ───────────────────────────────────────────────────────

INSERT INTO core_banking.racun (
    broj_racuna,
    id_zaposlenog,
    id_vlasnika,
    id_firme,
    id_valute,
    kategorija_racuna,
    vrsta_racuna,
    naziv_racuna,
    stanje_racuna,
    datum_kreiranja,
    datum_isteka,
    status
)
SELECT
    r.broj_racuna,
    2,                          -- id_zaposlenog = trezor@exbanka.rs
    2,                          -- id_vlasnika   = trezor@exbanka.rs
    f.id,
    v.id,
    r.kategorija_racuna,
    'POSLOVNI',
    r.naziv_racuna,
    r.stanje_racuna,
    NOW(),
    NOW() + INTERVAL '10 years',
    'AKTIVAN'
FROM core_banking.firma f
CROSS JOIN (VALUES
    -- (broj_racuna,              oznaka, kategorija,   naziv_racuna,               stanje)
    ('666000112200000009', 'RSD', 'TEKUCI',  'EXBanka Trezor RSD',  10000000000.00),
    ('666000122100000009', 'EUR', 'DEVIZNI', 'EXBanka Trezor EUR',    100000000.00),
    ('666000122200000008', 'USD', 'DEVIZNI', 'EXBanka Trezor USD',    100000000.00),
    ('666000122300000007', 'CHF', 'DEVIZNI', 'EXBanka Trezor CHF',    100000000.00),
    ('666000122400000006', 'GBP', 'DEVIZNI', 'EXBanka Trezor GBP',    100000000.00),
    ('666000122500000005', 'JPY', 'DEVIZNI', 'EXBanka Trezor JPY',    100000000.00),
    ('666000122600000004', 'CAD', 'DEVIZNI', 'EXBanka Trezor CAD',    100000000.00),
    ('666000122700000003', 'AUD', 'DEVIZNI', 'EXBanka Trezor AUD',    100000000.00)
) AS r(broj_racuna, oznaka, kategorija_racuna, naziv_racuna, stanje_racuna)
JOIN core_banking.valuta v ON v.oznaka = r.oznaka
WHERE f.maticni_broj = '12345678'
ON CONFLICT (broj_racuna) DO NOTHING;
