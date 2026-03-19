-- =============================================================================
-- Migration: 000005_add_krediti_tables
-- Service:   bank-service
-- Schema:    core_banking
--
-- Tables: kreditni_zahtev, kredit, rata
--
-- Arhitekturne napomene:
--   • kreditni_zahtev čuva podatke koje klijent unosi pri podnošenju zahteva.
--     Ovi podaci (plata, zaposlenje, svrha…) ne ulaze u konačni kredit zapis.
--     Pri odobravanju, servis kreira kredit i menja status zahteva u ODOBREN.
--
--   • kredit čuva kompletno stanje odobrenog kredita sa svim finansijskim
--     veličinama koje servisni sloj izračunava po odobrenju.
--
--   • rata čuva punu amortizacionu tablicu. Uvek postoji bar jedna rata u
--     budućnosti (pre-generated per Issue #3). Cron job (Issue #5) koristi
--     parcijalni indeks idx_rata_cron za brze pretrage dospelih rata.
--
--   • vlasnik_id je BIGINT bez FK — cross-service referenca na user-service.
--   • broj_racuna je VARCHAR FK na racun.broj_racuna (UNIQUE kolona).
-- =============================================================================

-- ─── 1. kreditni_zahtev ──────────────────────────────────────────────────────
-- Čuva sve podatke koje klijent upisuje pri podnošenju zahteva za kredit.

CREATE TABLE core_banking.kreditni_zahtev (
    id                  BIGINT         GENERATED ALWAYS AS IDENTITY PRIMARY KEY,

    -- Klijent koji podnosi zahtev (cross-service; bez FK ograničenja).
    vlasnik_id          BIGINT         NOT NULL,

    -- Vrsta i tip kamate — isti skup vrednosti kao u kredit tabeli.
    vrsta_kredita       VARCHAR(30)    NOT NULL
        CONSTRAINT kz_vrsta_check
            CHECK (vrsta_kredita IN ('GOTOVINSKI','STAMBENI','AUTO','REFINANSIRAJUCI','STUDENTSKI')),
    tip_kamate          VARCHAR(20)    NOT NULL
        CONSTRAINT kz_tip_kamate_check
            CHECK (tip_kamate IN ('FIKSNI','VARIJABILNI')),

    -- Finansijski parametri zahteva.
    iznos_kredita       NUMERIC(15,2)  NOT NULL,
    valuta              VARCHAR(10)    NOT NULL,   -- ISO 4217
    rok_otplate         INTEGER        NOT NULL,   -- u mesecima

    -- Podaci o zaposlenju i kontaktu (specifični za zahtev, ne ulaze u kredit).
    svrha_kredita       TEXT,
    iznos_mesecne_plate NUMERIC(15,2)  NOT NULL,
    status_zaposlenja   VARCHAR(20)    NOT NULL
        CONSTRAINT kz_status_zap_check
            CHECK (status_zaposlenja IN ('STALNO','PRIVREMENO','NEZAPOSLEN')),
    period_zaposlenja   INTEGER        NOT NULL,   -- u mesecima
    kontakt_telefon     VARCHAR(30)    NOT NULL,

    -- Račun na koji će biti uplaćen iznos i sa kojeg se skidaju rate.
    broj_racuna         VARCHAR(18)    NOT NULL
        REFERENCES core_banking.racun (broj_racuna),

    -- Status obrade.
    status              VARCHAR(20)    NOT NULL DEFAULT 'NA_CEKANJU'
        CONSTRAINT kz_status_check
            CHECK (status IN ('NA_CEKANJU','ODOBREN','ODBIJEN')),
    datum_podnosenja    TIMESTAMPTZ    NOT NULL DEFAULT NOW(),

    CONSTRAINT kz_iznos_pozitivan CHECK (iznos_kredita > 0),
    CONSTRAINT kz_rok_pozitivan   CHECK (rok_otplate  > 0)
);

CREATE INDEX idx_kreditni_zahtev_vlasnik
    ON core_banking.kreditni_zahtev (vlasnik_id);

CREATE INDEX idx_kreditni_zahtev_status
    ON core_banking.kreditni_zahtev (status);

CREATE INDEX idx_kreditni_zahtev_racun
    ON core_banking.kreditni_zahtev (broj_racuna);

CREATE INDEX idx_kreditni_zahtev_datum
    ON core_banking.kreditni_zahtev (datum_podnosenja DESC);

-- ─── 2. kredit ───────────────────────────────────────────────────────────────
-- Čuva kompletno stanje aktivnog ili zatvorenog kredita.
-- Kreira se pri odobravanju zahteva; sve finansijske veličine izračunava servis.

CREATE TABLE core_banking.kredit (
    id                      BIGINT        GENERATED ALWAYS AS IDENTITY PRIMARY KEY,

    -- Čitljivi, jedinstveni identifikator kredita (generiše servis).
    broj_kredita            VARCHAR(30)   NOT NULL UNIQUE,

    -- Veza sa zahteviom koji je doveo do ovog kredita (nullable: može se
    -- kreirati i bez zahteva u edge case scenarijima ili seeding-u).
    kreditni_zahtev_id      BIGINT
        REFERENCES core_banking.kreditni_zahtev (id),

    -- Račun klijenta (denormalizovano radi brzih pretraga u cron job-u).
    broj_racuna             VARCHAR(18)   NOT NULL
        REFERENCES core_banking.racun (broj_racuna),

    -- Vlasnik (cross-service; bez FK ograničenja).
    vlasnik_id              BIGINT        NOT NULL,

    -- Vrsta i tip kamate.
    vrsta_kredita           VARCHAR(30)   NOT NULL
        CONSTRAINT k_vrsta_check
            CHECK (vrsta_kredita IN ('GOTOVINSKI','STAMBENI','AUTO','REFINANSIRAJUCI','STUDENTSKI')),
    tip_kamate              VARCHAR(20)   NOT NULL
        CONSTRAINT k_tip_kamate_check
            CHECK (tip_kamate IN ('FIKSNI','VARIJABILNI')),

    -- Finansijske veličine.
    iznos_kredita           NUMERIC(15,2) NOT NULL,
    period_otplate          INTEGER       NOT NULL,   -- inicijalni broj rata (meseci)
    nominalna_kamatna_stopa NUMERIC(7,4)  NOT NULL,   -- godišnja nominalna stopa, npr. 6.5000
    efektivna_kamatna_stopa NUMERIC(7,4)  NOT NULL,   -- efektivna kamatna stopa (EKS)
    iznos_mesecne_rate      NUMERIC(15,2) NOT NULL,
    preostalo_dugovanje     NUMERIC(15,2) NOT NULL,
    valuta                  VARCHAR(10)   NOT NULL,

    -- Vremenski okvir kredita.
    datum_ugovaranja        DATE          NOT NULL,
    datum_isplate           DATE,                     -- NULL dok novac nije uplaćen na račun
    datum_sledece_rate      DATE,                     -- NULL za OTPLACEN kredit

    -- Status kredita.
    status                  VARCHAR(20)   NOT NULL DEFAULT 'ODOBREN'
        CONSTRAINT k_status_check
            CHECK (status IN ('ODOBREN','OTPLACEN','U_KASNJENJU')),

    created_at              TIMESTAMPTZ   NOT NULL DEFAULT NOW(),

    CONSTRAINT k_iznos_pozitivan  CHECK (iznos_kredita   > 0),
    CONSTRAINT k_period_pozitivan CHECK (period_otplate  > 0)
);

CREATE INDEX idx_kredit_vlasnik
    ON core_banking.kredit (vlasnik_id);

CREATE INDEX idx_kredit_broj_racuna
    ON core_banking.kredit (broj_racuna);

-- Za listu klijenata (sortirano opadajuće po iznosu per Issue #1).
CREATE INDEX idx_kredit_vlasnik_iznos
    ON core_banking.kredit (vlasnik_id, iznos_kredita DESC);

CREATE INDEX idx_kredit_status
    ON core_banking.kredit (status);

-- Parcijalni indeks za cron job (Issue #5): samo aktivni krediti sa budućim ratama.
CREATE INDEX idx_kredit_sledeca_rata
    ON core_banking.kredit (datum_sledece_rate)
    WHERE status = 'ODOBREN';

-- ─── 3. rata ─────────────────────────────────────────────────────────────────
-- Čuva punu amortizacionu tablicu svakog kredita.
-- Uvek postoji bar jedna rata u budućnosti (pre-generisana per Issue #3).
-- Polja broj_pokusaja i sledeci_pokusaj podržavaju retry logiku (Issue #5).
-- Pravi datum dospeća se popunjava isključivo pri uspešnom skidanju sredstava.

CREATE TABLE core_banking.rata (
    id                      BIGINT        GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    kredit_id               BIGINT        NOT NULL
        REFERENCES core_banking.kredit (id),

    -- Finansijske veličine ove konkretne rate.
    iznos_rate              NUMERIC(15,2) NOT NULL,   -- ukupan iznos rate (A)
    iznos_kamate            NUMERIC(15,2) NOT NULL,   -- kamatni deo ove rate
    valuta                  VARCHAR(10)   NOT NULL,

    -- Datumi.
    ocekivani_datum_dospeca DATE          NOT NULL,
    -- Popunjava se isključivo kada novac bude uspešno skinut (Issue #3 edge case).
    pravi_datum_dospeca     DATE,

    -- Status plaćanja.
    status_placanja         VARCHAR(20)   NOT NULL DEFAULT 'NEPLACENO'
        CONSTRAINT r_status_check
            CHECK (status_placanja IN ('PLACENO','NEPLACENO','KASNI')),

    -- Retry mehanizam (Issue #5): broj pokušaja naplate i termin sledećeg pokušaja.
    broj_pokusaja           INTEGER       NOT NULL DEFAULT 0,
    sledeci_pokusaj         TIMESTAMPTZ,

    CONSTRAINT r_iznos_rate_pozitivan   CHECK (iznos_rate   > 0),
    CONSTRAINT r_iznos_kamate_pozitivan CHECK (iznos_kamate >= 0)
);

CREATE INDEX idx_rata_kredit_id
    ON core_banking.rata (kredit_id);

-- Parcijalni indeks za cron job — samo neizmirene rate (Issue #5).
-- Pokriva i NEPLACENO (prvi pokušaj) i KASNI (retry pokušaji).
CREATE INDEX idx_rata_cron
    ON core_banking.rata (ocekivani_datum_dospeca, sledeci_pokusaj)
    WHERE status_placanja IN ('NEPLACENO','KASNI');
