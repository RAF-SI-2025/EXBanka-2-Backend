UPDATE core_banking.valuta
SET status = TRUE
WHERE oznaka IN ('IDR', 'INR', 'PLN', 'UAH', 'ARS', 'TRY', 'SEK', 'HUF', 'BGN');
