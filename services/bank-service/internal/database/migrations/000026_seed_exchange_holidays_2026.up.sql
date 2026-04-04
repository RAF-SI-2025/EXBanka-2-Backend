-- =============================================================================
-- Migration: 000026_seed_exchange_holidays_2026
-- Service:   bank-service
-- Schema:    core_banking
--
-- Puni tabelu exchange_holiday zvaničnim neradnim danima berzi za 2026. godinu.
-- Grupisano po polity vrednostima (tačno kao u tabeli core_banking.exchange).
--
-- Metoda:
--   Datumi vikenda (subota/nedelja) su izostavljeni jer berze ionako ne trguju.
--   Kada zvanični praznik pada na vikend, unet je surogatni radni dan (zamena)
--   gde to propisuje lokalni zakon.
--
-- Napomene o tačnosti:
--   • Islamski praznici (Indonesia, Turkey) — izračunati astronomski prema
--     lunarnom kalendaru; mogu varirati ±1 dan zavisno od zvaničnog oglašavanja
--     mesečevog mlađaka.
--   • Hindu praznici (India) — aproksimativni; BSE/NSE godišnje objavljuju
--     tačan raspored.
--   • Japanski Jesenski ekvinocij (Sep 23) — potvrđen astronomski za 2026.
--
-- Datum izrade: 2026-04-04
-- =============================================================================

INSERT INTO core_banking.exchange_holiday (polity, date) VALUES

-- ─── United States ────────────────────────────────────────────────────────────
-- NYSE, NASDAQ, NYSE American, CBOE, CME i ostale US berze
    ('United States', '2026-01-01'),  -- New Year's Day
    ('United States', '2026-01-19'),  -- Martin Luther King Jr. Day  (3. ponedeljak januara)
    ('United States', '2026-02-16'),  -- Presidents' Day             (3. ponedeljak februara)
    ('United States', '2026-04-03'),  -- Good Friday
    ('United States', '2026-05-25'),  -- Memorial Day                (poslednji ponedeljak maja)
    ('United States', '2026-06-19'),  -- Juneteenth National Independence Day
    ('United States', '2026-07-03'),  -- Independence Day — observed (4. jul pada u subotu → petak)
    ('United States', '2026-09-07'),  -- Labor Day                   (1. ponedeljak septembra)
    ('United States', '2026-11-26'),  -- Thanksgiving Day            (4. četvrtak novembra)
    ('United States', '2026-12-25'),  -- Christmas Day

-- ─── United Kingdom ───────────────────────────────────────────────────────────
-- LSE, LME i ostale UK berze
    ('United Kingdom', '2026-01-01'),  -- New Year's Day
    ('United Kingdom', '2026-04-03'),  -- Good Friday
    ('United Kingdom', '2026-04-06'),  -- Easter Monday
    ('United Kingdom', '2026-05-04'),  -- Early May Bank Holiday      (1. ponedeljak maja)
    ('United Kingdom', '2026-05-25'),  -- Spring Bank Holiday         (poslednji ponedeljak maja)
    ('United Kingdom', '2026-08-31'),  -- Summer Bank Holiday         (poslednji ponedeljak avgusta)
    ('United Kingdom', '2026-12-25'),  -- Christmas Day
    ('United Kingdom', '2026-12-28'),  -- Boxing Day — observed       (26.12. pada u subotu → pon)

-- ─── Germany ──────────────────────────────────────────────────────────────────
-- Deutsche Boerse XETRA, Frankfurt Stock Exchange, Tradegate
    ('Germany', '2026-01-01'),  -- Neujahrstag
    ('Germany', '2026-04-03'),  -- Karfreitag (Good Friday)
    ('Germany', '2026-04-06'),  -- Ostermontag (Easter Monday)
    ('Germany', '2026-05-01'),  -- Tag der Arbeit (Labour Day)
    ('Germany', '2026-12-24'),  -- Heiligabend (Christmas Eve — Xetra zatvoren ceo dan)
    ('Germany', '2026-12-25'),  -- 1. Weihnachtstag (Christmas Day)

