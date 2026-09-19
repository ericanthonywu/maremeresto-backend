-- 000012_add_staff_username.up.sql
-- Add username column to users table for staff authentication and credentials management.

ALTER TABLE users ADD COLUMN IF NOT EXISTS username VARCHAR(100);
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_username ON users(LOWER(username)) WHERE username IS NOT NULL;

-- Backfill default usernames for existing seeded staff accounts
UPDATE users SET username = 'owner' WHERE email = 'owner@cafeolga.id' AND (username IS NULL OR username = '');
UPDATE users SET username = 'admin.kerten' WHERE email = 'admin.kerten@cafeolga.id' AND (username IS NULL OR username = '');
UPDATE users SET username = 'admin.makamhaji' WHERE email = 'admin.makamhaji@cafeolga.id' AND (username IS NULL OR username = '');
UPDATE users SET username = 'admin.makdjan' WHERE email = 'admin.makdjan@cafeolga.id' AND (username IS NULL OR username = '');
