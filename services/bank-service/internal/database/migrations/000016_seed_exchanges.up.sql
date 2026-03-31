-- =============================================================================
-- Migration: 000016_seed_exchanges
-- Service:   bank-service
-- Schema:    core_banking
--
-- Seeder za tabelu exchange.
-- Podaci su preuzeti iz data/exchanges.csv.
-- currency_id se mapira iz oznake valute subupitom na core_banking.valuta.
-- Berze čija valuta ne postoji u bazi se preskači (ON CONFLICT DO NOTHING +
-- NOT EXISTS guard na valuta lookup-u).
-- =============================================================================

INSERT INTO core_banking.exchange (name, acronym, mic_code, polity, currency_id, timezone)
SELECT t.name, t.acronym, t.mic_code, t.polity, v.id, t.timezone
FROM (VALUES
    ('New York Stock Exchange',    'NYSE',     'XNYS', 'United States',   'USD', 'America/New_York'),
    ('NASDAQ',                     'NASDAQ',   'XNAS', 'United States',   'USD', 'America/New_York'),
    ('NYSE American',              'NYSEMKT',  'XASE', 'United States',   'USD', 'America/New_York'),
    ('Chicago Board Options Exchange', 'CBOE', 'XCBO', 'United States',   'USD', 'America/Chicago'),
    ('London Stock Exchange',      'LSE',      'XLON', 'United Kingdom',  'GBP', 'Europe/London'),
    ('Deutsche Boerse XETRA',      'XETRA',    'XETR', 'Germany',         'EUR', 'Europe/Berlin'),
    ('Euronext Paris',             'ENX',      'XPAR', 'France',          'EUR', 'Europe/Paris'),
    ('Borsa Italiana',             'BIT',      'XMIL', 'Italy',           'EUR', 'Europe/Rome'),
    ('Euronext Amsterdam',         'AEX',      'XAMS', 'Netherlands',     'EUR', 'Europe/Amsterdam'),
    ('Euronext Brussels',          'ENXB',     'XBRU', 'Belgium',         'EUR', 'Europe/Brussels'),
    ('Bolsa de Madrid',            'BME',      'XMAD', 'Spain',           'EUR', 'Europe/Madrid'),
    ('Vienna Stock Exchange',      'VSE',      'XVIE', 'Austria',         'EUR', 'Europe/Vienna'),
    ('Euronext Lisbon',            'EURONEXT', 'XLIS', 'Portugal',        'EUR', 'Europe/Lisbon'),
    ('Irish Stock Exchange',       'ISE',      'XDUB', 'Ireland',         'EUR', 'Europe/Dublin'),
    ('Athens Exchange',            'ATHEX',    'ASEX', 'Greece',          'EUR', 'Europe/Athens'),
    ('Helsinki Stock Exchange',    'OMXH',     'XHEL', 'Finland',         'EUR', 'Europe/Helsinki'),
    ('Tokyo Stock Exchange',       'TSE',      'XTKS', 'Japan',           'JPY', 'Asia/Tokyo'),
    ('Osaka Exchange',             'OSE',      'XOSE', 'Japan',           'JPY', 'Asia/Tokyo'),
    ('Toronto Stock Exchange',     'TSX',      'XTSE', 'Canada',          'CAD', 'America/Toronto'),
    ('TSX Venture Exchange',       'TSXV',     'XTSX', 'Canada',          'CAD', 'America/Toronto'),
    ('Australian Securities Exchange', 'ASX',  'XASX', 'Australia',       'AUD', 'Australia/Sydney'),
    ('SIX Swiss Exchange',         'SIX',      'XSWX', 'Switzerland',     'CHF', 'Europe/Zurich'),
    ('Belgrade Stock Exchange',    'BELEX',    'XBEL', 'Serbia',          'RSD', 'Europe/Belgrade')
) AS t(name, acronym, mic_code, polity, currency_code, timezone)
JOIN core_banking.valuta v ON v.oznaka = t.currency_code
ON CONFLICT (mic_code) DO NOTHING;
