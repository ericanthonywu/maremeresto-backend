-- 000014_seed_makamhaji_menu.up.sql
-- Seed Mareme Makamhaji (branch 22222222-2222-2222-2222-222222222222) menu
-- from the official printed menu (Jl. Slamet Riyadi No. 456 Makamhaji Sukoharjo).

-- Ensure Mareme Makamhaji contact details & operating hours match the printed menu
UPDATE branches
SET
    address = 'Jl. Slamet Riyadi No. 456 Makamhaji Sukoharjo',
    phone = '08967-397-0088',
    updated_at = NOW()
WHERE id = '22222222-2222-2222-2222-222222222222' OR slug = 'makamhaji';

UPDATE branch_settings
SET
    whatsapp_number = '089673970088',
    operating_hours = '{"weekday": {"open": "11:00", "close": "21:00"}, "weekend": {"open": "11:00", "close": "21:00"}}'::jsonb,
    updated_at = NOW()
WHERE branch_id = '22222222-2222-2222-2222-222222222222';

-- Ensure all menu categories exist
INSERT INTO categories (id, name, slug, emoji, sort_order)
VALUES
    ('c1111111-1111-1111-1111-111111111111', 'Aneka Nasi Goreng',   'aneka-nasi-goreng',   '🍛', 1),
    ('c2222222-2222-2222-2222-222222222222', 'Masakan Daging Ayam', 'masakan-daging-ayam', '🍗', 2),
    ('c3333333-3333-3333-3333-333333333333', 'Aneka Masakan',       'aneka-masakan',       '🍜', 3),
    ('c4444444-4444-4444-4444-444444444444', 'Aneka Soup',          'aneka-soup',          '🍲', 4),
    ('c5555555-5555-5555-5555-555555555555', 'Masakan Telur Asin',  'masakan-telur-asin',  '🥚', 5),
    ('c6666666-6666-6666-6666-666666666666', 'Masakan Gurame',      'masakan-gurame',      '🐟', 6),
    ('c7777777-7777-7777-7777-777777777777', 'Masakan Udang',       'masakan-udang',       '🦐', 7),
    ('c8888888-8888-8888-8888-888888888888', 'Masakan Daging Sapi', 'masakan-daging-sapi', '🥩', 8),
    ('c9999999-9999-9999-9999-999999999999', 'Masakan Cumi',        'masakan-cumi',        '🦑', 9),
    ('caaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'Minuman',             'minuman',             '🥤', 10)
ON CONFLICT (slug) DO UPDATE SET
    name = EXCLUDED.name,
    emoji = EXCLUDED.emoji,
    sort_order = EXCLUDED.sort_order;

-- Replace existing Makamhaji menu items cleanly
DELETE FROM menu_items WHERE branch_id = '22222222-2222-2222-2222-222222222222';

DO $$
DECLARE
    b_makamhaji   UUID := '22222222-2222-2222-2222-222222222222';
    cat_nasgor    UUID := 'c1111111-1111-1111-1111-111111111111';
    cat_ayam      UUID := 'c2222222-2222-2222-2222-222222222222';
    cat_masakan   UUID := 'c3333333-3333-3333-3333-333333333333';
    cat_soup      UUID := 'c4444444-4444-4444-4444-444444444444';
    cat_gurame    UUID := 'c6666666-6666-6666-6666-666666666666';
    cat_udang     UUID := 'c7777777-7777-7777-7777-777777777777';
    cat_sapi      UUID := 'c8888888-8888-8888-8888-888888888888';
    cat_cumi      UUID := 'c9999999-9999-9999-9999-999999999999';
    cat_minuman   UUID := 'caaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa';
BEGIN
    -- =========================================================================
    -- OUTLET: MAREME MAKAMHAJI (116 items)
    -- =========================================================================

    -- 1. ANEKA NASI GORENG (21 items)
    INSERT INTO menu_items (branch_id, category_id, name, description, price, icon, icon_bg_class, tag, is_available, sort_order)
    VALUES
        (b_makamhaji, cat_nasgor, 'Nasi Goreng Ijo',             '', 25500, 'fa-bowl-rice', 'bg-emerald-50', NULL, true, 1),
        (b_makamhaji, cat_nasgor, 'Nasi Goreng Ijo Istimewa',    '', 27000, 'fa-bowl-rice', 'bg-emerald-50', NULL, true, 2),
        (b_makamhaji, cat_nasgor, 'Nasi Goreng Ijo Udang',       '', 26500, 'fa-bowl-rice', 'bg-emerald-50', NULL, true, 3),
        (b_makamhaji, cat_nasgor, 'Nasi Goreng Ijo Udang Ist',   '', 28000, 'fa-bowl-rice', 'bg-emerald-50', NULL, true, 4),
        (b_makamhaji, cat_nasgor, 'Nasi Goreng Ijo Seafood',     '', 27500, 'fa-bowl-rice', 'bg-emerald-50', NULL, true, 5),
        (b_makamhaji, cat_nasgor, 'Nasi Goreng Ijo Seafood Ist', '', 29000, 'fa-bowl-rice', 'bg-emerald-50', NULL, true, 6),
        (b_makamhaji, cat_nasgor, 'Nasi Goreng Sosis',           '', 25500, 'fa-bowl-rice', 'bg-amber-50',   NULL, true, 7),
        (b_makamhaji, cat_nasgor, 'Nasi Goreng Bakso',           '', 25500, 'fa-bowl-rice', 'bg-amber-50',   NULL, true, 8),
        (b_makamhaji, cat_nasgor, 'Nasi Goreng Mawut',           '', 29000, 'fa-bowl-rice', 'bg-amber-50',   NULL, true, 9),
        (b_makamhaji, cat_nasgor, 'Nasi Goreng Ayam',            '', 25500, 'fa-bowl-rice', 'bg-amber-50',   NULL, true, 10),
        (b_makamhaji, cat_nasgor, 'Nasi Goreng Ayam Ist',        '', 27000, 'fa-bowl-rice', 'bg-amber-50',   NULL, true, 11),
        (b_makamhaji, cat_nasgor, 'Nasi Goreng Babat',           '', 25500, 'fa-bowl-rice', 'bg-amber-50',   NULL, true, 12),
        (b_makamhaji, cat_nasgor, 'Nasi Goreng Babat Ist',       '', 27000, 'fa-bowl-rice', 'bg-amber-50',   NULL, true, 13),
        (b_makamhaji, cat_nasgor, 'Nasi Goreng Seafood',         '', 27500, 'fa-bowl-rice', 'bg-amber-50',   NULL, true, 14),
        (b_makamhaji, cat_nasgor, 'Nasi Goreng Seafood Ist',     '', 29000, 'fa-bowl-rice', 'bg-amber-50',   NULL, true, 15),
        (b_makamhaji, cat_nasgor, 'Nasi Goreng Udang',           '', 26500, 'fa-bowl-rice', 'bg-amber-50',   NULL, true, 16),
        (b_makamhaji, cat_nasgor, 'Nasi Goreng Udang Ist',       '', 28000, 'fa-bowl-rice', 'bg-amber-50',   NULL, true, 17),
        (b_makamhaji, cat_nasgor, 'Nasi Goreng Sapi',            '', 26500, 'fa-bowl-rice', 'bg-amber-50',   NULL, true, 18),
        (b_makamhaji, cat_nasgor, 'Nasi Goreng Sapi Ist',        '', 28000, 'fa-bowl-rice', 'bg-amber-50',   NULL, true, 19),
        (b_makamhaji, cat_nasgor, 'Nasi Goreng Lada Hitam',      '', 27000, 'fa-bowl-rice', 'bg-stone-100',  NULL, true, 20),
        (b_makamhaji, cat_nasgor, 'Nasi Goreng Lada Hitam Ist',  '', 28500, 'fa-bowl-rice', 'bg-stone-100',  NULL, true, 21);

    -- 2. MASAKAN DAGING AYAM (20 items)
    INSERT INTO menu_items (branch_id, category_id, name, description, price, icon, icon_bg_class, tag, is_available, sort_order)
    VALUES
        (b_makamhaji, cat_ayam, 'Ayam Gongso',             '', 31000, 'fa-drumstick-bite', 'bg-amber-50',   NULL, true, 22),
        (b_makamhaji, cat_ayam, 'Ayam Lada Hitam',         '', 31000, 'fa-drumstick-bite', 'bg-amber-50',   NULL, true, 23),
        (b_makamhaji, cat_ayam, 'Ayam Masak Kecap Pedas',  '', 31000, 'fa-drumstick-bite', 'bg-amber-50',   NULL, true, 24),
        (b_makamhaji, cat_ayam, 'Ayam Goreng Mareme',      '', 31000, 'fa-drumstick-bite', 'bg-amber-50',   NULL, true, 25),
        (b_makamhaji, cat_ayam, 'Ayam Goreng Asam Manis',  '', 31000, 'fa-drumstick-bite', 'bg-amber-50',   NULL, true, 26),
        (b_makamhaji, cat_ayam, 'Ayam Goreng Mentega',     '', 31000, 'fa-drumstick-bite', 'bg-amber-50',   NULL, true, 27),
        (b_makamhaji, cat_ayam, 'Ayam Goreng Tepung',      '', 31000, 'fa-drumstick-bite', 'bg-amber-50',   NULL, true, 28),
        (b_makamhaji, cat_ayam, 'Ayam Goreng Masak Tauco', '', 31000, 'fa-drumstick-bite', 'bg-amber-50',   NULL, true, 29),
        (b_makamhaji, cat_ayam, 'Ayam Ca Jamur',           '', 31000, 'fa-drumstick-bite', 'bg-amber-50',   NULL, true, 30),
        (b_makamhaji, cat_ayam, 'Ayam Ca Kangkung',        '', 31000, 'fa-drumstick-bite', 'bg-emerald-50', NULL, true, 31),
        (b_makamhaji, cat_ayam, 'Ayam Ca Brokoli',         '', 31000, 'fa-drumstick-bite', 'bg-emerald-50', NULL, true, 32),
        (b_makamhaji, cat_ayam, 'Ayam Sayur Asin Ca/Kuah', '', 31000, 'fa-drumstick-bite', 'bg-amber-50',   NULL, true, 33),
        (b_makamhaji, cat_ayam, 'Ayam Nanking',            '', 31000, 'fa-drumstick-bite', 'bg-amber-50',   NULL, true, 34),
        (b_makamhaji, cat_ayam, 'Ayam Rebus',              '', 31000, 'fa-drumstick-bite', 'bg-stone-100',  NULL, true, 35),
        (b_makamhaji, cat_ayam, 'Sayap Masak Kecap',       '', 31000, 'fa-drumstick-bite', 'bg-amber-50',   NULL, true, 36),
        (b_makamhaji, cat_ayam, 'Bestik Ayam',             '', 31000, 'fa-drumstick-bite', 'bg-amber-50',   NULL, true, 37),
        (b_makamhaji, cat_ayam, 'Ko Loo Kee',              '', 31000, 'fa-drumstick-bite', 'bg-amber-50',   NULL, true, 38),
        (b_makamhaji, cat_ayam, 'Ngo Yang Kee',            '', 31000, 'fa-drumstick-bite', 'bg-amber-50',   NULL, true, 39),
        (b_makamhaji, cat_ayam, 'Pak Lay Ca',              '', 31000, 'fa-bowl-food',      'bg-amber-50',   NULL, true, 40),
        (b_makamhaji, cat_ayam, 'Kulit Goreng',            '', 31000, 'fa-drumstick-bite', 'bg-amber-50',   NULL, true, 41);

    -- 3. ANEKA MASAKAN (26 items)
    INSERT INTO menu_items (branch_id, category_id, name, description, price, icon, icon_bg_class, tag, is_available, sort_order)
    VALUES
        (b_makamhaji, cat_masakan, 'Bakmi Goreng',           '', 25000, 'fa-bowl-food', 'bg-amber-50',   NULL, true, 42),
        (b_makamhaji, cat_masakan, 'Bakmi Godog',            '', 25000, 'fa-bowl-food', 'bg-amber-50',   NULL, true, 43),
        (b_makamhaji, cat_masakan, 'Bihun Goreng',           '', 25000, 'fa-bowl-food', 'bg-amber-50',   NULL, true, 44),
        (b_makamhaji, cat_masakan, 'Bihun Godog',            '', 25000, 'fa-bowl-food', 'bg-amber-50',   NULL, true, 45),
        (b_makamhaji, cat_masakan, 'Capjay Goreng',          '', 26500, 'fa-bowl-food', 'bg-emerald-50', NULL, true, 46),
        (b_makamhaji, cat_masakan, 'Capjay Godog',           '', 26500, 'fa-bowl-food', 'bg-emerald-50', NULL, true, 47),
        (b_makamhaji, cat_masakan, 'Nasi Ayam Ca',           '', 27500, 'fa-bowl-rice', 'bg-amber-50',   NULL, true, 48),
        (b_makamhaji, cat_masakan, 'Nasi Campur',            '', 28000, 'fa-bowl-rice', 'bg-amber-50',   NULL, true, 49),
        (b_makamhaji, cat_masakan, 'Nasi Putih',             '', 5000,  'fa-bowl-rice', 'bg-stone-100',  NULL, true, 50),
        (b_makamhaji, cat_masakan, 'Masak Tahu',             '', 28000, 'fa-bowl-food', 'bg-amber-50',   NULL, true, 51),
        (b_makamhaji, cat_masakan, 'Tahu Masak Kecap',       '', 28500, 'fa-bowl-food', 'bg-amber-50',   NULL, true, 52),
        (b_makamhaji, cat_masakan, 'Misoa Kuah',             '', 27500, 'fa-bowl-food', 'bg-amber-50',   NULL, true, 53),
        (b_makamhaji, cat_masakan, 'Mun Tahu',               '', 28500, 'fa-bowl-food', 'bg-amber-50',   NULL, true, 54),
        (b_makamhaji, cat_masakan, 'Angsio Tahu',            '', 28500, 'fa-bowl-food', 'bg-amber-50',   NULL, true, 55),
        (b_makamhaji, cat_masakan, 'Sapo Tahu',              '', 33000, 'fa-bowl-food', 'bg-amber-50',   NULL, true, 56),
        (b_makamhaji, cat_masakan, 'Soon Goreng',            '', 27000, 'fa-bowl-food', 'bg-amber-50',   NULL, true, 57),
        (b_makamhaji, cat_masakan, 'Ta Mie Goreng',          '', 26500, 'fa-bowl-food', 'bg-amber-50',   NULL, true, 58),
        (b_makamhaji, cat_masakan, 'Aneka Sayur',            '', 24000, 'fa-leaf',      'bg-emerald-50', NULL, true, 59),
        (b_makamhaji, cat_masakan, 'Fuyung Hay',             '', 27500, 'fa-bowl-food', 'bg-amber-50',   NULL, true, 60),
        (b_makamhaji, cat_masakan, 'Kangkung Ca Polos',      '', 24000, 'fa-leaf',      'bg-emerald-50', NULL, true, 61),
        (b_makamhaji, cat_masakan, 'Kwee Tiauw Goreng/Kuah', '', 25500, 'fa-bowl-food', 'bg-amber-50',   NULL, true, 62),
        (b_makamhaji, cat_masakan, 'Kwee Tiauw Nyemek',      '', 25500, 'fa-bowl-food', 'bg-amber-50',   NULL, true, 63),
        (b_makamhaji, cat_masakan, 'Kamar Bola',             '', 27000, 'fa-bowl-food', 'bg-amber-50',   NULL, true, 64),
        (b_makamhaji, cat_masakan, 'Brokoli Polos',          '', 24500, 'fa-leaf',      'bg-emerald-50', NULL, true, 65),
        (b_makamhaji, cat_masakan, 'Babat Gongso',           '', 28500, 'fa-bowl-food', 'bg-amber-50',   NULL, true, 66),
        (b_makamhaji, cat_masakan, 'Nasi Capjay',            '', 25500, 'fa-bowl-rice', 'bg-emerald-50', NULL, true, 67);

    -- 4. ANEKA SOUP (5 items)
    INSERT INTO menu_items (branch_id, category_id, name, description, price, icon, icon_bg_class, tag, is_available, sort_order)
    VALUES
        (b_makamhaji, cat_soup, 'Soup Asparagus Kepiting', '', 43000, 'fa-bowl-food', 'bg-amber-50', NULL, true, 68),
        (b_makamhaji, cat_soup, 'Soup ShangHay',           '', 34000, 'fa-bowl-food', 'bg-amber-50', NULL, true, 69),
        (b_makamhaji, cat_soup, 'Soup Ayam',               '', 31000, 'fa-bowl-food', 'bg-amber-50', NULL, true, 70),
        (b_makamhaji, cat_soup, 'Soup Tahu Sayur Asin',    '', 31000, 'fa-bowl-food', 'bg-amber-50', NULL, true, 71),
        (b_makamhaji, cat_soup, 'Soup Sapi',               '', 31000, 'fa-bowl-food', 'bg-amber-50', NULL, true, 72);

    -- 5. MASAKAN GURAME (8 items)
    INSERT INTO menu_items (branch_id, category_id, name, description, price, icon, icon_bg_class, tag, is_available, sort_order)
    VALUES
        (b_makamhaji, cat_gurame, 'Soup Gurame Kuah Sayur Asin', '', 35000, 'fa-fish', 'bg-amber-50', NULL, true, 73),
        (b_makamhaji, cat_gurame, 'Gurame Goreng Mareme',        '', 31000, 'fa-fish', 'bg-amber-50', NULL, true, 74),
        (b_makamhaji, cat_gurame, 'Gurame Goreng Asam Manis',    '', 31000, 'fa-fish', 'bg-amber-50', NULL, true, 75),
        (b_makamhaji, cat_gurame, 'Gurame Goreng Mentega',       '', 31000, 'fa-fish', 'bg-amber-50', NULL, true, 76),
        (b_makamhaji, cat_gurame, 'Gurame Goreng Tepung',        '', 31000, 'fa-fish', 'bg-amber-50', NULL, true, 77),
        (b_makamhaji, cat_gurame, 'Gurame Masak Tauco',          '', 31000, 'fa-fish', 'bg-amber-50', NULL, true, 78),
        (b_makamhaji, cat_gurame, 'Gurame Masak Kecap',          '', 31000, 'fa-fish', 'bg-amber-50', NULL, true, 79),
        (b_makamhaji, cat_gurame, 'Gurame Gongso',               '', 31000, 'fa-fish', 'bg-amber-50', NULL, true, 80);

    -- 6. MASAKAN UDANG (10 items)
    INSERT INTO menu_items (branch_id, category_id, name, description, price, icon, icon_bg_class, tag, is_available, sort_order)
    VALUES
        (b_makamhaji, cat_udang, 'Udang Bago Spesial',      '', 35000, 'fa-shrimp', 'bg-amber-50',   NULL, true, 81),
        (b_makamhaji, cat_udang, 'Udang Bago Mentega',      '', 33000, 'fa-shrimp', 'bg-amber-50',   NULL, true, 82),
        (b_makamhaji, cat_udang, 'Udang Goreng Mareme',     '', 31000, 'fa-shrimp', 'bg-amber-50',   NULL, true, 83),
        (b_makamhaji, cat_udang, 'Udang Goreng Asam Manis', '', 31000, 'fa-shrimp', 'bg-amber-50',   NULL, true, 84),
        (b_makamhaji, cat_udang, 'Udang Goreng Mentega',    '', 31000, 'fa-shrimp', 'bg-amber-50',   NULL, true, 85),
        (b_makamhaji, cat_udang, 'Udang Goreng Tepung',     '', 31000, 'fa-shrimp', 'bg-amber-50',   NULL, true, 86),
        (b_makamhaji, cat_udang, 'Udang Gr Masak Tauco',    '', 31000, 'fa-shrimp', 'bg-amber-50',   NULL, true, 87),
        (b_makamhaji, cat_udang, 'Udang Ca Sayur',          '', 31000, 'fa-shrimp', 'bg-emerald-50', NULL, true, 88),
        (b_makamhaji, cat_udang, 'Udang Ca Jamur',          '', 31000, 'fa-shrimp', 'bg-amber-50',   NULL, true, 89),
        (b_makamhaji, cat_udang, 'Udang Ca Kangkung',       '', 31000, 'fa-shrimp', 'bg-emerald-50', NULL, true, 90);

    -- 7. MASAKAN DAGING SAPI (6 items)
    INSERT INTO menu_items (branch_id, category_id, name, description, price, icon, icon_bg_class, tag, is_available, sort_order)
    VALUES
        (b_makamhaji, cat_sapi, 'Sapi Ca Sayur',    '', 31000, 'fa-bowl-food', 'bg-emerald-50', NULL, true, 91),
        (b_makamhaji, cat_sapi, 'Sapi Ca Jamur',    '', 31000, 'fa-bowl-food', 'bg-amber-50',   NULL, true, 92),
        (b_makamhaji, cat_sapi, 'Sapi Ca Kangkung', '', 31000, 'fa-bowl-food', 'bg-emerald-50', NULL, true, 93),
        (b_makamhaji, cat_sapi, 'Sapi Ca Brokoli',  '', 31000, 'fa-bowl-food', 'bg-emerald-50', NULL, true, 94),
        (b_makamhaji, cat_sapi, 'Sapi Lada Hitam',  '', 31000, 'fa-bowl-food', 'bg-stone-100',  NULL, true, 95),
        (b_makamhaji, cat_sapi, 'Bestik Sapi',      '', 31000, 'fa-bowl-food', 'bg-amber-50',   NULL, true, 96);

    -- 8. MASAKAN CUMI (6 items)
    INSERT INTO menu_items (branch_id, category_id, name, description, price, icon, icon_bg_class, tag, is_available, sort_order)
    VALUES
        (b_makamhaji, cat_cumi, 'Cumi Goreng Mareme',     '', 31000, 'fa-bowl-food', 'bg-amber-50',   NULL, true, 97),
        (b_makamhaji, cat_cumi, 'Cumi Goreng Asam Manis', '', 31000, 'fa-bowl-food', 'bg-amber-50',   NULL, true, 98),
        (b_makamhaji, cat_cumi, 'Cumi Goreng Mentega',    '', 31000, 'fa-bowl-food', 'bg-amber-50',   NULL, true, 99),
        (b_makamhaji, cat_cumi, 'Cumi Goreng Tepung',     '', 31000, 'fa-bowl-food', 'bg-amber-50',   NULL, true, 100),
        (b_makamhaji, cat_cumi, 'Cumi Ca Sayur',          '', 31000, 'fa-bowl-food', 'bg-emerald-50', NULL, true, 101),
        (b_makamhaji, cat_cumi, 'Cumi Ca Jamur',          '', 31000, 'fa-bowl-food', 'bg-amber-50',   NULL, true, 102);

    -- 9. MINUMAN (14 items)
    INSERT INTO menu_items (branch_id, category_id, name, description, price, icon, icon_bg_class, tag, is_available, sort_order)
    VALUES
        (b_makamhaji, cat_minuman, 'Es Pelangi',       '', 8000,  'fa-glass-water', 'bg-pink-50',    NULL, true, 103),
        (b_makamhaji, cat_minuman, 'Lemon Tea',        '', 6500,  'fa-mug-hot',     'bg-amber-50',   NULL, true, 104),
        (b_makamhaji, cat_minuman, 'Wedang Teh Tawar', '', 3000,  'fa-mug-hot',     'bg-amber-50',   NULL, true, 105),
        (b_makamhaji, cat_minuman, 'Wedang Teh Manis', '', 4000,  'fa-mug-hot',     'bg-amber-50',   NULL, true, 106),
        (b_makamhaji, cat_minuman, 'Wedang Jeruk',     '', 6500,  'fa-mug-hot',     'bg-amber-50',   NULL, true, 107),
        (b_makamhaji, cat_minuman, 'Es Teh Tawar',     '', 3000,  'fa-glass-water', 'bg-amber-50',   NULL, true, 108),
        (b_makamhaji, cat_minuman, 'Es Teh Manis',     '', 4000,  'fa-glass-water', 'bg-amber-50',   NULL, true, 109),
        (b_makamhaji, cat_minuman, 'Jeruk Murni',      '', 25000, 'fa-glass-water', 'bg-amber-50',   NULL, true, 110),
        (b_makamhaji, cat_minuman, 'Es Hijau',         '', 8000,  'fa-glass-water', 'bg-emerald-50', NULL, true, 111),
        (b_makamhaji, cat_minuman, 'Es Jeruk',         '', 7000,  'fa-glass-water', 'bg-amber-50',   NULL, true, 112),
        (b_makamhaji, cat_minuman, 'Es Lemon Tea',     '', 6500,  'fa-glass-water', 'bg-amber-50',   NULL, true, 113),
        (b_makamhaji, cat_minuman, 'Es Soda Gembira',  '', 12000, 'fa-glass-water', 'bg-pink-50',    NULL, true, 114),
        (b_makamhaji, cat_minuman, 'Es Batu',          '', 1000,  'fa-glass-water', 'bg-stone-100',  NULL, true, 115),
        (b_makamhaji, cat_minuman, 'Air Mineral',      '', 5000,  'fa-glass-water', 'bg-stone-100',  NULL, true, 116);
END $$;
