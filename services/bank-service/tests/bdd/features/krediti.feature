Feature: Upravljanje kreditima i otplatom kredita

  # Scenario 33
  Scenario: Podnošenje zahteva za kredit
    Given klijent je ulogovan u aplikaciju
    When popuni formu za kredit sa validnim podacima
    Then sistem beleži zahtev za kredit sa statusom "NA_CEKANJU"

  # Scenario 34
  Scenario: Pregled kredita klijenta
    Given klijent je ulogovan u aplikaciju
    And klijent ima aktivne kredite
    When klijent otvori sekciju "Krediti"
    Then prikazuje se lista svih kredita klijenta

  # Scenario 35
  Scenario: Odobravanje kredita od strane zaposlenog
    Given zaposleni je ulogovan u portal za upravljanje kreditima
    And postoji zahtev za kredit od strane klijenta
    When zaposleni odobri zahtev za kredit
    Then kredit dobija status "ODOBREN"
    And iznos kredita se uplaćuje na račun klijenta

  # Scenario 36
  Scenario: Odbijanje zahteva za kredit
    Given zaposleni je na portalu za upravljanje kreditima
    And postoji zahtev za kredit klijenta
    When zaposleni klikne na dugme "Odbij"
    Then zahtev dobija status "ODBIJEN"

  # Scenario 37
  Scenario: Automatsko skidanje rate kredita
    Given postoji aktivan kredit
    And datum sledeće rate je današnji dan
    And klijent ima dovoljno sredstava na računu
    When sistem pokrene dnevni cron job
    Then iznos rate se automatski skida sa računa klijenta
    And sledeći datum plaćanja se pomera za jedan mesec
    And klijent dobija obaveštenje o uspešnoj naplati rate

  # Scenario 38
  Scenario: Kašnjenje u otplati kredita zbog nedovoljnih sredstava
    Given postoji aktivan kredit
    And datum sledeće rate je današnji dan
    And klijent nema dovoljno sredstava na računu
    When sistem pokrene cron job za naplatu rate
    Then rata dobija status "Kasni"
    And sistem planira novi pokušaj naplate nakon 72 sata
    And klijent dobija obaveštenje o neuspešnoj naplati