-- ─── France ───────────────────────────────────────────────────────────────────
-- Euronext Paris (standardni Euronext kalendar)
    ('France', '2026-01-01'),  -- Jour de l'An
    ('France', '2026-04-03'),  -- Vendredi Saint (Good Friday)
    ('France', '2026-04-06'),  -- Lundi de Pâques (Easter Monday)
    ('France', '2026-05-01'),  -- Fête du Travail (Labour Day)
    ('France', '2026-12-25'),  -- Noël (Christmas Day)

-- ─── Italy ────────────────────────────────────────────────────────────────────
-- Borsa Italiana (Euronext Milan)
    ('Italy', '2026-01-01'),  -- Capodanno
    ('Italy', '2026-04-03'),  -- Venerdì Santo (Good Friday)
    ('Italy', '2026-04-06'),  -- Lunedì dell'Angelo (Easter Monday)
    ('Italy', '2026-05-01'),  -- Festa del Lavoro
    ('Italy', '2026-12-24'),  -- Vigilia di Natale (Christmas Eve — Borsa Italiana zatvorena)
    ('Italy', '2026-12-25'),  -- Natale (Christmas Day)

-- ─── Netherlands ──────────────────────────────────────────────────────────────
-- Euronext Amsterdam (standardni Euronext kalendar)
    ('Netherlands', '2026-01-01'),  -- Nieuwjaarsdag
    ('Netherlands', '2026-04-03'),  -- Goede Vrijdag (Good Friday)
    ('Netherlands', '2026-04-06'),  -- Tweede Paasdag (Easter Monday)
    ('Netherlands', '2026-05-01'),  -- Dag van de Arbeid (Labour Day)
    ('Netherlands', '2026-12-25'),  -- Eerste Kerstdag (Christmas Day)

-- ─── Belgium ──────────────────────────────────────────────────────────────────
-- Euronext Brussels (standardni Euronext kalendar)
    ('Belgium', '2026-01-01'),  -- Nieuwjaar
    ('Belgium', '2026-04-03'),  -- Goede Vrijdag (Good Friday)
    ('Belgium', '2026-04-06'),  -- Paasmaandag (Easter Monday)
    ('Belgium', '2026-05-01'),  -- Dag van de Arbeid (Labour Day)
    ('Belgium', '2026-12-25'),  -- Kerstdag (Christmas Day)

-- ─── Spain ────────────────────────────────────────────────────────────────────
-- Bolsa de Madrid (BME / Bolsas y Mercados Españoles)
    ('Spain', '2026-01-01'),  -- Año Nuevo
    ('Spain', '2026-01-06'),  -- Epifanía del Señor (Reyes Magos)
    ('Spain', '2026-04-03'),  -- Viernes Santo (Good Friday)
    ('Spain', '2026-05-01'),  -- Día del Trabajo (Labour Day)
    ('Spain', '2026-10-12'),  -- Día de la Hispanidad (Fiesta Nacional de España)
    ('Spain', '2026-12-08'),  -- Inmaculada Concepción
    ('Spain', '2026-12-25'),  -- Navidad (Christmas Day)

-- ─── Austria ──────────────────────────────────────────────────────────────────
-- Wiener Börse (Vienna Stock Exchange)
-- Napomena: Karfreitag (Good Friday) NIJE austrijski državni praznik od 2019.
    ('Austria', '2026-01-01'),  -- Neujahr
    ('Austria', '2026-01-06'),  -- Heilige Drei Könige (Epiphany)
    ('Austria', '2026-04-06'),  -- Ostermontag (Easter Monday)
    ('Austria', '2026-05-01'),  -- Staatsfeiertag (Labour Day)
    ('Austria', '2026-05-14'),  -- Christi Himmelfahrt (Ascension — 39 dana posle Uskrsa)
    ('Austria', '2026-05-25'),  -- Pfingstmontag (Whit Monday — 50 dana posle Uskrsa)
    ('Austria', '2026-06-04'),  -- Fronleichnam (Corpus Christi — 60 dana posle Uskrsa)
    ('Austria', '2026-10-26'),  -- Nationalfeiertag (Austrian National Day)
    ('Austria', '2026-12-08'),  -- Mariä Empfängnis (Immaculate Conception)
    ('Austria', '2026-12-25'),  -- Christtag (Christmas Day)

