-- Seed branches
INSERT INTO branches (id, slug, name, address, phone, latitude, longitude, gradient_theme, icon, facility_tags, rating, is_open)
VALUES
    ('11111111-1111-1111-1111-111111111111', 'sudirman', 'Cafe Olga Sudirman', 'Jl. Jend. Sudirman No. 45, Gedung Plaza Central, Jakarta Pusat', '021-5550-0101', -6.2088, 106.8216, 'brand', 'fa-building', '["WiFi Cepat", "Full AC", "Parkir Valet"]'::jsonb, 4.8, true),
    ('22222222-2222-2222-2222-222222222222', 'kemang', 'Cafe Olga Kemang', 'Jl. Kemang Raya No. 12, Mampang Prapatan, Jakarta Selatan', '021-5550-0202', -6.2636, 106.8171, 'emerald', 'fa-tree', '["Outdoor Garden", "Pet Friendly", "Banyak Colokan"]'::jsonb, 4.9, true),
    ('33333333-3333-3333-3333-333333333333', 'bsd', 'Cafe Olga BSD City', 'Jl. Pahlawan Seribu No. 8, BSD Green Office Park, BSD City', '021-5550-0303', -6.3022, 106.6527, 'indigo', 'fa-city', '["Meeting Room", "Luas Parkir", "Manual Brew Bar"]'::jsonb, 4.7, true)
ON CONFLICT (slug) DO NOTHING;

-- Seed branch settings
INSERT INTO branch_settings (branch_id, max_delivery_radius_km, base_delivery_fee_near, base_delivery_fee_mid, base_delivery_fee_far, near_threshold_km, mid_threshold_km, service_fee, min_order_amount, free_delivery_threshold, whatsapp_number, description)
VALUES
    ('11111111-1111-1111-1111-111111111111', 10, 8000, 12000, 18000, 3, 7, 2000, 20000, 150000, '081234567890', 'Freshly Brewed, Warmly Served. Specializing in artisan coffee, artisan croissants, and hearty meals.'),
    ('22222222-2222-2222-2222-222222222222', 10, 7000, 11000, 16000, 3, 7, 2000, 20000, 150000, '081234567891', 'Garden sanctuary in Kemang. Specialty brew and brunch favorites.'),
    ('33333333-3333-3333-3333-333333333333', 10, 9000, 14000, 20000, 3, 7, 2000, 20000, 150000, '081234567892', 'Modern creative hub at BSD Green Office Park.')
ON CONFLICT (branch_id) DO NOTHING;

