-- =============================================================================
-- Migration: 000002_seed_trezor_user
-- Service:   user-service
--
-- Creates the treasury system-client that owns the EXBanka company entity
-- in bank-service (cross-service reference by plain BIGINT).
--
-- This user is NOT a real person and will never log in.
-- birth_date = 946684800000 ms = 2000-01-01 00:00:00 UTC (symbolic founding epoch).
-- =============================================================================

INSERT INTO users (email, password_hash, salt_password, user_type, first_name, last_name, birth_date, is_active)
VALUES (
    'trezor@exbanka.rs',
    '',               -- no password; system account only
    '',
    'CLIENT',
    'Trezor',
    'EXBanka',
    946684800000,     -- epoch-ms: 2000-01-01 00:00:00 UTC
    TRUE
)
ON CONFLICT (email) DO NOTHING;

-- client_details row is required for all CLIENT users
INSERT INTO client_details (user_id)
SELECT id FROM users WHERE email = 'trezor@exbanka.rs'
ON CONFLICT (user_id) DO NOTHING;
