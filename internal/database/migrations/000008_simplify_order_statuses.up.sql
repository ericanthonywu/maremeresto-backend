-- The post-payment operational flow is now deliberately minimal:
-- accepted (belum diantar) -> completed (diantar), with rejected/refunded
-- remaining terminal exceptions. Preserve the meaning of historical orders
-- while removing obsolete in-progress states from active use.
UPDATE orders
SET status = 'accepted', updated_at = NOW()
WHERE status IN ('preparing', 'ready', 'on_the_way');

UPDATE orders
SET status = 'completed', updated_at = NOW()
WHERE status IN ('delivered', 'picked_up');

COMMENT ON COLUMN orders.status IS
    'pending, accepted (belum diantar), completed (diantar), rejected, cancelled, refunded';