-- Seed categories
INSERT INTO categories (id, name, slug, emoji, sort_order)
VALUES
    ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'Kopi & Espresso', 'coffee', '☕', 1),
    ('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 'Non-Kopi', 'non-coffee', '🧊', 2),
    ('cccccccc-cccc-cccc-cccc-cccccccccccc', 'Pastry & Bakery', 'pastry', '🥐', 3),
    ('dddddddd-dddd-dddd-dddd-dddddddddddd', 'Main Course', 'meals', '🍝', 4),
    ('eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee', 'Desserts', 'dessert', '🍰', 5)
ON CONFLICT (slug) DO NOTHING;

-- Seed default owner and branch admin users
-- password is "password" hashed with bcrypt: $2a$10$7RkPqU0wFm57VzYJd0N2j.wM7f5nLsq8b7J0wXh/HlE2h9QyK.0h2
INSERT INTO users (id, phone, name, role, branch_id, password_hash)
VALUES
    ('99999999-9999-9999-9999-999999999999', '+6281100000001', 'Owner HQ', 'owner', NULL, '$2a$10$7RkPqU0wFm57VzYJd0N2j.wM7f5nLsq8b7J0wXh/HlE2h9QyK.0h2'),
    ('88888888-8888-8888-8888-888888888881', '+6281100000002', 'Admin Sudirman', 'branch_admin', '11111111-1111-1111-1111-111111111111', '$2a$10$7RkPqU0wFm57VzYJd0N2j.wM7f5nLsq8b7J0wXh/HlE2h9QyK.0h2'),
    ('88888888-8888-8888-8888-888888888882', '+6281100000003', 'Admin Kemang', 'branch_admin', '22222222-2222-2222-2222-222222222222', '$2a$10$7RkPqU0wFm57VzYJd0N2j.wM7f5nLsq8b7J0wXh/HlE2h9QyK.0h2'),
    ('88888888-8888-8888-8888-888888888883', '+6281100000004', 'Admin BSD', 'branch_admin', '33333333-3333-3333-3333-333333333333', '$2a$10$7RkPqU0wFm57VzYJd0N2j.wM7f5nLsq8b7J0wXh/HlE2h9QyK.0h2')
ON CONFLICT (phone) DO NOTHING;

-- Seed promos
INSERT INTO promos (code, type, discount_amount, is_active, min_spend)
VALUES
    ('OLGACOFFEE', 'fixed', 10000, true, 30000),
    ('GRATISONGKIR', 'free_delivery', 0, true, 50000)
ON CONFLICT (code) DO NOTHING;

-- Function to seed 17 menu items for a specific branch
DO $$
DECLARE
    b_id UUID;
    cat_coffee UUID := 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa';
    cat_non_coffee UUID := 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb';
    cat_pastry UUID := 'cccccccc-cccc-cccc-cccc-cccccccccccc';
    cat_meals UUID := 'dddddddd-dddd-dddd-dddd-dddddddddddd';
    cat_dessert UUID := 'eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee';
BEGIN
    FOR b_id IN SELECT id FROM branches LOOP
        -- Coffee & Espresso
        INSERT INTO menu_items (branch_id, category_id, name, description, price, icon, icon_bg_class, tag, is_available, sort_order)
        VALUES
            (b_id, cat_coffee, 'Americano', 'Classic black espresso roast, clean and aromatic.', 25000, 'fa-mug-hot', 'bg-stone-100', NULL, true, 1),
            (b_id, cat_coffee, 'Café Latte', 'Smooth double espresso with velvet steamed milk.', 32000, 'fa-mug-saucer', 'bg-amber-50', 'Best', true, 2),
            (b_id, cat_coffee, 'Cappuccino', 'Rich espresso with thick foamy milk & cocoa dust.', 30000, 'fa-coffee', 'bg-amber-50', NULL, true, 3),
            (b_id, cat_coffee, 'Caramel Macchiato', 'Vanilla syrup, steamed milk, espresso & caramel drizzle.', 38000, 'fa-whiskey-glass', 'bg-amber-50', 'Favorit', true, 4),
            (b_id, cat_coffee, 'Mocha', 'Espresso blend with premium dark chocolate & milk.', 35000, 'fa-mug-hot', 'bg-stone-100', NULL, true, 5),
            (b_id, cat_coffee, 'Espresso Single', 'Bold, aromatic & thick crema single shot.', 20000, 'fa-mug-saucer', 'bg-stone-100', NULL, true, 6);

        -- Non-Coffee Drinks
        INSERT INTO menu_items (branch_id, category_id, name, description, price, icon, icon_bg_class, tag, is_available, sort_order)
        VALUES
            (b_id, cat_non_coffee, 'Matcha Latte', 'Pure Uji Kyoto ceremonial matcha with fresh milk.', 33000, 'fa-leaf', 'bg-emerald-50', NULL, true, 7),
            (b_id, cat_non_coffee, 'Chocolate Deluxe', 'Rich Belgian dark chocolate blend with steamed milk.', 30000, 'fa-cookie-bite', 'bg-amber-50', NULL, true, 8),
            (b_id, cat_non_coffee, 'Iced Lychee Tea', 'Fresh brewed black tea with real whole sweet lychees.', 24000, 'fa-glass-water', 'bg-stone-100', NULL, true, 9),
            (b_id, cat_non_coffee, 'Taro Latte', 'Creamy aromatic purple taro with fresh warm milk.', 28000, 'fa-mug-hot', 'bg-purple-50', NULL, true, 10);

        -- Pastry & Bakery
        INSERT INTO menu_items (branch_id, category_id, name, description, price, icon, icon_bg_class, tag, is_available, sort_order)
        VALUES
            (b_id, cat_pastry, 'Butter Croissant', 'Golden, flaky & crispy layers of French butter goodness.', 22000, 'fa-bread-slice', 'bg-amber-50', NULL, true, 11),
            (b_id, cat_pastry, 'Almond Croissant', 'Filled with almond frangipane & toasted almond flakes.', 26000, 'fa-bread-slice', 'bg-amber-50', NULL, true, 12),
            (b_id, cat_pastry, 'Truffle Fries', 'Crispy shoestring fries tossed in white truffle oil & parmesan.', 24000, 'fa-bowl-food', 'bg-amber-50', NULL, true, 13);

        -- Main Course
        INSERT INTO menu_items (branch_id, category_id, name, description, price, icon, icon_bg_class, tag, is_available, sort_order)
        VALUES
            (b_id, cat_meals, 'Chicken Pasta Alfredo', 'Creamy parmesan fettuccine with juicy grilled chicken.', 48000, 'fa-bowl-food', 'bg-amber-50', NULL, true, 14),
            (b_id, cat_meals, 'Olga Beef Burger', '150g grilled Australian patty, melted cheddar on brioche.', 52000, 'fa-burger', 'bg-amber-50', NULL, true, 15);

        -- Desserts
        INSERT INTO menu_items (branch_id, category_id, name, description, price, icon, icon_bg_class, tag, is_available, sort_order)
        VALUES
            (b_id, cat_dessert, 'Tiramisu Slice', 'Authentic Italian mascarpone with espresso-soaked ladyfingers.', 35000, 'fa-cake-candles', 'bg-pink-50', NULL, true, 16),
            (b_id, cat_dessert, 'Classic Affogato', 'Hot double espresso shot poured over cold Madagascar vanilla gelato.', 30000, 'fa-ice-cream', 'bg-stone-100', NULL, true, 17);
    END LOOP;
END $$;
