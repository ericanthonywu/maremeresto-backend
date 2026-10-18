-- 000015_add_is_best_seller.down.sql
DROP INDEX IF EXISTS idx_menu_items_best_seller;
ALTER TABLE menu_items DROP COLUMN IF EXISTS is_best_seller;
