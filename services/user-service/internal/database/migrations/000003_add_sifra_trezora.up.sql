-- Migration: 000003_add_sifra_trezora
-- Dodaje polje sifra_trezora u client_details za PIN/lozinku trezora klijenta.
-- Vrednost se čuva kao bcrypt hash — nikad plain text.

ALTER TABLE client_details
    ADD COLUMN IF NOT EXISTS sifra_trezora VARCHAR(255);
