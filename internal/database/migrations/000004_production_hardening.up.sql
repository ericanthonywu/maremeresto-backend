-- Production hardening: remove fabricated demo data, add columns the app now needs.

-- 1. Driver details were seeded with placeholder defaults ("Andi Pratama",
--    "B 1234 ABC", ...). A driver is now assigned explicitly by branch staff,
--    so the columns must start empty.
ALTER TABLE orders ALTER COLUMN driver_name    DROP DEFAULT;
ALTER TABLE orders ALTER COLUMN driver_phone   DROP DEFAULT;
ALTER TABLE orders ALTER COLUMN driver_vehicle DROP DEFAULT;
ALTER TABLE orders ALTER COLUMN driver_plate   DROP DEFAULT;
ALTER TABLE orders ALTER COLUMN driver_rating  DROP DEFAULT;

UPDATE orders
SET driver_name    = NULL,
    driver_phone   = NULL,
    driver_vehicle = NULL,
    driver_plate   = NULL,
    driver_rating  = NULL
WHERE driver_name   = 'Andi Pratama'
   OR driver_plate  = 'B 1234 ABC'
   OR driver_phone  = '081234567890';

ALTER TABLE orders ADD COLUMN IF NOT EXISTS driver_assigned_at TIMESTAMP WITH TIME ZONE;

-- 2. Admin "new order" notifications: staff acknowledge an order so the unread
--    badge survives a page reload instead of living only in browser memory.
ALTER TABLE orders ADD COLUMN IF NOT EXISTS acknowledged_at TIMESTAMP WITH TIME ZONE;
ALTER TABLE orders ADD COLUMN IF NOT EXISTS acknowledged_by UUID REFERENCES users(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_orders_branch_status    ON orders(branch_id, status);
CREATE INDEX IF NOT EXISTS idx_orders_branch_created   ON orders(branch_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_orders_unacknowledged   ON orders(branch_id) WHERE acknowledged_at IS NULL;

-- 3. The placeholder outlet WhatsApp number must not be a working-looking one.
ALTER TABLE branch_settings ALTER COLUMN whatsapp_number DROP DEFAULT;

-- 4. Promo validity windows and redemption caps (previously unbounded).
ALTER TABLE promos ADD COLUMN IF NOT EXISTS valid_from  TIMESTAMP WITH TIME ZONE;
ALTER TABLE promos ADD COLUMN IF NOT EXISTS valid_until TIMESTAMP WITH TIME ZONE;
ALTER TABLE promos ADD COLUMN IF NOT EXISTS max_redemptions INT;
ALTER TABLE promos ADD COLUMN IF NOT EXISTS redemption_count INT NOT NULL DEFAULT 0;

-- 5. Menu items are ordered by sort_order; make sure the admin can rely on it.
CREATE INDEX IF NOT EXISTS idx_menu_items_branch_sort ON menu_items(branch_id, sort_order);

-- 6. Staff sign in with an email address. It used to be mapped to a phone
--    number by a hardcoded switch in the service layer, so adding an outlet
--    meant editing Go code.
ALTER TABLE users ADD COLUMN IF NOT EXISTS email VARCHAR(255);
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email ON users(LOWER(email)) WHERE email IS NOT NULL;

UPDATE users SET email = 'owner@cafeolga.id',            name = 'Owner HQ'          WHERE id = '99999999-9999-9999-9999-999999999999';
UPDATE users SET email = 'admin.kerten@cafeolga.id',     name = 'Admin Kerten'      WHERE id = '88888888-8888-8888-8888-888888888881';
UPDATE users SET email = 'admin.makamhaji@cafeolga.id',  name = 'Admin Makamhaji'   WHERE id = '88888888-8888-8888-8888-888888888882';
UPDATE users SET email = 'admin.makdjan@cafeolga.id',    name = 'Admin Mak Djan'    WHERE id = '88888888-8888-8888-8888-888888888883';

-- The seeded password_hash does not correspond to any known password and the
-- only way anyone ever signed in was a hardcoded bypass in AdminLogin, now
-- removed. Clear the unusable hashes so the accounts are explicitly locked
-- until a real password is set with:  go run ./cmd/admintool set-password
UPDATE users
SET password_hash = NULL
WHERE role IN ('branch_admin', 'owner')
  AND password_hash = '$2a$10$7RkPqU0wFm57VzYJd0N2j.wM7f5nLsq8b7J0wXh/HlE2h9QyK.0h2';

-- A staff account with no branch can see nothing; make the intent explicit.
ALTER TABLE users ADD CONSTRAINT chk_branch_admin_has_branch
    CHECK (role <> 'branch_admin' OR branch_id IS NOT NULL) NOT VALID;
