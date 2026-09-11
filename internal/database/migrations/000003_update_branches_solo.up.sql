-- Update Branches to Solo / Surakarta outlets
DELETE FROM order_status_history;
DELETE FROM payments;
DELETE FROM order_items;
DELETE FROM orders;
DELETE FROM menu_items;
DELETE FROM branch_settings;
DELETE FROM branches;

-- Insert 3 new branches with exact lat/long
INSERT INTO branches (id, slug, name, address, phone, latitude, longitude, gradient_theme, icon, facility_tags, rating, is_open)
VALUES
    ('11111111-1111-1111-1111-111111111111', 'kerten', 'Mareme Kerten', 'Jl. Samratulangi No.65, Kerten, Kec. Laweyan, Kota Surakarta, Jawa Tengah 57143', '0271-712301', -7.5597, 110.7942, 'brand', 'fa-store', '["Sebelah SMA Batik 2", "WiFi Cepat", "Area Parkir Luas", "Full AC"]'::jsonb, 4.9, true),
    ('22222222-2222-2222-2222-222222222222', 'makamhaji', 'Mareme Makamhaji', 'Jl. Slamet Riyadi No.456, Dusun I, Makamhaji, Kec. Kartasura, Kabupaten Sukoharjo, Jawa Tengah 57147', '0271-712302', -7.5662, 110.7788, 'emerald', 'fa-utensils', '["Dekat Underpass", "Dine In Nyaman", "Ramah Anak", "Musholla"]'::jsonb, 4.8, true),
    ('33333333-3333-3333-3333-333333333333', 'mak-djan', 'Mak Djan', 'Jl. R. M. Said No.54a, Ketelan, Kec. Banjarsari, Kota Surakarta, Jawa Tengah 57132', '0271-712303', -7.5647, 110.8227, 'indigo', 'fa-bowl-rice', '["Pusat Kota Solo", "Dekat Mangkunegaran", "Authentic Recipe", "Outdoor"]'::jsonb, 4.9, true)
ON CONFLICT (id) DO UPDATE SET
    slug = EXCLUDED.slug,
    name = EXCLUDED.name,
    address = EXCLUDED.address,
    phone = EXCLUDED.phone,
    latitude = EXCLUDED.latitude,
    longitude = EXCLUDED.longitude,
    gradient_theme = EXCLUDED.gradient_theme,
    icon = EXCLUDED.icon,
    facility_tags = EXCLUDED.facility_tags;

-- Insert branch settings
INSERT INTO branch_settings (branch_id, max_delivery_radius_km, base_delivery_fee_near, base_delivery_fee_mid, base_delivery_fee_far, near_threshold_km, mid_threshold_km, service_fee, min_order_amount, free_delivery_threshold, whatsapp_number, description)
VALUES
    ('11111111-1111-1111-1111-111111111111', 12, 8000, 12000, 18000, 3, 7, 2000, 20000, 150000, '081234567891', 'Mareme Kerten - Sajian Selalu Halal. Spesialis masakan lezat dan kopi nikmat.'),
    ('22222222-2222-2222-2222-222222222222', 12, 7000, 11000, 16000, 3, 7, 2000, 20000, 150000, '081234567892', 'Mareme Makamhaji Kartasura - Sajian hangat untuk keluarga dan rombongan.'),
    ('33333333-3333-3333-3333-333333333333', 12, 8000, 12000, 18000, 3, 7, 2000, 20000, 150000, '081234567893', 'Mak Djan Ketelan Banjarsari - Cita rasa legendaris pusat kota Solo.')
ON CONFLICT (branch_id) DO UPDATE SET
    description = EXCLUDED.description,
    whatsapp_number = EXCLUDED.whatsapp_number;

