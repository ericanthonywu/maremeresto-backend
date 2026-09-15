-- Enhance order_feedback to support restaurant rating, application rating,
-- specific reasons for each rating, and item-by-item feedback for ordered items.
ALTER TABLE order_feedback
    ADD COLUMN IF NOT EXISTS resto_rating SMALLINT CHECK (resto_rating IS NULL OR (resto_rating BETWEEN 1 AND 5)),
    ADD COLUMN IF NOT EXISTS app_rating SMALLINT CHECK (app_rating IS NULL OR (app_rating BETWEEN 1 AND 5)),
    ADD COLUMN IF NOT EXISTS resto_reason TEXT,
    ADD COLUMN IF NOT EXISTS app_reason TEXT,
    ADD COLUMN IF NOT EXISTS items_feedback JSONB NOT NULL DEFAULT '[]'::jsonb;

-- Also store rating and review reason directly on order_items for item-level analysis
ALTER TABLE order_items
    ADD COLUMN IF NOT EXISTS rating SMALLINT CHECK (rating IS NULL OR (rating BETWEEN 1 AND 5)),
    ADD COLUMN IF NOT EXISTS review_reason TEXT;
