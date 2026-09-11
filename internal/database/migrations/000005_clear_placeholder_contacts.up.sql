-- 000003 seeded the 3 real Surakarta branches with invented, sequential-looking
-- contact numbers (0271-712301/2/3 and 081234567891/2/3) as placeholders. They
-- look like real numbers but aren't, and branch_settings.whatsapp_number is
-- actually surfaced to customers (the "Hubungi outlet" button on the order
-- success page). Clear them so the operator fills in real numbers via the
-- admin "Profil Outlet" / "Kontak Outlet" editors, instead of customers being
-- silently sent to a number that isn't real.
UPDATE branches
SET phone = '', updated_at = NOW()
WHERE phone IN ('0271-712301', '0271-712302', '0271-712303');

UPDATE branch_settings
SET whatsapp_number = '', updated_at = NOW()
WHERE whatsapp_number IN ('081234567891', '081234567892', '081234567893');
