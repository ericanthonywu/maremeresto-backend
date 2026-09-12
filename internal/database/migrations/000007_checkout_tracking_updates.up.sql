-- Fixed, network-wide delivery policy:
-- 0–1 km free, >1–5 km Rp8.000, >5–10 km Rp12.000, maximum 10 km.
UPDATE branch_settings
SET max_delivery_radius_km = 10,
    base_delivery_fee_near = 0,
    base_delivery_fee_mid = 8000,
    base_delivery_fee_far = 12000,
    near_threshold_km = 1,
    mid_threshold_km = 5,
    free_delivery_threshold = 0,
    updated_at = NOW();

-- Keep each outlet immediately recognisable in the customer app.
UPDATE branches
SET gradient_theme = CASE slug
    WHEN 'kerten' THEN 'shopee'
    WHEN 'makamhaji' THEN 'gojek'
    WHEN 'mak-djan' THEN 'gopay'
    ELSE gradient_theme
END,
updated_at = NOW();

-- One private feedback response per order. It can be edited by the same
-- customer later, but is never exposed as an unmoderated public review.
CREATE TABLE IF NOT EXISTS order_feedback (
    order_id UUID PRIMARY KEY REFERENCES orders(id) ON DELETE CASCADE,
    rating SMALLINT NOT NULL CHECK (rating BETWEEN 1 AND 5),
    comment TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_order_feedback_created_at ON order_feedback(created_at DESC);
