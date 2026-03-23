-- client_details removed via ON DELETE CASCADE on users.id
DELETE FROM users WHERE email = 'trezor@exbanka.rs';
