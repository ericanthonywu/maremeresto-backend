-- A delivery is now dispatched with one status action. Courier identity and
-- vehicle details are deliberately not collected or shown to customers.
ALTER TABLE orders
    DROP COLUMN IF EXISTS driver_name,
    DROP COLUMN IF EXISTS driver_phone,
    DROP COLUMN IF EXISTS driver_vehicle,
    DROP COLUMN IF EXISTS driver_plate,
    DROP COLUMN IF EXISTS driver_rating,
    DROP COLUMN IF EXISTS driver_assigned_at;

-- Start the simplified operational flow with no legacy order, payment, item,
-- status-history, or feedback records. Dependent order tables are cleared by
-- their foreign-key cascade.
TRUNCATE TABLE orders CASCADE;

-- Correct the production outlet name without changing its stable slug.
UPDATE branches
SET name = 'Mareme Slamet Riyadi', updated_at = NOW()
WHERE slug = 'mak-djan';