-- ─── Portugal ─────────────────────────────────────────────────────────────────
-- Euronext Lisbon (standardni Euronext kalendar)
    ('Portugal', '2026-01-01'),  -- Ano Novo
    ('Portugal', '2026-04-03'),  -- Sexta-Feira Santa (Good Friday)
    ('Portugal', '2026-04-06'),  -- Segunda-Feira de Páscoa (Easter Monday)
    ('Portugal', '2026-05-01'),  -- Dia do Trabalhador (Labour Day)
    ('Portugal', '2026-12-25'),  -- Natal (Christmas Day)

-- ─── Ireland ──────────────────────────────────────────────────────────────────
-- Euronext Dublin (Irish Stock Exchange) + irski državni praznici
    ('Ireland', '2026-01-01'),  -- New Year's Day
    ('Ireland', '2026-03-17'),  -- St. Patrick's Day
    ('Ireland', '2026-04-03'),  -- Good Friday            (Euronext Dublin zatvoren)
    ('Ireland', '2026-04-06'),  -- Easter Monday
    ('Ireland', '2026-05-04'),  -- May Bank Holiday       (1. ponedeljak maja)
    ('Ireland', '2026-06-01'),  -- June Bank Holiday      (1. ponedeljak juna — jun 1 = pon)
    ('Ireland', '2026-08-03'),  -- August Bank Holiday    (1. ponedeljak avgusta)
    ('Ireland', '2026-10-26'),  -- October Bank Holiday   (poslednji ponedeljak oktobra)
    ('Ireland', '2026-12-25'),  -- Christmas Day
    ('Ireland', '2026-12-28'),  -- St. Stephen's Day — observed (26.12. pada u subotu → pon)

