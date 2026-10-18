-- 000015_add_is_best_seller.up.sql
-- Add is_best_seller column to menu_items table
ALTER TABLE menu_items ADD COLUMN IF NOT EXISTS is_best_seller BOOLEAN NOT NULL DEFAULT false;

CREATE INDEX IF NOT EXISTS idx_menu_items_best_seller ON menu_items(branch_id, is_best_seller) WHERE is_best_seller = true;

-- Seed existing popular / tagged items as best seller
UPDATE menu_items
SET is_best_seller = true
WHERE tag ILIKE '%best%'
   OR tag ILIKE '%favorit%'
   OR name ILIKE '%nasi goreng spesial%'
   OR name ILIKE '%es kopi susu gula aren%'
   OR name ILIKE '%ayam geprek%'
   OR name ILIKE '%mie nyemek%'
   OR name ILIKE '%nasi goreng mawut%'
   OR name ILIKE '%nasi goreng ijo istimewa%';
