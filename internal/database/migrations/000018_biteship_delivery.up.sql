-- 000018_biteship_delivery.up.sql
-- Add Biteship courier selection, booking, and live tracking columns to orders.

ALTER TABLE orders
    ADD COLUMN IF NOT EXISTS courier_company VARCHAR(30),
    ADD COLUMN IF NOT EXISTS courier_type VARCHAR(30),
    ADD COLUMN IF NOT EXISTS courier_name VARCHAR(100),
    ADD COLUMN IF NOT EXISTS biteship_order_id VARCHAR(100),
    ADD COLUMN IF NOT EXISTS biteship_tracking_id VARCHAR(100),
    ADD COLUMN IF NOT EXISTS biteship_waybill_id VARCHAR(100),
    ADD COLUMN IF NOT EXISTS biteship_tracking_url TEXT,
    ADD COLUMN IF NOT EXISTS biteship_status VARCHAR(50),
    ADD COLUMN IF NOT EXISTS driver_name VARCHAR(100),
    ADD COLUMN IF NOT EXISTS driver_phone VARCHAR(50),
    ADD COLUMN IF NOT EXISTS driver_plate VARCHAR(50),
    ADD COLUMN IF NOT EXISTS driver_photo_url TEXT,
    ADD COLUMN IF NOT EXISTS courier_dispatched_at TIMESTAMP WITH TIME ZONE;

CREATE INDEX IF NOT EXISTS idx_orders_biteship_order_id ON orders(biteship_order_id) WHERE biteship_order_id IS NOT NULL;
