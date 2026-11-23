-- 000018_biteship_delivery.down.sql
DROP INDEX IF EXISTS idx_orders_biteship_order_id;

ALTER TABLE orders
    DROP COLUMN IF EXISTS courier_dispatched_at,
    DROP COLUMN IF EXISTS driver_photo_url,
    DROP COLUMN IF EXISTS driver_plate,
    DROP COLUMN IF EXISTS driver_phone,
    DROP COLUMN IF EXISTS driver_name,
    DROP COLUMN IF EXISTS biteship_status,
    DROP COLUMN IF EXISTS biteship_tracking_url,
    DROP COLUMN IF EXISTS biteship_waybill_id,
    DROP COLUMN IF EXISTS biteship_tracking_id,
    DROP COLUMN IF EXISTS biteship_order_id,
    DROP COLUMN IF EXISTS courier_name,
    DROP COLUMN IF EXISTS courier_type,
    DROP COLUMN IF EXISTS courier_company;
