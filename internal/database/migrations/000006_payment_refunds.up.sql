-- Refunds: staff can refund a paid order through Midtrans from the admin
-- portal. This is additive and safe to run on a live database.

-- 1. Track what was actually refunded on the payment row. amount/reason are
--    what Midtrans confirmed, not just what was requested.
ALTER TABLE payments ADD COLUMN IF NOT EXISTS refund_amount INT NOT NULL DEFAULT 0;
ALTER TABLE payments ADD COLUMN IF NOT EXISTS refund_reason TEXT;
ALTER TABLE payments ADD COLUMN IF NOT EXISTS refunded_at TIMESTAMP WITH TIME ZONE;
ALTER TABLE payments ADD COLUMN IF NOT EXISTS refunded_by UUID REFERENCES users(id) ON DELETE SET NULL;
ALTER TABLE payments ADD COLUMN IF NOT EXISTS midtrans_refund_response JSONB;

-- 2. 'refund' joins the existing payments.status values (pending, settlement,
--    expire, cancel, deny) and 'refunded' is a new terminal orders.status,
--    reached from any paid status rather than through the normal state
--    machine. Both columns are plain VARCHAR with no CHECK constraint, so no
--    schema change is needed beyond the comment below.
COMMENT ON COLUMN orders.status IS
    'pending, accepted, preparing, ready, on_the_way, picked_up, delivered, completed, rejected, cancelled, refunded';
COMMENT ON COLUMN payments.status IS
    'pending, settlement, expire, cancel, deny, refund';

CREATE INDEX IF NOT EXISTS idx_payments_refunded_at ON payments(refunded_at) WHERE refunded_at IS NOT NULL;
