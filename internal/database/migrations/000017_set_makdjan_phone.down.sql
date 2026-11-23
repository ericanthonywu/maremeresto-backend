-- 000017_set_makdjan_phone.down.sql
UPDATE branches
SET phone = '', updated_at = NOW()
WHERE id = '33333333-3333-3333-3333-333333333333' OR slug = 'mak-djan';