-- ─── Greece ───────────────────────────────────────────────────────────────────
-- Athens Exchange (ATHEX) — praznici po Pravoslavnom kalendaru
-- Pravoslavni Uskrs 2026 = 19. april
    ('Greece', '2026-01-01'),  -- Πρωτοχρονιά (New Year's Day)
    ('Greece', '2026-01-06'),  -- Θεοφάνεια (Epiphany)
    ('Greece', '2026-03-02'),  -- Καθαρά Δευτέρα (Clean Monday — 48 dana pre Prv. Uskrsa)
    ('Greece', '2026-03-25'),  -- Εθνική Εορτή (Greek Independence Day)
    ('Greece', '2026-04-17'),  -- Μεγάλη Παρασκευή (Orthodox Good Friday)
    ('Greece', '2026-04-20'),  -- Δευτέρα του Πάσχα (Orthodox Easter Monday)
    ('Greece', '2026-05-01'),  -- Εργατική Πρωτομαγιά (Labour Day)
    ('Greece', '2026-06-08'),  -- Δευτέρα Αγίου Πνεύματος (Orthodox Whit Monday — 50 dana)
    ('Greece', '2026-10-28'),  -- Ημέρα του Όχι (Ohi Day)
    ('Greece', '2026-12-25'),  -- Χριστούγεννα (Christmas Day)

-- ─── Finland ──────────────────────────────────────────────────────────────────
-- Nasdaq Helsinki (OMX Helsinki)
    ('Finland', '2026-01-01'),  -- Uudenvuodenpäivä (New Year's Day)
    ('Finland', '2026-01-06'),  -- Loppiainen (Epiphany)
    ('Finland', '2026-04-03'),  -- Pitkäperjantai (Good Friday)
    ('Finland', '2026-04-06'),  -- Toinen pääsiäispäivä (Easter Monday)
    ('Finland', '2026-05-01'),  -- Vappu (May Day)
    ('Finland', '2026-05-14'),  -- Helatorstai (Ascension Day — 39 dana posle Uskrsa)
    ('Finland', '2026-06-19'),  -- Juhannusaatto (Midsummer Eve — petak između 19–25.6.)
    ('Finland', '2026-12-24'),  -- Jouluaatto (Christmas Eve)
    ('Finland', '2026-12-25'),  -- Joulupäivä (Christmas Day)

-- ─── Japan ────────────────────────────────────────────────────────────────────
-- Tokyo Stock Exchange (TSE), Osaka Exchange (OSE)
-- Napomena: Sep 22 postaje državni praznik jer je okružen sa oba strane praznicima
--           (Sep 21 + Sep 23), po japanskom zakonu o "sandvič" prazniku.
    ('Japan', '2026-01-01'),  -- 元日 Ganjitsu (New Year's Day)
    ('Japan', '2026-01-02'),  -- お正月 Oshōgatsu (TSE zatvoren 2.1. — tržišna konvencija)
    ('Japan', '2026-01-12'),  -- 成人の日 Seijin no Hi (Coming of Age Day — 2. ponedeljak jan)
    ('Japan', '2026-02-11'),  -- 建国記念の日 Kenkoku Kinen no Hi (National Foundation Day)
    ('Japan', '2026-02-23'),  -- 天皇誕生日 Tennō Tanjōbi (Emperor's Birthday)
    ('Japan', '2026-03-20'),  -- 春分の日 Shunbun no Hi (Vernal Equinox Day)
    ('Japan', '2026-04-29'),  -- 昭和の日 Shōwa no Hi (Showa Day)
    ('Japan', '2026-05-04'),  -- みどりの日 Midori no Hi (Greenery Day)
    ('Japan', '2026-05-05'),  -- こどもの日 Kodomo no Hi (Children's Day)
    ('Japan', '2026-05-06'),  -- 振替休日 Furikae (zamena za Ustav. dan 3.5. = nedelja)
    ('Japan', '2026-07-20'),  -- 海の日 Umi no Hi (Marine Day — 3. ponedeljak jula)
    ('Japan', '2026-08-11'),  -- 山の日 Yama no Hi (Mountain Day)
    ('Japan', '2026-09-21'),  -- 敬老の日 Keirō no Hi (Respect for the Aged — 3. pon. sep)
    ('Japan', '2026-09-22'),  -- 国民の休日 Kokumin no Kyūjitsu (sandvič između Sep 21 i 23)
    ('Japan', '2026-09-23'),  -- 秋分の日 Shūbun no Hi (Autumnal Equinox Day)
    ('Japan', '2026-10-12'),  -- スポーツの日 Supōtsu no Hi (Health and Sports — 2. pon. okt)
    ('Japan', '2026-11-03'),  -- 文化の日 Bunka no Hi (Culture Day)
    ('Japan', '2026-11-23'),  -- 勤労感謝の日 Kinrō Kansha no Hi (Labour Thanksgiving Day)
    ('Japan', '2026-12-31'),  -- 大納会 Ō-Nōkai (Year-End Market Close — TSE konvencija)

-- ─── Canada ───────────────────────────────────────────────────────────────────
-- Toronto Stock Exchange (TSX), TSX Venture Exchange
    ('Canada', '2026-01-01'),  -- New Year's Day
    ('Canada', '2026-02-16'),  -- Family Day                (3. ponedeljak februara — Ontario)
    ('Canada', '2026-04-03'),  -- Good Friday
    ('Canada', '2026-05-18'),  -- Victoria Day              (poslednji ponedeljak pre 25.5.)
    ('Canada', '2026-07-01'),  -- Canada Day
    ('Canada', '2026-08-03'),  -- Civic Holiday             (1. ponedeljak avgusta — Ontario)
    ('Canada', '2026-09-07'),  -- Labour Day                (1. ponedeljak septembra)
    ('Canada', '2026-10-12'),  -- Thanksgiving Day          (2. ponedeljak oktobra)
    ('Canada', '2026-12-25'),  -- Christmas Day
    ('Canada', '2026-12-28'),  -- Boxing Day — observed     (26.12. pada u subotu → pon)

-- ─── Australia ────────────────────────────────────────────────────────────────
-- ASX (Australian Securities Exchange)
-- Napomena: ANZAC Day (25.4.) pada u subotu — nema surogatnog praznika za berzu
    ('Australia', '2026-01-01'),  -- New Year's Day
    ('Australia', '2026-01-26'),  -- Australia Day
    ('Australia', '2026-04-03'),  -- Good Friday
    ('Australia', '2026-04-06'),  -- Easter Monday
    ('Australia', '2026-06-08'),  -- King's Birthday          (2. ponedeljak juna — NSW/ACT)
    ('Australia', '2026-12-25'),  -- Christmas Day
    ('Australia', '2026-12-28'),  -- Boxing Day — observed    (26.12. pada u subotu → pon)

-- ─── Switzerland ──────────────────────────────────────────────────────────────
-- SIX Swiss Exchange
-- Napomena: Nacionalni dan (1. avg.) pada u subotu — bez surogatnog praznika
    ('Switzerland', '2026-01-01'),  -- Neujahrstag (New Year's Day)
    ('Switzerland', '2026-01-02'),  -- Berchtoldstag
    ('Switzerland', '2026-04-03'),  -- Karfreitag (Good Friday)
    ('Switzerland', '2026-04-06'),  -- Ostermontag (Easter Monday)
    ('Switzerland', '2026-05-01'),  -- Tag der Arbeit (Labour Day — kanton Zürich)
    ('Switzerland', '2026-05-14'),  -- Auffahrt (Ascension Day)
    ('Switzerland', '2026-05-25'),  -- Pfingstmontag (Whit Monday)
    ('Switzerland', '2026-12-24'),  -- Heiligabend (Christmas Eve — SIX zatvoren ceo dan)
    ('Switzerland', '2026-12-25'),  -- Weihnachtstag (Christmas Day)
    ('Switzerland', '2026-12-31'),  -- Silvester (New Year's Eve — SIX zatvoren)

-- ─── Serbia ───────────────────────────────────────────────────────────────────
-- Beogradska Berza (BELEX)
-- Pravoslavni Uskrs 2026 = 19. april
    ('Serbia', '2026-01-01'),  -- Nova Godina (prvi dan)
    ('Serbia', '2026-01-02'),  -- Nova Godina (drugi dan)
    ('Serbia', '2026-01-07'),  -- Božić — Pravoslavni (7. januar po Jul. kalendaru)
    ('Serbia', '2026-01-08'),  -- Božić (drugi dan)
    ('Serbia', '2026-02-16'),  -- Dan državnosti — zamena (15.2. = nedelja → ponedeljak)
    ('Serbia', '2026-04-17'),  -- Veliki petak (Pravoslavni Good Friday)
    ('Serbia', '2026-04-20'),  -- Uskrs — ponedeljak (Pravoslavni Easter Monday)
    ('Serbia', '2026-05-01'),  -- Praznik rada (prvi dan)
    ('Serbia', '2026-05-04'),  -- Praznik rada — zamena (2.5. = subota → ponedeljak)
    ('Serbia', '2026-11-11'),  -- Dan primirja u Prvom svetskom ratu

-- ─── Indonesia ────────────────────────────────────────────────────────────────
-- Jakarta Futures Exchange (BBJ)
-- Islamski praznici: proračun prema astronomskom lunarnom kalendaru 1447/1448H
--   Ramazan 1447H počinje ~18. feb → Idul Fitri ~20. mar
--   Idul Adha 1447H (10 Dzulhijjah) ~27. maj
--   Tahun Baru Islam 1448H (1 Muharram) ~16. jun
--   Maulid Nabi 1448H (12 Rabi'ul Awal) ~25. avg
    ('Indonesia', '2026-01-01'),  -- Tahun Baru Masehi (New Year's Day)
    ('Indonesia', '2026-01-15'),  -- Isra Mi'raj Nabi Muhammad SAW (~27 Rajab 1447H)
    ('Indonesia', '2026-02-17'),  -- Tahun Baru Imlek 2577 (Chinese New Year — Tahun Kuda)
    ('Indonesia', '2026-03-20'),  -- Hari Raya Idul Fitri 1447H (~1 Syawal)
    ('Indonesia', '2026-04-03'),  -- Jumat Agung (Good Friday)
    ('Indonesia', '2026-05-01'),  -- Hari Buruh Internasional (Labour Day)
    ('Indonesia', '2026-05-14'),  -- Kenaikan Isa Almasih (Ascension Day)
    ('Indonesia', '2026-05-27'),  -- Hari Raya Idul Adha 1447H (~10 Dzulhijjah)
    ('Indonesia', '2026-06-01'),  -- Hari Lahir Pancasila (Pancasila Day)
    ('Indonesia', '2026-06-16'),  -- Tahun Baru Islam 1448H (~1 Muharram)
    ('Indonesia', '2026-08-17'),  -- Hari Kemerdekaan RI (Independence Day)
    ('Indonesia', '2026-08-25'),  -- Maulid Nabi Muhammad SAW (~12 Rabi'ul Awal 1448H)
    ('Indonesia', '2026-12-25'),  -- Hari Natal (Christmas Day)

-- ─── India ────────────────────────────────────────────────────────────────────
-- MCX (Multi Commodity Exchange), Clearcorp ASTROID
-- Hindu praznici su aproksimativni; BSE/NSE svake godine objavljuju tačan raspored
    ('India', '2026-01-26'),  -- Republic Day (Dan Republike)
    ('India', '2026-03-03'),  -- Holi / Dhuleti (Phalguna Purnima — aproksimativno)
    ('India', '2026-04-03'),  -- Good Friday
    ('India', '2026-04-14'),  -- Dr. Ambedkar Jayanti
    ('India', '2026-05-01'),  -- Maharashtra Day / Labour Day
    ('India', '2026-10-02'),  -- Gandhi Jayanti (Mahatma Gandhi's Birthday)
    ('India', '2026-11-24'),  -- Guru Nanak Jayanti (~Kartik Purnima — aproksimativno)
    ('India', '2026-12-25'),  -- Christmas Day

-- ─── Ukraine ──────────────────────────────────────────────────────────────────
-- PFTS Stock Exchange (Kijevska berza)
-- Pravoslavni Uskrs 2026 = 19. april; Božić = 25. decembar (od 2023)
    ('Ukraine', '2026-01-01'),  -- Новий рік (New Year's Day)
    ('Ukraine', '2026-03-09'),  -- Міжнародний жіночий день — замена (8.3. = нед → пон)
    ('Ukraine', '2026-04-20'),  -- Великдень (Orthodox Easter Monday)
    ('Ukraine', '2026-05-01'),  -- День праці (Labour Day)
    ('Ukraine', '2026-06-29'),  -- День Конституції — замена (28.6. = нед → пон)
    ('Ukraine', '2026-08-24'),  -- День Незалежності України (Independence Day)
    ('Ukraine', '2026-10-14'),  -- День захисників і захисниць України (Defenders Day)
    ('Ukraine', '2026-12-25'),  -- Різдво Христове (Christmas Day — od 2023 po Greg. kal.)

-- ─── Argentina ────────────────────────────────────────────────────────────────
-- MERVAL (Mercado de Valores de Buenos Aires S.A.)
-- "Trasladables" — pomični praznici koji padaju na radni dan blizak utorku/sredu
--   idu na prethodni ponedeljak; četvrtak/petak idu na sledeći ponedeljak.
-- Nov 20 (Suverenost) pada u petak → pomerio se na ponedeljak 23. nov.
    ('Argentina', '2026-01-01'),  -- Año Nuevo
    ('Argentina', '2026-02-16'),  -- Lunes de Carnaval    (47 dana pre Uskrsa)
    ('Argentina', '2026-02-17'),  -- Martes de Carnaval
    ('Argentina', '2026-03-24'),  -- Día Nacional de la Memoria por la Verdad y la Justicia
    ('Argentina', '2026-04-02'),  -- Día del Veterano y de los Caídos en la Guerra de Malvinas
    ('Argentina', '2026-04-03'),  -- Viernes Santo (Good Friday)
    ('Argentina', '2026-05-25'),  -- Día de la Revolución de Mayo
    ('Argentina', '2026-07-09'),  -- Día de la Independencia
    ('Argentina', '2026-08-17'),  -- Paso a la Inmortalidad del Gral. San Martín (trasladable)
    ('Argentina', '2026-10-12'),  -- Día de la Diversidad Cultural (trasladable)
    ('Argentina', '2026-11-23'),  -- Día de la Soberanía Nacional (trasladable — 20.11. = pet → 23.11. = pon)
    ('Argentina', '2026-12-08'),  -- Inmaculada Concepción de María
    ('Argentina', '2026-12-25'),  -- Navidad (Christmas Day)

-- ─── Turkey ───────────────────────────────────────────────────────────────────
-- EXIST (Electricity Day-ahead Market / Elektrik Piyasaları)
-- Islamski praznici: Ramazan Bayramı (~20. mar), Kurban Bayramı (~27. maj)
-- Borsa Istanbul (BIST) zatvara berzu i za arife (dan pre Bajrama).
-- Zafer Bayramı (30.8.) pada u nedelju → zamena 31.8. (ponedeljak).
    ('Turkey', '2026-01-01'),  -- Yılbaşı (New Year's Day)
    ('Turkey', '2026-03-19'),  -- Ramazan Bayramı Arifesi (~arife, dan pre Eid ul-Fitri)
    ('Turkey', '2026-03-20'),  -- Ramazan Bayramı 1. Günü (Eid ul-Fitr)
    ('Turkey', '2026-04-23'),  -- Ulusal Egemenlik ve Çocuk Bayramı
    ('Turkey', '2026-05-01'),  -- Emek ve Dayanışma Günü (Labour Day)
    ('Turkey', '2026-05-19'),  -- Atatürk'ü Anma, Gençlik ve Spor Bayramı
    ('Turkey', '2026-05-26'),  -- Kurban Bayramı Arifesi (~arife)
    ('Turkey', '2026-05-27'),  -- Kurban Bayramı 1. Günü (Eid ul-Adha)
    ('Turkey', '2026-05-28'),  -- Kurban Bayramı 2. Günü
    ('Turkey', '2026-05-29'),  -- Kurban Bayramı 3. Günü
    ('Turkey', '2026-07-15'),  -- Demokrasi ve Millî Birlik Günü
    ('Turkey', '2026-08-31'),  -- Zafer Bayramı — zamena (30.8. = нед → 31.8. = pon)
    ('Turkey', '2026-10-29'),  -- Cumhuriyet Bayramı (Republic Day)

-- ─── Sweden ───────────────────────────────────────────────────────────────────
-- Nasdaq Stockholm / Svenska Handelsbanken SVEX
-- Midsommarafton: petak između 19. i 25. juna → 19.6.2026
-- Alla helgons dag (1.11.) pada u nedelju → bez surogatnog praznika za berzu
    ('Sweden', '2026-01-01'),  -- Nyårsdagen
    ('Sweden', '2026-01-06'),  -- Trettondag jul (Epiphany)
    ('Sweden', '2026-04-03'),  -- Långfredag (Good Friday)
    ('Sweden', '2026-04-06'),  -- Annandag påsk (Easter Monday)
    ('Sweden', '2026-05-01'),  -- Första maj (Labour Day)
    ('Sweden', '2026-05-14'),  -- Kristi himmelsfärdsdag (Ascension Day)
    ('Sweden', '2026-06-19'),  -- Midsommarafton (Midsummer Eve — petak 19.–25.6.)
    ('Sweden', '2026-12-24'),  -- Julafton (Christmas Eve)
    ('Sweden', '2026-12-25'),  -- Juldagen (Christmas Day)

-- ─── Hungary ──────────────────────────────────────────────────────────────────
-- Budapesti Értéktőzsde (BÉT) / Xtend
-- Március 15. (Nem. ünnep) pada u nedelju → zamena ponedeljak 16.3.
-- Mindenszentek (1.11.) pada u nedelju → zamena ponedeljak 2.11.
    ('Hungary', '2026-01-01'),  -- Újév (New Year's Day)
    ('Hungary', '2026-03-16'),  -- Nemzeti ünnep — zamena (15.3. = нед → 16.3. = pon)
    ('Hungary', '2026-04-03'),  -- Nagypéntek (Good Friday — obnovljen praznik od 2017)
    ('Hungary', '2026-04-06'),  -- Húsvét hétfő (Easter Monday)
    ('Hungary', '2026-05-01'),  -- A Munka Ünnepe (Labour Day)
    ('Hungary', '2026-05-25'),  -- Pünkösd hétfő (Whit Monday)
    ('Hungary', '2026-08-20'),  -- Az államalapítás ünnepe (St. Stephen's Day)
    ('Hungary', '2026-10-23'),  -- Nemzeti ünnep — 1956 (Republic Day)
    ('Hungary', '2026-11-02'),  -- Mindenszentek — zamena (1.11. = нед → 2.11. = pon)
    ('Hungary', '2026-12-24'),  -- Karácsony este (Christmas Eve)
    ('Hungary', '2026-12-25'),  -- Karácsony első napja (Christmas Day)

-- ─── Bulgaria ─────────────────────────────────────────────────────────────────
-- Bulgarska Fondova Bursa (BSE) — Alternative Market (ABUL)
-- Pravoslavni Uskrs 2026 = 19. april
-- Praznici koji padaju u nedelju dobijaju zamenu u ponedeljak:
--   Ден на будителите (1.11. = нед → 2.11.), Съединение (6.9. = нед → 7.9.),
--   Кирил и Методий (24.5. = нед → 25.5.)
    ('Bulgaria', '2026-01-01'),  -- Нова Година (New Year's Day)
    ('Bulgaria', '2026-03-03'),  -- Ден на Освобождението (Liberation Day)
    ('Bulgaria', '2026-04-17'),  -- Велики Петък (Orthodox Good Friday)
    ('Bulgaria', '2026-04-20'),  -- Великден (Orthodox Easter Monday)
    ('Bulgaria', '2026-05-01'),  -- Ден на труда и международната работническа солидарност
    ('Bulgaria', '2026-05-06'),  -- Гергьовден / Ден на храбростта и Българската армия
    ('Bulgaria', '2026-05-25'),  -- Ден на Кирил и Методий — замена (24.5. = нед → 25.5.)
    ('Bulgaria', '2026-09-07'),  -- Ден на Съединението — замена (6.9. = нед → 7.9.)
    ('Bulgaria', '2026-09-22'),  -- Ден на Независимостта на България
    ('Bulgaria', '2026-11-02'),  -- Ден на народните будители — замена (1.11. = нед → 2.11.)
    ('Bulgaria', '2026-12-24'),  -- Бъдни вечер (Christmas Eve)
    ('Bulgaria', '2026-12-25'),  -- Коледа (Christmas Day)

-- ─── Luxembourg ───────────────────────────────────────────────────────────────
-- LuxSE (Luxembourg Stock Exchange), BIL, BCEE
-- Toussaint (1.11.) pada u nedelju → zamena ponedeljak 2.11.
    ('Luxembourg', '2026-01-01'),  -- Jour de l'An
    ('Luxembourg', '2026-04-03'),  -- Vendredi Saint (Good Friday)
    ('Luxembourg', '2026-04-06'),  -- Lundi de Pâques (Easter Monday)
    ('Luxembourg', '2026-05-01'),  -- Fête du Travail (Labour Day)
    ('Luxembourg', '2026-05-14'),  -- Ascension
    ('Luxembourg', '2026-05-25'),  -- Lundi de Pentecôte (Whit Monday)
    ('Luxembourg', '2026-06-23'),  -- Fête Nationale (Anniversaire du Grand-Duc)
    ('Luxembourg', '2026-11-02'),  -- Toussaint — zamena (1.11. = нед → 2.11. = pon)
    ('Luxembourg', '2026-12-25'),  -- Noël (Christmas Day)

-- ─── Poland ───────────────────────────────────────────────────────────────────
-- GPW (Giełda Papierów Wartościowych w Warszawie)
-- Uwaga: Święto Konstytucji (3.5.) pada u nedelju — berza ne daje surogatni dan
-- Wniebowzięcie NMP (15.8.) pada u subotu — bez efekta
-- Wszystkich Świętych (1.11.) pada u nedelju — berza ne daje surogatni dan
    ('Poland', '2026-01-01'),  -- Nowy Rok (New Year's Day)
    ('Poland', '2026-01-06'),  -- Trzech Króli (Epiphany)
    ('Poland', '2026-04-06'),  -- Poniedziałek Wielkanocny (Easter Monday)
    ('Poland', '2026-05-01'),  -- Święto Pracy (Labour Day)
    ('Poland', '2026-05-25'),  -- Zielone Świątki / Zesłanie Ducha Świętego (Whit Monday)
    ('Poland', '2026-06-04'),  -- Boże Ciało (Corpus Christi — 60 dana posle Uskrsa)
    ('Poland', '2026-11-11'),  -- Narodowe Święto Niepodległości (Independence Day)
    ('Poland', '2026-12-25')   -- Boże Narodzenie (Christmas Day)

ON CONFLICT (polity, date) DO NOTHING;
