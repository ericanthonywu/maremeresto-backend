-- 000017_set_makdjan_phone.up.sql
-- Set Mak Djan outlet contact phone number (+6282147551009).

UPDATE branches
SET phone = '+6282147551009', updated_at = NOW()
WHERE id = '33333333-3333-3333-3333-333333333333' OR slug = 'mak-djan';

UPDATE branch_settings
SET whatsapp_number = '+6282147551009', updated_at = NOW()
WHERE branch_id = '33333333-3333-3333-3333-333333333333';
