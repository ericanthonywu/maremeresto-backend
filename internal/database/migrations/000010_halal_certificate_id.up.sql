-- The public halal-certificate identifier is maintained per outlet.  An
-- empty value is intentional for existing outlets until staff enters the
-- certificate number in Profil Outlet; do not invent a legal certificate ID.
ALTER TABLE branch_settings
    ADD COLUMN IF NOT EXISTS halal_certificate_id VARCHAR(160) NOT NULL DEFAULT '';
