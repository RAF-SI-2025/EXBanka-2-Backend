-- Reverse order: accounts first, then the firm
DELETE FROM core_banking.racun
WHERE id_firme = (SELECT id FROM core_banking.firma WHERE maticni_broj = '12345678');

DELETE FROM core_banking.firma
WHERE maticni_broj = '12345678';
