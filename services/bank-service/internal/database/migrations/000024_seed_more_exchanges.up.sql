-- =============================================================================
-- Migration: 000024_seed_more_exchanges
-- Service:   bank-service
-- Schema:    core_banking
--
-- 1. Dodaje valute koje nedostaju u core_banking.valuta
--    (IDR, INR, PLN, UAH, ARS, TRY, SEK, HUF, BGN)
-- 2. Insertuje ~91 novu berzu iz data/exchanges.csv
--    Berze cija valuta ne postoji u tabeli valuta ce biti preskocene.
--    ON CONFLICT (mic_code) DO NOTHING stiti od duplikata.
-- =============================================================================

-- ─── 1. Nedostajuce valute ───────────────────────────────────────────────────

-- status = FALSE: ove valute postoje samo radi FK veze sa berza tabelom.
-- Ne prikazuju se u UI-u (menjačnica, kreiranje računa itd.) jer GetAll filtrira po status=TRUE.
INSERT INTO core_banking.valuta (naziv, oznaka, simbol, zemlja, status)
VALUES
    ('Indonezijska rupija',  'IDR', 'Rp',  'Indonezija', FALSE),
    ('Indijska rupija',      'INR', '₹',   'Indija',     FALSE),
    ('Poljski zloti',        'PLN', 'zł',  'Poljska',    FALSE),
    ('Ukrajinska hrivnja',   'UAH', '₴',   'Ukraina',    FALSE),
    ('Argentinski pezos',    'ARS', '$',   'Argentina',  FALSE),
    ('Turska lira',          'TRY', '₺',   'Turska',     FALSE),
    ('Svedska kruna',        'SEK', 'kr',  'Svedska',    FALSE),
    ('Madarska forinta',     'HUF', 'Ft',  'Madarska',   FALSE),
    ('Bugarski lev',         'BGN', 'лв',  'Bugarska',   FALSE)
ON CONFLICT (oznaka) DO NOTHING;

-- ─── 2. Nove berze iz exchanges.csv ─────────────────────────────────────────

