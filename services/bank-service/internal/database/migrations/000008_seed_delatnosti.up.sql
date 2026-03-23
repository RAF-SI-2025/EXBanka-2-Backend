INSERT INTO core_banking.delatnost (sifra, naziv, grana, sektor)
VALUES
    -- Poljoprivreda, šumarstvo i ribarstvo
    ('1.11',  'Uzgoj žitarica i mahunarki',                                          'Uzgoj žitarica i mahunarki',                                          'Poljoprivreda, šumarstvo i ribarstvo'),
    ('1.13',  'Uzgoj povrća',                                                         'Uzgoj povrća',                                                         'Poljoprivreda, šumarstvo i ribarstvo'),

    -- Prerađivačka industrija
    ('13.1',  'Priprema i predenje tekstilnih vlakana',                               'Priprema i predenje tekstilnih vlakana',                               'Prerađivačka industrija'),
    ('24.1',  'Proizvodnja gvožđa i čelika',                                          'Proizvodnja gvožđa i čelika',                                          'Prerađivačka industrija'),
    ('24.2',  'Proizvodnja čeličnih cevi, šupljih profila i fitinga',                 'Proizvodnja čeličnih cevi, šupljih profila i fitinga',                 'Prerađivačka industrija'),

    -- Građevinarstvo
    ('41.1',  'Razvoj građevinskih projekata',                                        'Razvoj građevinskih projekata',                                        'Građevinarstvo'),
    ('41.2',  'Izgradnja stambenih i nestambenih zgrada',                             'Izgradnja stambenih i nestambenih zgrada',                             'Građevinarstvo'),
    ('42.11', 'Izgradnja puteva i autoputeva',                                        'Izgradnja puteva i autoputeva',                                        'Građevinarstvo'),
    ('42.12', 'Izgradnja železničkih i podzemnih pruga',                              'Izgradnja železničkih i podzemnih pruga',                              'Građevinarstvo'),
    ('42.13', 'Izgradnja mostova i tunela',                                           'Izgradnja mostova i tunela',                                           'Građevinarstvo'),
    ('42.21', 'Izgradnja vodovodnih projekata',                                       'Izgradnja vodovodnih projekata',                                       'Građevinarstvo'),
    ('42.22', 'Izgradnja elektroenergetskih i telekomunikacionih mreža',              'Izgradnja elektroenergetskih i telekomunikacionih mreža',              'Građevinarstvo'),

    -- Rudarstvo
    ('5.1',   'Vađenje uglja',                                                        'Vađenje uglja',                                                        'Rudarstvo'),
    ('7.1',   'Vađenje gvozdenih ruda',                                               'Vađenje gvozdenih ruda',                                               'Rudarstvo'),
    ('7.21',  'Vađenje uranijuma i torijuma',                                         'Vađenje uranijuma i torijuma',                                         'Rudarstvo'),
    ('8.11',  'Eksploatacija ukrasnog i građevinskog kamena',                         'Eksploatacija ukrasnog i građevinskog kamena',                         'Rudarstvo'),
    ('8.92',  'Ekstrakcija treseta',                                                  'Ekstrakcija treseta',                                                  'Rudarstvo'),

    -- Trgovina
    ('47.11', 'Trgovina u nespecijalizovanim prodavnicama sa hranom i pićem',         'Trgovina u nespecijalizovanim prodavnicama sa hranom i pićem',         'Trgovina'),

    -- Ugostiteljstvo
    ('56.1',  'Restorani i pokretni ugostiteljski objekti',                           'Restorani i pokretni ugostiteljski objekti',                           'Ugostiteljstvo'),

    -- IT
    ('62.01', 'Računarsko programiranje',                                             'Računarsko programiranje',                                             'IT'),
    ('62.09', 'Ostale IT usluge',                                                     'Ostale IT usluge',                                                     'IT'),
    ('63.11', 'Obrada podataka, hosting i slične delatnosti',                         'Obrada podataka, hosting i slične delatnosti',                         'IT'),

    -- Finansijske delatnosti
    ('64.19', 'Ostale monetarne posredničke delatnosti',                              'Ostale monetarne posredničke delatnosti',                              'Finansijske delatnosti'),
    ('64.2',  'Holding kompanije',                                                    'Holding kompanije',                                                    'Finansijske delatnosti'),
    ('64.91', 'Finansijski lizing',                                                   'Finansijski lizing',                                                   'Finansijske delatnosti'),
    ('66.3',  'Fondovi i slične finansijske delatnosti',                              'Fondovi i slične finansijske delatnosti',                              'Finansijske delatnosti'),

    -- Osiguranje
    ('65.11', 'Životno osiguranje',                                                   'Životno osiguranje',                                                   'Osiguranje'),
    ('65.12', 'Neživotno osiguranje',                                                 'Neživotno osiguranje',                                                 'Osiguranje'),
    ('65.2',  'Reosiguranje',                                                         'Reosiguranje',                                                         'Osiguranje'),
    ('66.21', 'Procena rizika i štete',                                               'Procena rizika i štete',                                               'Osiguranje'),

    -- Poslovanje nekretninama
    ('68.1',  'Upravljanje nekretninama na osnovu naknade ili ugovora',               'Upravljanje nekretninama na osnovu naknade ili ugovora',               'Poslovanje nekretninama'),
    ('68.2',  'Izdavanje i upravljanje nekretninama u sopstvenom ili iznajmljenom vlasništvu', 'Izdavanje i upravljanje nekretninama u sopstvenom ili iznajmljenom vlasništvu', 'Poslovanje nekretninama'),

    -- Saobraćaj i skladištenje
    ('53.1',  'Poštanske aktivnosti',                                                 'Poštanske aktivnosti',                                                 'Saobraćaj i skladištenje'),
    ('53.2',  'Kurirske aktivnosti',                                                  'Kurirske aktivnosti',                                                  'Saobraćaj i skladištenje'),

    -- Obrazovanje
    ('85.1',  'Predškolsko obrazovanje',                                              'Predškolsko obrazovanje',                                              'Obrazovanje'),
    ('85.2',  'Osnovno obrazovanje',                                                  'Osnovno obrazovanje',                                                  'Obrazovanje'),

    -- Zdravstvena zaštita
    ('86.1',  'Bolničke aktivnosti',                                                  'Bolničke aktivnosti',                                                  'Zdravstvena zaštita'),
    ('86.21', 'Opšta medicinska praksa',                                              'Opšta medicinska praksa',                                              'Zdravstvena zaštita'),
    ('86.22', 'Specijalistička medicinska praksa',                                    'Specijalistička medicinska praksa',                                    'Zdravstvena zaštita'),
    ('86.9',  'Ostale aktivnosti zdravstvene zaštite',                                'Ostale aktivnosti zdravstvene zaštite',                                'Zdravstvena zaštita'),

    -- Javna uprava i odbrana
    ('84.12', 'Regulisanje delatnosti privrede',                                      'Regulisanje delatnosti privrede',                                      'Javna uprava i odbrana'),

    -- Kultura, sport i rekreacija
    ('90.01', 'Delatnost pozorišta',                                                  'Delatnost pozorišta',                                                  'Kultura, sport i rekreacija'),
    ('90.02', 'Delatnost muzeja',                                                     'Delatnost muzeja',                                                     'Kultura, sport i rekreacija'),
    ('90.04', 'Delatnost botaničkih i zooloških vrtova',                              'Delatnost botaničkih i zooloških vrtova',                              'Kultura, sport i rekreacija'),

    -- Sportske i rekreativne delatnosti
    ('93.11', 'Delovanje sportskih objekata',                                         'Delovanje sportskih objekata',                                         'Sportske i rekreativne delatnosti'),
    ('93.13', 'Delovanje teretana',                                                   'Delovanje teretana',                                                   'Sportske i rekreativne delatnosti'),
    ('93.19', 'Ostale sportske delatnosti',                                           'Ostale sportske delatnosti',                                           'Sportske i rekreativne delatnosti'),

    -- Proizvodnja elektronskih komponenti
    ('26.11', 'Proizvodnja elektronskih komponenti',                                  'Proizvodnja elektronskih komponenti',                                  'Proizvodnja elektronskih komponenti'),

    -- Proizvodnja električne opreme
    ('27.12', 'Proizvodnja električnih panela i ploča',                               'Proizvodnja električnih panela i ploča',                               'Proizvodnja električne opreme'),

    -- Proizvodnja motornih vozila
    ('29.1',  'Proizvodnja motornih vozila',                                          'Proizvodnja motornih vozila',                                          'Proizvodnja motornih vozila')

ON CONFLICT (sifra) DO NOTHING;