-- Seed menu items for all 3 branches
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
        -- Kopi & Minuman
        INSERT INTO menu_items (branch_id, category_id, name, description, price, icon, icon_bg_class, tag, is_available, sort_order)
        VALUES
            (b_id, cat_coffee, 'Es Kopi Susu Gula Aren', 'Espresso roast khas dipadu susu segar & legitnya gula aren organik.', 22000, 'fa-mug-saucer', 'bg-amber-50', 'Best', true, 1),
            (b_id, cat_coffee, 'Americano Dingin', 'Ekstraksi ganda espresso premium, segar dan harum.', 18000, 'fa-mug-hot', 'bg-stone-100', NULL, true, 2),
            (b_id, cat_coffee, 'Caramel Latte', 'Espresso lembut dengan steamed milk dan saus karamel gurih.', 28000, 'fa-whiskey-glass', 'bg-amber-50', 'Favorit', true, 3),
            (b_id, cat_non_coffee, 'Matcha Latte Uji', 'Matcha autentik Kyoto berpadu susu segar dingin/hangat.', 26000, 'fa-leaf', 'bg-emerald-50', NULL, true, 4),
            (b_id, cat_non_coffee, 'Es Lychee Tea', 'Teh hitam segar disajikan dengan buah leci manis asli.', 20000, 'fa-glass-water', 'bg-stone-100', NULL, true, 5);

        -- Sajian Utama & Makanan Khas
        INSERT INTO menu_items (branch_id, category_id, name, description, price, icon, icon_bg_class, tag, is_available, sort_order)
        VALUES
            (b_id, cat_meals, 'Nasi Ayam Geprek Sambal Korek', 'Ayam krispi gurih dengan sambal bawang pedas khas dan lalapan.', 28000, 'fa-bowl-food', 'bg-amber-50', 'Best', true, 6),
            (b_id, cat_meals, 'Nasi Goreng Spesial Mareme', 'Nasi goreng bumbu rempah dengan suwiran ayam, telur, & kerupuk.', 30000, 'fa-bowl-rice', 'bg-amber-50', 'Favorit', true, 7),
            (b_id, cat_meals, 'Mie Nyemek Spesial', 'Mie kuah kental gurih dengan telur orak-arik, bakso, dan sayuran.', 26000, 'fa-bowl-food', 'bg-amber-50', NULL, true, 8),
            (b_id, cat_meals, 'Ayam Bakar Madu Solo', 'Ayam ungkep bumbu rempah manis gurih dipanggang arang harum.', 32000, 'fa-utensils', 'bg-amber-50', NULL, true, 9),
            (b_id, cat_meals, 'Sop Iga Sapi Gurih', 'Iga sapi empuk dengan kuah kaldu rempah bening hangat menyegarkan.', 45000, 'fa-bowl-food', 'bg-amber-50', 'Pilihan Chef', true, 10);

        -- Pastry & Cemilan
        INSERT INTO menu_items (branch_id, category_id, name, description, price, icon, icon_bg_class, tag, is_available, sort_order)
        VALUES
            (b_id, cat_pastry, 'Tahu Cabe Garam', 'Tahu sutra krispi ditumis cabai merah, bawang, dan garam gurih.', 18000, 'fa-bowl-food', 'bg-amber-50', NULL, true, 11),
            (b_id, cat_pastry, 'Tempe Mendoan Hangat', 'Tempe kedelai berbalut tepung daun bawang disajikan sambal kecap.', 16000, 'fa-bread-slice', 'bg-amber-50', 'Favorit', true, 12),
            (b_id, cat_pastry, 'Pisang Goreng Keju Aren', 'Pisang kepok manis renyah dengan taburan keju parut & saus aren.', 20000, 'fa-cookie-bite', 'bg-amber-50', NULL, true, 13),
            (b_id, cat_pastry, 'French Fries Truffle', 'Kentang goreng renyah dengan aroma truffle dan taburan parmesan.', 22000, 'fa-bowl-food', 'bg-stone-100', NULL, true, 14);

        -- Penutup / Dessert
        INSERT INTO menu_items (branch_id, category_id, name, description, price, icon, icon_bg_class, tag, is_available, sort_order)
        VALUES
            (b_id, cat_dessert, 'Es Campur Spesial Solo', 'Aneka jelly, kelapa muda, nangka, alpukat dengan sirup kental.', 22000, 'fa-ice-cream', 'bg-pink-50', 'Segar', true, 15),
            (b_id, cat_dessert, 'Roti Bakar Cokelat Keju', 'Roti tebal bakar mentega dengan lumeran cokelat dan keju cheddar.', 24000, 'fa-bread-slice', 'bg-amber-50', NULL, true, 16),
            (b_id, cat_dessert, 'Affogato Es Krim Vanilla', 'Dua skup es krim vanilla disiram double shot espresso pekat.', 25000, 'fa-mug-hot', 'bg-stone-100', NULL, true, 17);
    END LOOP;
END $$;