INSERT INTO core_banking.exchange (name, acronym, mic_code, polity, currency_id, timezone, open_time, close_time)
SELECT t.name, t.acronym, t.mic_code, t.polity, v.id, t.timezone, t.open_time::TIME, t.close_time::TIME
FROM (VALUES
    ('Jakarta Futures Exchange',                                         'BBJ',          'XBBJ', 'Indonesia',              'IDR', 'Asia/Jakarta',                    '09:00', '17:30'),
    ('Asx - Trade24',                                                    'SFE',          'XSFE', 'Australia',              'AUD', 'Australia/Melbourne',             '10:00', '16:00'),
    ('Cboe Edga U.s. Equities Exchange Dark',                            'EDGADARK',     'EDGD', 'United States',          'USD', 'America/New_York',                '09:30', '16:00'),
    ('Clear Street',                                                     'CLST',         'CLST', 'United States',          'USD', 'America/New_York',                '09:30', '16:00'),
    ('Wall Street Access Nyc',                                           'WABR',         'WABR', 'United States',          'USD', 'America/New_York',                '09:30', '16:00'),
    ('Marex Spectron Europe Limited - Otf',                              'MSEL OTF',     'MSEL', 'Ireland',                'EUR', 'Europe/Dublin',                   '08:00', '16:30'),
    ('Borsa Italiana Equity Mtf',                                        'BITEQMTF',     'MTAH', 'Italy',                  'EUR', 'Europe/Rome',                     '09:00', '17:25'),
    ('Clearcorp Dealing Systems India Limited - Astroid',                'ASTROID',      'ASTR', 'India',                  'INR', 'Asia/Kolkata',                    '09:15', '15:30'),
    ('Memx Llc Equities',                                                'MEMX',         'MEMX', 'United States',          'USD', 'America/New_York',                '09:30', '16:00'),
    ('Natixis - Systematic Internaliser',                                'NATX',         'NATX', 'France',                 'EUR', 'Europe/Paris',                    '09:00', '17:30'),
    ('Currenex Ireland Mtf - Rfq',                                       'CNX MTF',      'ICXR', 'Ireland',                'EUR', 'Europe/Dublin',                   '08:00', '16:30'),
    ('Neo Exchange - Neo-l',                                             'NEO-L',        'NEOE', 'Canada',                 'CAD', 'America/Montreal',                '09:30', '16:00'),
    ('Polish Trading Point',                                             'PTP',          'PTPG', 'Poland',                 'PLN', 'Europe/Warsaw',                   '09:00', '17:35'),
    ('Pfts Stock Exchange',                                              'PFTS',         'PFTS', 'Ukraine',                'UAH', 'Europe/Kiev',                     '10:00', '17:30'),
    ('Cboe Australia - Transferable Custody Receipt Market',             'CHI-X',        'CXAR', 'Australia',              'AUD', 'Australia/Melbourne',             '10:00', '16:00'),
    ('Essex Radez Llc',                                                  'GLPS',         'GLPS', 'United States',          'USD', 'America/New_York',                '09:30', '16:00'),
    ('London Metal Exchange',                                            'LME',          'XLME', 'United Kingdom',         'GBP', 'Europe/London',                   '08:00', '16:00'),
    ('Multi Commodity Exchange Of India Ltd.',                           'MCX',          'XIMC', 'India',                  'INR', 'Asia/Kolkata',                    '09:15', '15:30'),
    ('Cassa Di Compensazione E Garanzia - Ccp Agricultural Derivatives', 'CCGAGRIDER',   'CGGD', 'Italy',                  'EUR', 'Europe/Rome',                     '09:00', '17:25'),
    ('Toronto Stock Exchange - Drk',                                     'TSX DRK',      'XDRK', 'Canada',                 'CAD', 'America/Montreal',                '09:30', '16:00'),
    ('Chicago Mercantile Exchange',                                      'CME',          'XCME', 'United States',          'USD', 'America/New_York',                '09:30', '16:00'),
    ('Neo Connect',                                                      'NEO CONNECT',  'NEOC', 'Canada',                 'CAD', 'America/Montreal',                '09:30', '16:00'),
    ('Ubs Ats',                                                          'UBS ATS',      'UBSA', 'United States',          'USD', 'America/New_York',                '09:30', '16:00'),
    ('Athens Exchange - Apa',                                            'ATHEX APA',    'AAPA', 'Greece',                 'EUR', 'Europe/Athens',                   '10:15', '17:20'),
    ('Bny Mellon - Systematic Internaliser',                             'BNYM',         'BKLF', 'United Kingdom',         'GBP', 'Europe/London',                   '08:00', '16:00'),
    ('Bank Of Montreal - London Branch - Systematic Internaliser',       'BMO',          'BMLB', 'United Kingdom',         'GBP', 'Europe/London',                   '08:00', '16:00'),
    ('Financial Information Contributors Exchange',                      'FICONEX',      'FICX', 'Germany',                'EUR', 'Europe/Berlin',                   '08:00', '20:00'),
    ('Abn Amro Clearing Bank - Systematic Internaliser',                 'AACB SI',      'ABNC', 'Netherlands',            'EUR', 'Europe/Amsterdam',                '09:00', '17:40'),
    ('Credit Industriel Et Commercial - Systematic Internaliser',        'CIC',          'CMCI', 'France',                 'EUR', 'Europe/Paris',                    '09:00', '17:30'),
    ('Posit Rfq',                                                        'RFQ',          'XRFQ', 'Ireland',                'EUR', 'Europe/Dublin',                   '08:00', '16:30'),
    ('Aquis Exchange Plc',                                               'AQX',          'AQXE', 'United Kingdom',         'GBP', 'Europe/London',                   '08:00', '16:00'),
    ('Financial And Risk Transactions Services Ireland - Fxall Rfs Mtf', 'FRTSIL',       'FXRS', 'Ireland',                'EUR', 'Europe/Dublin',                   '08:00', '16:30'),
    ('Berenberg Fixed Income Uk - Systematic Internaliser',              'BGFU',         'BGFU', 'United Kingdom',         'GBP', 'Europe/London',                   '08:00', '16:00'),
    ('National Bank Financial Inc. - Systematic Internaliser',           'NBF',          'NBFL', 'United Kingdom',         'GBP', 'Europe/London',                   '08:00', '16:00'),
    ('Jpbx',                                                             'JPBX',         'JPBX', 'United States',          'USD', 'America/New_York',                '09:30', '16:00'),
    ('Miax Pearl Llc',                                                   'MPRL',         'MPRL', 'United States',          'USD', 'America/New_York',                '09:30', '16:00'),
    ('Canadian Securities Exchange',                                     'CSE LISTED',   'XCNQ', 'Canada',                 'CAD', 'America/Montreal',                '09:30', '16:00'),
    ('Two Sigma Securities Llc',                                         'SOHO',         'SOHO', 'United States',          'USD', 'America/New_York',                '09:30', '16:00'),
    ('Currenex Ldfx',                                                    'CX LDFX',      'LCUR', 'United Kingdom',         'GBP', 'Europe/London',                   '08:00', '16:00'),
    ('Puma Capital Llc - Options',                                       'PUMA',         'PUMX', 'United States',          'USD', 'America/New_York',                '09:30', '16:00'),
    ('Virtual Auction Global Markets - Mtf',                             'VAGM MTF',     'VAGM', 'United Kingdom',         'GBP', 'Europe/London',                   '08:00', '16:00'),
    ('Miami International Holdings Inc.',                                'MIHI',         'MIHI', 'United States',          'USD', 'America/New_York',                '09:30', '16:00'),
    ('Cboe Europe - Lis Service',                                        'CBOE LIS',     'LISX', 'United Kingdom',         'GBP', 'Europe/London',                   '08:00', '16:00'),
    ('Global Securities Exchange',                                       'GSX',          'XGSX', 'United Kingdom',         'GBP', 'Europe/London',                   '08:00', '16:00'),
    ('Tw Sef Llc',                                                       'TWSEF',        'TWSF', 'United States',          'USD', 'America/New_York',                '09:30', '16:00'),
    ('Gtx Sef Llc',                                                      'GTX',          'GTXS', 'United States',          'USD', 'America/New_York',                '09:30', '16:00'),
    ('Freight Investor Services Limited',                                'FIS',          'FISU', 'United Kingdom',         'GBP', 'Europe/London',                   '08:00', '16:00'),
    ('Mercado De Valores De Buenos Aires S.a.',                          'MERVAL',       'XMEV', 'Argentina',              'ARS', 'America/Argentina/Buenos_Aires',  '11:00', '17:00'),
    ('Electricity Day-ahead Market',                                     'EXIST',        'XEDA', 'Turkey',                 'TRY', 'Asia/Istanbul',                   '09:00', '17:30'),
    ('Emerging Markets Bond Exchange Limited',                           'EMBX',         'EMBX', 'United Kingdom',         'GBP', 'Europe/London',                   '08:00', '16:00'),
    ('Six Swiss Bilateral Trading Platform For Structured Otc Products', 'SIX',          'XBTR', 'Switzerland',            'CHF', 'Europe/Zurich',                   '09:00', '17:30'),
    ('Minneapolis Grain Exchange',                                       'MGE',          'XMGE', 'United States',          'USD', 'America/New_York',                '09:30', '16:00'),
    ('Svenska Handelsbanken Ab - Svex',                                  'SVEX',         'SVEX', 'Sweden',                 'SEK', 'Europe/Oslo',                     '09:00', '16:30'),
    ('Global Derivatives Exchange',                                      'GDX',          'XGDX', 'United Kingdom',         'GBP', 'Europe/London',                   '08:00', '16:00'),
    ('Bnp Paribas Securities Services London Branch - Si',               'BP2S LB SI',   'BSPL', 'United Kingdom',         'GBP', 'Europe/London',                   '08:00', '16:00'),
    ('Cboe Europe Equities - European Equities (nl)',                    'CBOE EUROPE',  'CCXE', 'Netherlands',            'EUR', 'Europe/Amsterdam',                '09:00', '17:40'),
    ('Bnp Paribas Sa - Systematic Internaliser',                         'BNPP SA SI',   'BNPS', 'France',                 'EUR', 'Europe/Paris',                    '09:00', '17:30'),
    ('Tradegate Exchange - Systematic Internaliser',                     'TGAG',         'TGSI', 'Germany',                'EUR', 'Europe/Berlin',                   '08:00', '20:00'),
    ('Exane Bnp Paribas',                                                'EXEU',         'EXEU', 'United Kingdom',         'GBP', 'Europe/London',                   '08:00', '16:00'),
    ('Liquidnet H20',                                                    'LQNT H20',     'LIQH', 'United Kingdom',         'GBP', 'Europe/London',                   '08:00', '16:00'),
    ('New York Portfolio Clearing',                                      'NYPC',         'NYPC', 'United States',          'USD', 'America/New_York',                '09:30', '16:00'),
    ('Gemma (gilt Edged Market Makers Association)',                      'GEMMA',        'GEMX', 'United Kingdom',         'GBP', 'Europe/London',                   '08:00', '16:00'),
    ('Hudson River Trading (hrt)',                                       'HRT',          'HRTF', 'United States',          'USD', 'America/New_York',                '09:30', '16:00'),
    ('Santander Uk - Systematic Internaliser',                           'SNUK',         'SNUK', 'United Kingdom',         'GBP', 'Europe/London',                   '08:00', '16:00'),
    ('Aquis Exchange Plc - Eix Infrastructure Bond Market',              'AQUIS-EIX',    'EIXE', 'United Kingdom',         'GBP', 'Europe/London',                   '08:00', '16:00'),
    ('Henex S.a.',                                                       'HENEX',        'HEMO', 'Greece',                 'EUR', 'Europe/Athens',                   '10:15', '17:20'),
    ('Jane Street Financial Ltd - Systematic Internaliser',              'JSF',          'JSSI', 'United Kingdom',         'GBP', 'Europe/London',                   '08:00', '16:00'),
    ('Securitised Derivatives Market',                                   'SEDEX',        'SEDX', 'Italy',                  'EUR', 'Europe/Rome',                     '09:00', '17:25'),
    ('Rivercross Dark',                                                  'RIVERX',       'RICD', 'United States',          'USD', 'America/New_York',                '09:30', '16:00'),
    ('Bank Polska Kasa Opieki S.a. - Systematic Internaliser',           'PEKAO',        'PKOP', 'Poland',                 'PLN', 'Europe/Warsaw',                   '09:00', '17:35'),
    ('Xtend',                                                            'XTND',         'XTND', 'Hungary',                'HUF', 'Europe/Budapest',                 '09:00', '17:00'),
    ('The Green Stock Exchange - Acb Impact Markets',                    'GSE',          'GRSE', 'United Kingdom',         'GBP', 'Europe/London',                   '08:00', '16:00'),
    ('Susquehanna International Securities Limited - Si',                'SISI',         'SISI', 'Ireland',                'EUR', 'Europe/Dublin',                   '08:00', '16:30'),
    ('Banque Et Caisse D epargne De L etat Luxembourg - Si',             'BCEE',         'BCEE', 'Luxembourg',             'EUR', 'Europe/Luxembourg',               '09:00', '17:35'),
    ('Merrill Lynch International - Rfq - Systematic Internaliser',      'MLI',          'MLRQ', 'United Kingdom',         'GBP', 'Europe/London',                   '08:00', '16:00'),
    ('Westpac Banking Corporation - Systematic Internaliser',            'WBC SI',       'WSIN', 'United Kingdom',         'GBP', 'Europe/London',                   '08:00', '16:00'),
    ('Bulgarian Stock Exchange - Alternative Market',                    'BSE',          'ABUL', 'Bulgaria',               'BGN', 'Europe/Sofia',                    '10:10', '16:55'),
    ('Gfi Securities Llc - Creditmatch (latg)',                          'LATG',         'LATG', 'United States',          'USD', 'America/New_York',                '09:30', '16:00'),
    ('Italian Derivatives Market',                                       'IDEM',         'XDMI', 'Italy',                  'EUR', 'Europe/Rome',                     '09:00', '17:25'),
    ('Hudson River Trading',                                             'HRT',          'HRTX', 'United States',          'USD', 'America/New_York',                '09:30', '16:00'),
    ('Nex Sef Mtf - Reset - Risk Mitigation Services',                   'NSL',          'REST', 'United Kingdom',         'GBP', 'Europe/London',                   '08:00', '16:00'),
    ('Banque Internationale A Luxembourg S.a. - Si',                     'BIL',          'BILU', 'Luxembourg',             'EUR', 'Europe/Luxembourg',               '09:00', '17:35'),
    ('Liquidnet Inc.',                                                   'LQNT',         'LIUS', 'United States',          'USD', 'America/New_York',                '09:30', '16:00'),
    ('Bnp Paribas Securities Services - Systematic Internaliser',        'BP2S',         'BPSX', 'France',                 'EUR', 'Europe/Paris',                    '09:00', '17:30'),
    ('Ice Endex European Gas Spot',                                      'ICE ENDEX',    'NDXS', 'Netherlands',            'EUR', 'Europe/Amsterdam',                '09:00', '17:40'),
    ('Memx Llc Dark',                                                    'MEMXDARK',     'MEMD', 'United States',          'USD', 'America/New_York',                '09:30', '16:00'),
    ('Societe Generale - Systematic Internaliser',                       'SG SI',        'XSGA', 'France',                 'EUR', 'Europe/Paris',                    '09:00', '17:30'),
    ('Level Ats - Vwap Cross',                                           'LEVEL',        'EBXV', 'United States',          'USD', 'America/New_York',                '09:30', '16:00'),
    ('Wells Fargo Securities Europe S.a.',                               'WFSE',         'WFSE', 'France',                 'EUR', 'Europe/Paris',                    '09:00', '17:30'),
    ('Bnp Paribas Sa London Branch - Systematic Internaliser',           'BNPP SA LB SI','BNPL', 'United Kingdom',         'GBP', 'Europe/London',                   '08:00', '16:00'),
    ('Warsaw Stock Exchange/indices',                                    'GPWB',         'WIND', 'Poland',                 'PLN', 'Europe/Warsaw',                   '09:00', '17:35')
) AS t(name, acronym, mic_code, polity, currency_code, timezone, open_time, close_time)
JOIN core_banking.valuta v ON v.oznaka = t.currency_code
ON CONFLICT (mic_code) DO NOTHING;
