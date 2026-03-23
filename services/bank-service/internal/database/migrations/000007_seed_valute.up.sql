INSERT INTO core_banking.valuta (naziv, oznaka, simbol, zemlja, status)
VALUES
    ('Srpski dinar',        'RSD', 'din.', 'Srbija',                          TRUE),
    ('Euro',                'EUR', '€',    'Belgija, Francuska, Italija...',   TRUE),
    ('Švajcarski franak',   'CHF', 'CHF',  'Švajcarska',                      TRUE),
    ('Američki dolar',      'USD', '$',    'SAD',                              TRUE),
    ('Britanska funta',     'GBP', '£',    'Ujedinjeno Kraljevstvo',           TRUE),
    ('Japanski jen',        'JPY', '¥',    'Japan',                            TRUE),
    ('Kanadski dolar',      'CAD', 'C$',   'Kanada',                           TRUE),
    ('Australijski dolar',  'AUD', 'A$',   'Australija',                       TRUE)
ON CONFLICT (oznaka) DO NOTHING;
