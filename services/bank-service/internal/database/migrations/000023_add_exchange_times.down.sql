ALTER TABLE core_banking.exchange
    DROP COLUMN IF EXISTS open_time,
    DROP COLUMN IF EXISTS close_time;
