-- 000013_seed_samratulangi_menu.up.sql
-- Clear all menu items and categories, update outlet names/contacts, and seed:
-- 1. Mareme Kerten / Samratulangi (branch 11111111-1111-1111-1111-111111111111)
-- 2. Mak Djan (branch 33333333-3333-3333-3333-333333333333)
-- No improvised descriptions or tags.

DELETE FROM menu_items;
DELETE FROM categories;

-- Rename Mareme Slamet Riyadi back to Mak Djan
UPDATE branches
SET name = 'Mak Djan', updated_at = NOW()
WHERE id = '33333333-3333-3333-3333-333333333333' OR slug = 'mak-djan';

-- Update Mareme Kerten contact details from the printed menu
UPDATE branches
SET phone = '(0271) 712452', updated_at = NOW()
WHERE id = '11111111-1111-1111-1111-111111111111' AND (phone IS NULL OR phone = '');

UPDATE branch_settings
SET whatsapp_number = '085867222301', updated_at = NOW()
WHERE branch_id = '11111111-1111-1111-1111-111111111111' AND (whatsapp_number IS NULL OR whatsapp_number = '');

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

DO $$
DECLARE
    b_kerten      UUID := '11111111-1111-1111-1111-111111111111';
    b_makdjan     UUID := '33333333-3333-3333-3333-333333333333';
    cat_nasgor    UUID := 'c1111111-1111-1111-1111-111111111111';
    cat_ayam      UUID := 'c2222222-2222-2222-2222-222222222222';
    cat_masakan   UUID := 'c3333333-3333-3333-3333-333333333333';
    cat_soup      UUID := 'c4444444-4444-4444-4444-444444444444';
    cat_telurasin UUID := 'c5555555-5555-5555-5555-555555555555';
    cat_gurame    UUID := 'c6666666-6666-6666-6666-666666666666';
    cat_udang     UUID := 'c7777777-7777-7777-7777-777777777777';
    cat_sapi      UUID := 'c8888888-8888-8888-8888-888888888888';
    cat_cumi      UUID := 'c9999999-9999-9999-9999-999999999999';
    cat_minuman   UUID := 'caaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa';
BEGIN
    -- =========================================================================
    -- OUTLET 1: MAREME KERTEN (138 items)
    -- =========================================================================

    -- 1. ANEKA NASI GORENG (23 items)
    INSERT INTO menu_items (branch_id, category_id, name, description, price, icon, icon_bg_class, tag, is_available, sort_order)
    VALUES
        (b_kerten, cat_nasgor, 'Nasi Goreng Ijo',             '', 25500, 'fa-bowl-rice', 'bg-emerald-50', NULL, true, 1),
        (b_kerten, cat_nasgor, 'Nasi Goreng Ijo Istimewa',    '', 27000, 'fa-bowl-rice', 'bg-emerald-50', NULL, true, 2),
        (b_kerten, cat_nasgor, 'Nasi Goreng Ijo Udang',       '', 26500, 'fa-bowl-rice', 'bg-emerald-50', NULL, true, 3),
        (b_kerten, cat_nasgor, 'Nasi Goreng Ijo Udang Ist',   '', 28000, 'fa-bowl-rice', 'bg-emerald-50', NULL, true, 4),
        (b_kerten, cat_nasgor, 'Nasi Goreng Ijo Seafood',     '', 27500, 'fa-bowl-rice', 'bg-emerald-50', NULL, true, 5),
        (b_kerten, cat_nasgor, 'Nasi Goreng Ijo Seafood Ist', '', 29000, 'fa-bowl-rice', 'bg-emerald-50', NULL, true, 6),
        (b_kerten, cat_nasgor, 'Nasi Goreng Sosis',           '', 25500, 'fa-bowl-rice', 'bg-amber-50',   NULL, true, 7),
        (b_kerten, cat_nasgor, 'Nasi Goreng Bakso',           '', 25500, 'fa-bowl-rice', 'bg-amber-50',   NULL, true, 8),
        (b_kerten, cat_nasgor, 'Nasi Goreng Mawut',           '', 29000, 'fa-bowl-rice', 'bg-amber-50',   NULL, true, 9),
        (b_kerten, cat_nasgor, 'Nasi Goreng Ayam',            '', 25500, 'fa-bowl-rice', 'bg-amber-50',   NULL, true, 10),
        (b_kerten, cat_nasgor, 'Nasi Goreng Ayam Ist',        '', 27000, 'fa-bowl-rice', 'bg-amber-50',   NULL, true, 11),
        (b_kerten, cat_nasgor, 'Nasi Goreng Kambing',         '', 25500, 'fa-bowl-rice', 'bg-amber-50',   NULL, true, 12),
        (b_kerten, cat_nasgor, 'Nasi Goreng Kambing Ist',     '', 27000, 'fa-bowl-rice', 'bg-amber-50',   NULL, true, 13),
        (b_kerten, cat_nasgor, 'Nasi Goreng Babat',           '', 25500, 'fa-bowl-rice', 'bg-amber-50',   NULL, true, 14),
        (b_kerten, cat_nasgor, 'Nasi Goreng Babat Ist',       '', 27000, 'fa-bowl-rice', 'bg-amber-50',   NULL, true, 15),
        (b_kerten, cat_nasgor, 'Nasi Goreng Seafood',         '', 27500, 'fa-bowl-rice', 'bg-amber-50',   NULL, true, 16),
        (b_kerten, cat_nasgor, 'Nasi Goreng Seafood Ist',     '', 29000, 'fa-bowl-rice', 'bg-amber-50',   NULL, true, 17),
        (b_kerten, cat_nasgor, 'Nasi Goreng Udang',           '', 26500, 'fa-bowl-rice', 'bg-amber-50',   NULL, true, 18),
        (b_kerten, cat_nasgor, 'Nasi Goreng Udang Ist',       '', 28000, 'fa-bowl-rice', 'bg-amber-50',   NULL, true, 19),
        (b_kerten, cat_nasgor, 'Nasi Goreng Sapi',            '', 26500, 'fa-bowl-rice', 'bg-amber-50',   NULL, true, 20),
        (b_kerten, cat_nasgor, 'Nasi Goreng Sapi Ist',        '', 28000, 'fa-bowl-rice', 'bg-amber-50',   NULL, true, 21),
        (b_kerten, cat_nasgor, 'Nasi Goreng Lada Hitam',      '', 27000, 'fa-bowl-rice', 'bg-stone-100',  NULL, true, 22),
        (b_kerten, cat_nasgor, 'Nasi Goreng Lada Hitam Ist',  '', 28500, 'fa-bowl-rice', 'bg-stone-100',  NULL, true, 23);

    -- 2. MASAKAN DAGING AYAM (22 items)
    INSERT INTO menu_items (branch_id, category_id, name, description, price, icon, icon_bg_class, tag, is_available, sort_order)
    VALUES
        (b_kerten, cat_ayam, 'Ayam Teriyaki',           '', 30500, 'fa-drumstick-bite', 'bg-amber-50',   NULL, true, 24),
        (b_kerten, cat_ayam, 'Ayam Gongso',             '', 30500, 'fa-drumstick-bite', 'bg-amber-50',   NULL, true, 25),
        (b_kerten, cat_ayam, 'Ayam Lada Hitam',         '', 30500, 'fa-drumstick-bite', 'bg-amber-50',   NULL, true, 26),
        (b_kerten, cat_ayam, 'Ayam Masak Kecap Pedas',  '', 30500, 'fa-drumstick-bite', 'bg-amber-50',   NULL, true, 27),
        (b_kerten, cat_ayam, 'Ayam Goreng Mareme',      '', 30500, 'fa-drumstick-bite', 'bg-amber-50',   NULL, true, 28),
        (b_kerten, cat_ayam, 'Ayam Goreng Asam Manis',  '', 30500, 'fa-drumstick-bite', 'bg-amber-50',   NULL, true, 29),
        (b_kerten, cat_ayam, 'Ayam Goreng Mentega',     '', 30500, 'fa-drumstick-bite', 'bg-amber-50',   NULL, true, 30),
        (b_kerten, cat_ayam, 'Ayam Goreng Tepung',      '', 30500, 'fa-drumstick-bite', 'bg-amber-50',   NULL, true, 31),
        (b_kerten, cat_ayam, 'Ayam Goreng Masak Tauco', '', 30500, 'fa-drumstick-bite', 'bg-amber-50',   NULL, true, 32),
        (b_kerten, cat_ayam, 'Ayam Ca Jamur',           '', 30500, 'fa-drumstick-bite', 'bg-amber-50',   NULL, true, 33),
        (b_kerten, cat_ayam, 'Ayam Ca Kaylan',          '', 30500, 'fa-drumstick-bite', 'bg-emerald-50', NULL, true, 34),
        (b_kerten, cat_ayam, 'Ayam Ca Kangkung',        '', 30500, 'fa-drumstick-bite', 'bg-emerald-50', NULL, true, 35),
        (b_kerten, cat_ayam, 'Ayam Ca Brokoli',         '', 30500, 'fa-drumstick-bite', 'bg-emerald-50', NULL, true, 36),
        (b_kerten, cat_ayam, 'Ayam Sayur Asin Ca/Kuah', '', 30500, 'fa-drumstick-bite', 'bg-amber-50',   NULL, true, 37),
        (b_kerten, cat_ayam, 'Ayam Nanking',            '', 30500, 'fa-drumstick-bite', 'bg-amber-50',   NULL, true, 38),
        (b_kerten, cat_ayam, 'Ayam Rebus',              '', 30500, 'fa-drumstick-bite', 'bg-stone-100',  NULL, true, 39),
        (b_kerten, cat_ayam, 'Sayap Masak Kecap',       '', 30500, 'fa-drumstick-bite', 'bg-amber-50',   NULL, true, 40),
        (b_kerten, cat_ayam, 'Bestik Ayam',             '', 30500, 'fa-drumstick-bite', 'bg-amber-50',   NULL, true, 41),
        (b_kerten, cat_ayam, 'Ko Loo Kee',              '', 30500, 'fa-drumstick-bite', 'bg-amber-50',   NULL, true, 42),
        (b_kerten, cat_ayam, 'Ngo Yang Kee',            '', 30500, 'fa-drumstick-bite', 'bg-amber-50',   NULL, true, 43),
        (b_kerten, cat_ayam, 'Pak Lay Ca',              '', 30000, 'fa-bowl-food',      'bg-amber-50',   NULL, true, 44),
        (b_kerten, cat_ayam, 'Kulit Goreng',            '', 30000, 'fa-drumstick-bite', 'bg-amber-50',   NULL, true, 45);

    -- 3. ANEKA MASAKAN (30 items)
    INSERT INTO menu_items (branch_id, category_id, name, description, price, icon, icon_bg_class, tag, is_available, sort_order)
    VALUES
        (b_kerten, cat_masakan, 'Bakmi Goreng',           '', 25000, 'fa-bowl-food', 'bg-amber-50',   NULL, true, 46),
        (b_kerten, cat_masakan, 'Bakmi Godog',            '', 25000, 'fa-bowl-food', 'bg-amber-50',   NULL, true, 47),
        (b_kerten, cat_masakan, 'Bihun Goreng',           '', 25000, 'fa-bowl-food', 'bg-amber-50',   NULL, true, 48),
        (b_kerten, cat_masakan, 'Bihun Godog',            '', 25000, 'fa-bowl-food', 'bg-amber-50',   NULL, true, 49),
        (b_kerten, cat_masakan, 'Capjay Goreng',          '', 26500, 'fa-bowl-food', 'bg-emerald-50', NULL, true, 50),
        (b_kerten, cat_masakan, 'Capjay Godog',           '', 26500, 'fa-bowl-food', 'bg-emerald-50', NULL, true, 51),
        (b_kerten, cat_masakan, 'Nasi Ayam Ca',           '', 27500, 'fa-bowl-rice', 'bg-amber-50',   NULL, true, 52),
        (b_kerten, cat_masakan, 'Nasi Campur',            '', 28000, 'fa-bowl-rice', 'bg-amber-50',   NULL, true, 53),
        (b_kerten, cat_masakan, 'Nasi Putih',             '', 5000,  'fa-bowl-rice', 'bg-stone-100',  NULL, true, 54),
        (b_kerten, cat_masakan, 'Masak Tahu',             '', 28000, 'fa-bowl-food', 'bg-amber-50',   NULL, true, 55),
        (b_kerten, cat_masakan, 'Tahu Masak Kecap',       '', 28500, 'fa-bowl-food', 'bg-amber-50',   NULL, true, 56),
        (b_kerten, cat_masakan, 'Misoa Kuah',             '', 27500, 'fa-bowl-food', 'bg-amber-50',   NULL, true, 57),
        (b_kerten, cat_masakan, 'Mun Tahu',               '', 28500, 'fa-bowl-food', 'bg-amber-50',   NULL, true, 58),
        (b_kerten, cat_masakan, 'Angsio Tahu',            '', 28500, 'fa-bowl-food', 'bg-amber-50',   NULL, true, 59),
        (b_kerten, cat_masakan, 'Sapo Tahu',              '', 33000, 'fa-bowl-food', 'bg-amber-50',   NULL, true, 60),
        (b_kerten, cat_masakan, 'Soon Goreng',            '', 27000, 'fa-bowl-food', 'bg-amber-50',   NULL, true, 61),
        (b_kerten, cat_masakan, 'Ta Mie Goreng',          '', 26500, 'fa-bowl-food', 'bg-amber-50',   NULL, true, 62),
        (b_kerten, cat_masakan, 'Yam Yien',               '', 27000, 'fa-bowl-food', 'bg-amber-50',   NULL, true, 63),
        (b_kerten, cat_masakan, 'Aneka Sayur',            '', 24000, 'fa-leaf',      'bg-emerald-50', NULL, true, 64),
        (b_kerten, cat_masakan, 'Fuyung Hay',             '', 27500, 'fa-bowl-food', 'bg-amber-50',   NULL, true, 65),
        (b_kerten, cat_masakan, 'Kay Lan Ca Polos',       '', 24000, 'fa-leaf',      'bg-emerald-50', NULL, true, 66),
        (b_kerten, cat_masakan, 'Kangkung Ca Polos',      '', 24000, 'fa-leaf',      'bg-emerald-50', NULL, true, 67),
        (b_kerten, cat_masakan, 'Kwee Tiauw Goreng/Kuah', '', 25500, 'fa-bowl-food', 'bg-amber-50',   NULL, true, 68),
        (b_kerten, cat_masakan, 'Kwee Tiauw Nyemek',      '', 25500, 'fa-bowl-food', 'bg-amber-50',   NULL, true, 69),
        (b_kerten, cat_masakan, 'Kamar Bola',             '', 27000, 'fa-bowl-food', 'bg-amber-50',   NULL, true, 70),
        (b_kerten, cat_masakan, 'Loo Mie',                '', 27000, 'fa-bowl-food', 'bg-amber-50',   NULL, true, 71),
        (b_kerten, cat_masakan, 'Brokoli Polos',          '', 24500, 'fa-leaf',      'bg-emerald-50', NULL, true, 72),
        (b_kerten, cat_masakan, 'Babat Gongso',           '', 28500, 'fa-bowl-food', 'bg-amber-50',   NULL, true, 73),
        (b_kerten, cat_masakan, 'Orak Arik',              '', 28500, 'fa-bowl-food', 'bg-amber-50',   NULL, true, 74),
        (b_kerten, cat_masakan, 'Nasi Capjay',            '', 25500, 'fa-bowl-rice', 'bg-emerald-50', NULL, true, 75);

    -- 4. ANEKA SOUP (5 items)
    INSERT INTO menu_items (branch_id, category_id, name, description, price, icon, icon_bg_class, tag, is_available, sort_order)
    VALUES
        (b_kerten, cat_soup, 'Soup Asparagus Kepiting', '', 43500, 'fa-bowl-food', 'bg-amber-50', NULL, true, 76),
        (b_kerten, cat_soup, 'Soup ShangHay',           '', 34500, 'fa-bowl-food', 'bg-amber-50', NULL, true, 77),
        (b_kerten, cat_soup, 'Soup Ayam',               '', 30500, 'fa-bowl-food', 'bg-amber-50', NULL, true, 78),
        (b_kerten, cat_soup, 'Soup Tahu Sayur Asin',    '', 30500, 'fa-bowl-food', 'bg-amber-50', NULL, true, 79),
        (b_kerten, cat_soup, 'Soup Sapi',               '', 31000, 'fa-bowl-food', 'bg-amber-50', NULL, true, 80);

    -- 5. MASAKAN TELUR ASIN (4 items)
    INSERT INTO menu_items (branch_id, category_id, name, description, price, icon, icon_bg_class, tag, is_available, sort_order)
    VALUES
        (b_kerten, cat_telurasin, 'Ayam Telur Asin',   '', 37500, 'fa-drumstick-bite', 'bg-amber-50', NULL, true, 81),
        (b_kerten, cat_telurasin, 'Cumi Telur Asin',   '', 37500, 'fa-bowl-food',      'bg-amber-50', NULL, true, 82),
        (b_kerten, cat_telurasin, 'Udang Telur Asin',  '', 37500, 'fa-bowl-food',      'bg-amber-50', NULL, true, 83),
        (b_kerten, cat_telurasin, 'Gurame Telur Asin', '', 37500, 'fa-fish',           'bg-amber-50', NULL, true, 84);

    -- 6. MASAKAN GURAME (9 items)
    INSERT INTO menu_items (branch_id, category_id, name, description, price, icon, icon_bg_class, tag, is_available, sort_order)
    VALUES
        (b_kerten, cat_gurame, 'Soup Gurame Kuah Sayur Asin', '', 34500, 'fa-fish', 'bg-amber-50', NULL, true, 85),
        (b_kerten, cat_gurame, 'Gurame Goreng Mareme',        '', 30500, 'fa-fish', 'bg-amber-50', NULL, true, 86),
        (b_kerten, cat_gurame, 'Gurame Goreng Asam Manis',    '', 30500, 'fa-fish', 'bg-amber-50', NULL, true, 87),
        (b_kerten, cat_gurame, 'Gurame Goreng Mentega',       '', 30500, 'fa-fish', 'bg-amber-50', NULL, true, 88),
        (b_kerten, cat_gurame, 'Gurame Goreng Tepung',        '', 30500, 'fa-fish', 'bg-amber-50', NULL, true, 89),
        (b_kerten, cat_gurame, 'Gurame Masak Tauco',          '', 30500, 'fa-fish', 'bg-amber-50', NULL, true, 90),
        (b_kerten, cat_gurame, 'Gurame Masak Kecap',          '', 30500, 'fa-fish', 'bg-amber-50', NULL, true, 91),
        (b_kerten, cat_gurame, 'Gurame Gongso',               '', 30500, 'fa-fish', 'bg-amber-50', NULL, true, 92),
        (b_kerten, cat_gurame, 'Gurame Oven Garam',           '', 30500, 'fa-fish', 'bg-amber-50', NULL, true, 93);

    -- 7. MASAKAN UDANG (11 items)
    INSERT INTO menu_items (branch_id, category_id, name, description, price, icon, icon_bg_class, tag, is_available, sort_order)
    VALUES
        (b_kerten, cat_udang, 'Udang Bago Spesial',      '', 33000, 'fa-shrimp', 'bg-amber-50',   NULL, true, 94),
        (b_kerten, cat_udang, 'Udang Bago Mentega',      '', 31000, 'fa-shrimp', 'bg-amber-50',   NULL, true, 95),
        (b_kerten, cat_udang, 'Udang Goreng Mareme',     '', 30500, 'fa-shrimp', 'bg-amber-50',   NULL, true, 96),
        (b_kerten, cat_udang, 'Udang Goreng Asam Manis', '', 30500, 'fa-shrimp', 'bg-amber-50',   NULL, true, 97),
        (b_kerten, cat_udang, 'Udang Goreng Mentega',    '', 30500, 'fa-shrimp', 'bg-amber-50',   NULL, true, 98),
        (b_kerten, cat_udang, 'Udang Goreng Tepung',     '', 30500, 'fa-shrimp', 'bg-amber-50',   NULL, true, 99),
        (b_kerten, cat_udang, 'Udang Gr Masak Tauco',    '', 30500, 'fa-shrimp', 'bg-amber-50',   NULL, true, 100),
        (b_kerten, cat_udang, 'Udang Ca Sayur',          '', 30500, 'fa-shrimp', 'bg-emerald-50', NULL, true, 101),
        (b_kerten, cat_udang, 'Udang Ca Jamur',          '', 30500, 'fa-shrimp', 'bg-amber-50',   NULL, true, 102),
        (b_kerten, cat_udang, 'Udang Ca Kay Lan',        '', 30500, 'fa-shrimp', 'bg-emerald-50', NULL, true, 103),
        (b_kerten, cat_udang, 'Udang Ca Kangkung',       '', 30500, 'fa-shrimp', 'bg-emerald-50', NULL, true, 104);

    -- 8. MASAKAN DAGING SAPI (9 items)
    INSERT INTO menu_items (branch_id, category_id, name, description, price, icon, icon_bg_class, tag, is_available, sort_order)
    VALUES
        (b_kerten, cat_sapi, 'Sapi Teriyaki',    '', 31000, 'fa-bowl-food', 'bg-amber-50',   NULL, true, 105),
        (b_kerten, cat_sapi, 'Kambing Teriyaki', '', 31000, 'fa-bowl-food', 'bg-amber-50',   NULL, true, 106),
        (b_kerten, cat_sapi, 'Sapi Ca Sayur',    '', 31000, 'fa-bowl-food', 'bg-emerald-50', NULL, true, 107),
        (b_kerten, cat_sapi, 'Sapi Ca Jamur',    '', 31000, 'fa-bowl-food', 'bg-amber-50',   NULL, true, 108),
        (b_kerten, cat_sapi, 'Sapi Ca Kaylan',   '', 31000, 'fa-bowl-food', 'bg-emerald-50', NULL, true, 109),
        (b_kerten, cat_sapi, 'Sapi Ca Kangkung', '', 31000, 'fa-bowl-food', 'bg-emerald-50', NULL, true, 110),
        (b_kerten, cat_sapi, 'Sapi Ca Brokoli',  '', 31000, 'fa-bowl-food', 'bg-emerald-50', NULL, true, 111),
        (b_kerten, cat_sapi, 'Sapi Lada Hitam',  '', 31000, 'fa-bowl-food', 'bg-stone-100',  NULL, true, 112),
        (b_kerten, cat_sapi, 'Bestik Sapi',      '', 31000, 'fa-bowl-food', 'bg-amber-50',   NULL, true, 113);

    -- 9. MASAKAN CUMI (6 items)
    INSERT INTO menu_items (branch_id, category_id, name, description, price, icon, icon_bg_class, tag, is_available, sort_order)
    VALUES
        (b_kerten, cat_cumi, 'Cumi Goreng Mareme',     '', 30500, 'fa-bowl-food', 'bg-amber-50',   NULL, true, 114),
        (b_kerten, cat_cumi, 'Cumi Goreng Asam Manis', '', 30500, 'fa-bowl-food', 'bg-amber-50',   NULL, true, 115),
        (b_kerten, cat_cumi, 'Cumi Goreng Mentega',    '', 30500, 'fa-bowl-food', 'bg-amber-50',   NULL, true, 116),
        (b_kerten, cat_cumi, 'Cumi Goreng Tepung',     '', 30500, 'fa-bowl-food', 'bg-amber-50',   NULL, true, 117),
        (b_kerten, cat_cumi, 'Cumi Ca Sayur',          '', 30500, 'fa-bowl-food', 'bg-emerald-50', NULL, true, 118),
        (b_kerten, cat_cumi, 'Cumi Ca Jamur',          '', 30500, 'fa-bowl-food', 'bg-amber-50',   NULL, true, 119);

    -- 10. MINUMAN (19 items)
    INSERT INTO menu_items (branch_id, category_id, name, description, price, icon, icon_bg_class, tag, is_available, sort_order)
    VALUES
        (b_kerten, cat_minuman, 'Es Pelangi',             '', 8000,  'fa-glass-water', 'bg-pink-50',    NULL, true, 120),
        (b_kerten, cat_minuman, 'Lemon Tea',              '', 6500,  'fa-mug-hot',     'bg-amber-50',   NULL, true, 121),
        (b_kerten, cat_minuman, 'Wedang Teh Tawar',       '', 3000,  'fa-mug-hot',     'bg-amber-50',   NULL, true, 122),
        (b_kerten, cat_minuman, 'Wedang Teh Manis',       '', 4000,  'fa-mug-hot',     'bg-amber-50',   NULL, true, 123),
        (b_kerten, cat_minuman, 'Wedang Jeruk',           '', 6500,  'fa-mug-hot',     'bg-amber-50',   NULL, true, 124),
        (b_kerten, cat_minuman, 'Wedang Kopi',            '', 6000,  'fa-mug-hot',     'bg-stone-100',  NULL, true, 125),
        (b_kerten, cat_minuman, 'Wedang Kopi Susu',       '', 9000,  'fa-mug-saucer',  'bg-amber-50',   NULL, true, 126),
        (b_kerten, cat_minuman, 'Es Teh Tawar',           '', 3000,  'fa-glass-water', 'bg-amber-50',   NULL, true, 127),
        (b_kerten, cat_minuman, 'Es Teh Manis',           '', 4000,  'fa-glass-water', 'bg-amber-50',   NULL, true, 128),
        (b_kerten, cat_minuman, 'Es Hijau',               '', 8000,  'fa-glass-water', 'bg-emerald-50', NULL, true, 129),
        (b_kerten, cat_minuman, 'Es Kopi',                '', 7000,  'fa-glass-water', 'bg-stone-100',  NULL, true, 130),
        (b_kerten, cat_minuman, 'Es Kopi Susu',           '', 9000,  'fa-mug-saucer',  'bg-amber-50',   NULL, true, 131),
        (b_kerten, cat_minuman, 'Es Jeruk',               '', 7000,  'fa-glass-water', 'bg-amber-50',   NULL, true, 132),
        (b_kerten, cat_minuman, 'Es Sirup',               '', 5500,  'fa-glass-water', 'bg-pink-50',    NULL, true, 133),
        (b_kerten, cat_minuman, 'Es Lemon Tea',           '', 6500,  'fa-glass-water', 'bg-amber-50',   NULL, true, 134),
        (b_kerten, cat_minuman, 'Es Soda Gembira',        '', 12000, 'fa-glass-water', 'bg-pink-50',    NULL, true, 135),
        (b_kerten, cat_minuman, 'Es Batu',                '', 1000,  'fa-glass-water', 'bg-stone-100',  NULL, true, 136),
        (b_kerten, cat_minuman, 'Coca Cola/Sprite/Fanta', '', 6000,  'fa-glass-water', 'bg-stone-100',  NULL, true, 137),
        (b_kerten, cat_minuman, 'Air Mineral',            '', 5000,  'fa-glass-water', 'bg-stone-100',  NULL, true, 138);

    -- =========================================================================
    -- OUTLET 2: MAK DJAN (64 items from Mak Djan menu)
    -- =========================================================================

    -- 1. MASAKAN TELOR ASIN (4 items)
    INSERT INTO menu_items (branch_id, category_id, name, description, price, icon, icon_bg_class, tag, is_available, sort_order)
    VALUES
        (b_makdjan, cat_telurasin, 'Gurame Grng Telor Asin', '', 36000, 'fa-fish',           'bg-amber-50', NULL, true, 1),
        (b_makdjan, cat_telurasin, 'Ayam Grng Telor Asin',   '', 36000, 'fa-drumstick-bite', 'bg-amber-50', NULL, true, 2),
        (b_makdjan, cat_telurasin, 'Udang Grng Telor Asin',  '', 36000, 'fa-shrimp',         'bg-amber-50', NULL, true, 3),
        (b_makdjan, cat_telurasin, 'Cumi Grng Telor Asin',   '', 36000, 'fa-bowl-food',      'bg-amber-50', NULL, true, 4);

    -- 2. MASAKAN SAPI (10 items)
    INSERT INTO menu_items (branch_id, category_id, name, description, price, icon, icon_bg_class, tag, is_available, sort_order)
    VALUES
        (b_makdjan, cat_sapi, 'Sapi Ca Sayur',    '', 30000, 'fa-bowl-food', 'bg-emerald-50', NULL, true, 5),
        (b_makdjan, cat_sapi, 'Sapi Ca Jamur',    '', 30000, 'fa-bowl-food', 'bg-amber-50',   NULL, true, 6),
        (b_makdjan, cat_sapi, 'Sapi Ca Kaylan',   '', 30000, 'fa-bowl-food', 'bg-emerald-50', NULL, true, 7),
        (b_makdjan, cat_sapi, 'Sapi Ca Kangkung', '', 30000, 'fa-bowl-food', 'bg-emerald-50', NULL, true, 8),
        (b_makdjan, cat_sapi, 'Sapi Ca Brokoli',  '', 30000, 'fa-bowl-food', 'bg-emerald-50', NULL, true, 9),
        (b_makdjan, cat_sapi, 'Sapi Lada Hitam',  '', 30000, 'fa-bowl-food', 'bg-stone-100',  NULL, true, 10),
        (b_makdjan, cat_sapi, 'Sapi Masak Jahe',  '', 32000, 'fa-bowl-food', 'bg-amber-50',   NULL, true, 11),
        (b_makdjan, cat_sapi, 'Sapi Teriyaki',    '', 30000, 'fa-bowl-food', 'bg-amber-50',   NULL, true, 12),
        (b_makdjan, cat_sapi, 'Bestik Sapi',      '', 30000, 'fa-bowl-food', 'bg-amber-50',   NULL, true, 13),
        (b_makdjan, cat_sapi, 'Babat Gongso',     '', 25500, 'fa-bowl-food', 'bg-amber-50',   NULL, true, 14);

    -- 3. MASAKAN GURAME (9 items)
    INSERT INTO menu_items (branch_id, category_id, name, description, price, icon, icon_bg_class, tag, is_available, sort_order)
    VALUES
        (b_makdjan, cat_gurame, 'Sup Gurame Kuah Sayur Asin', '', 33000, 'fa-fish', 'bg-amber-50', NULL, true, 15),
        (b_makdjan, cat_gurame, 'Gurame Oven Garam',          '', 32000, 'fa-fish', 'bg-amber-50', NULL, true, 16),
        (b_makdjan, cat_gurame, 'Gurame Goreng Mak Djan',     '', 29000, 'fa-fish', 'bg-amber-50', NULL, true, 17),
        (b_makdjan, cat_gurame, 'Gurame Goreng Asam Manis',   '', 29000, 'fa-fish', 'bg-amber-50', NULL, true, 18),
        (b_makdjan, cat_gurame, 'Gurame Goreng Mentega',      '', 29000, 'fa-fish', 'bg-amber-50', NULL, true, 19),
        (b_makdjan, cat_gurame, 'Gurame Goreng Tepung',       '', 29000, 'fa-fish', 'bg-amber-50', NULL, true, 20),
        (b_makdjan, cat_gurame, 'Gurame Masak Taoco',         '', 29000, 'fa-fish', 'bg-amber-50', NULL, true, 21),
        (b_makdjan, cat_gurame, 'Gurame Masak Kecap',         '', 29000, 'fa-fish', 'bg-amber-50', NULL, true, 22),
        (b_makdjan, cat_gurame, 'Gurame Gongso',              '', 29000, 'fa-fish', 'bg-amber-50', NULL, true, 23);

    -- 4. MASAKAN CUMI (6 items)
    INSERT INTO menu_items (branch_id, category_id, name, description, price, icon, icon_bg_class, tag, is_available, sort_order)
    VALUES
        (b_makdjan, cat_cumi, 'Cumi Goreng Mak Djan',   '', 29000, 'fa-bowl-food', 'bg-amber-50',   NULL, true, 24),
        (b_makdjan, cat_cumi, 'Cumi Goreng Asam Manis', '', 29000, 'fa-bowl-food', 'bg-amber-50',   NULL, true, 25),
        (b_makdjan, cat_cumi, 'Cumi Goreng Mentega',    '', 29000, 'fa-bowl-food', 'bg-amber-50',   NULL, true, 26),
        (b_makdjan, cat_cumi, 'Cumi Goreng Tepung',     '', 29000, 'fa-bowl-food', 'bg-amber-50',   NULL, true, 27),
        (b_makdjan, cat_cumi, 'Cumi Ca Sayur',          '', 29000, 'fa-bowl-food', 'bg-emerald-50', NULL, true, 28),
        (b_makdjan, cat_cumi, 'Cumi Ca Jamur',          '', 29000, 'fa-bowl-food', 'bg-amber-50',   NULL, true, 29);

    -- 5. MASAKAN UDANG (11 items)
    INSERT INTO menu_items (branch_id, category_id, name, description, price, icon, icon_bg_class, tag, is_available, sort_order)
    VALUES
        (b_makdjan, cat_udang, 'Udang Bago Spesial',       '', 32000, 'fa-shrimp', 'bg-amber-50',   NULL, true, 30),
        (b_makdjan, cat_udang, 'Udang Bago Grng Mentega',  '', 30000, 'fa-shrimp', 'bg-amber-50',   NULL, true, 31),
        (b_makdjan, cat_udang, 'Udang Goreng Mak Djan',    '', 29000, 'fa-shrimp', 'bg-amber-50',   NULL, true, 32),
        (b_makdjan, cat_udang, 'Udang Goreng Asam Manis',  '', 29000, 'fa-shrimp', 'bg-amber-50',   NULL, true, 33),
        (b_makdjan, cat_udang, 'Udang Goreng Mentega',     '', 29000, 'fa-shrimp', 'bg-amber-50',   NULL, true, 34),
        (b_makdjan, cat_udang, 'Udang Goreng Tepung',      '', 29000, 'fa-shrimp', 'bg-amber-50',   NULL, true, 35),
        (b_makdjan, cat_udang, 'Udang Goreng Masak Taoco', '', 29000, 'fa-shrimp', 'bg-amber-50',   NULL, true, 36),
        (b_makdjan, cat_udang, 'Udang Ca Sayur',           '', 29000, 'fa-shrimp', 'bg-emerald-50', NULL, true, 37),
        (b_makdjan, cat_udang, 'Udang Ca Jamur',           '', 29000, 'fa-shrimp', 'bg-amber-50',   NULL, true, 38),
        (b_makdjan, cat_udang, 'Udang Ca Kaylan',          '', 29000, 'fa-shrimp', 'bg-emerald-50', NULL, true, 39),
        (b_makdjan, cat_udang, 'Udang Ca Kangkung',        '', 29000, 'fa-shrimp', 'bg-emerald-50', NULL, true, 40);

    -- 6. ANEKA SOUP (6 items)
    INSERT INTO menu_items (branch_id, category_id, name, description, price, icon, icon_bg_class, tag, is_available, sort_order)
    VALUES
        (b_makdjan, cat_soup, 'Soup Asparagus Kepiting',       '', 42000, 'fa-bowl-food', 'bg-amber-50', NULL, true, 41),
        (b_makdjan, cat_soup, 'Soup Asparagus Kepiting (1/2)', '', 25000, 'fa-bowl-food', 'bg-amber-50', NULL, true, 42),
        (b_makdjan, cat_soup, 'Soup Shanghay',                 '', 33000, 'fa-bowl-food', 'bg-amber-50', NULL, true, 43),
        (b_makdjan, cat_soup, 'Soup Ayam',                     '', 29000, 'fa-bowl-food', 'bg-amber-50', NULL, true, 44),
        (b_makdjan, cat_soup, 'Soup Tahu Sayur Asin',          '', 29000, 'fa-bowl-food', 'bg-amber-50', NULL, true, 45),
        (b_makdjan, cat_soup, 'Soup Sapi',                     '', 30000, 'fa-bowl-food', 'bg-amber-50', NULL, true, 46);

    -- 7. ANEKA MINUMAN (18 items)
    INSERT INTO menu_items (branch_id, category_id, name, description, price, icon, icon_bg_class, tag, is_available, sort_order)
    VALUES
        (b_makdjan, cat_minuman, 'Es Pelangi',             '', 8000,  'fa-glass-water', 'bg-pink-50',    NULL, true, 47),
        (b_makdjan, cat_minuman, 'Lemon Tea',              '', 6000,  'fa-mug-hot',     'bg-amber-50',   NULL, true, 48),
        (b_makdjan, cat_minuman, 'Wedang Teh Tawar',       '', 3000,  'fa-mug-hot',     'bg-amber-50',   NULL, true, 49),
        (b_makdjan, cat_minuman, 'Wedang Teh Manis',       '', 5000,  'fa-mug-hot',     'bg-amber-50',   NULL, true, 50),
        (b_makdjan, cat_minuman, 'Wedang Jeruk',           '', 7000,  'fa-mug-hot',     'bg-amber-50',   NULL, true, 51),
        (b_makdjan, cat_minuman, 'Wedang Kopi',            '', 6000,  'fa-mug-hot',     'bg-stone-100',  NULL, true, 52),
        (b_makdjan, cat_minuman, 'Wedang Kopi Susu',       '', 8000,  'fa-mug-saucer',  'bg-amber-50',   NULL, true, 53),
        (b_makdjan, cat_minuman, 'Es Teh Tawar',           '', 4000,  'fa-glass-water', 'bg-amber-50',   NULL, true, 54),
        (b_makdjan, cat_minuman, 'Es Teh Manis',           '', 5000,  'fa-glass-water', 'bg-amber-50',   NULL, true, 55),
        (b_makdjan, cat_minuman, 'Es Hijau',               '', 8000,  'fa-glass-water', 'bg-emerald-50', NULL, true, 56),
        (b_makdjan, cat_minuman, 'Es Kopi',                '', 7000,  'fa-glass-water', 'bg-stone-100',  NULL, true, 57),
        (b_makdjan, cat_minuman, 'Es Jeruk',               '', 8000,  'fa-glass-water', 'bg-amber-50',   NULL, true, 58),
        (b_makdjan, cat_minuman, 'Es Sirup',               '', 6000,  'fa-glass-water', 'bg-pink-50',    NULL, true, 59),
        (b_makdjan, cat_minuman, 'Es Lemon Tea',           '', 7000,  'fa-glass-water', 'bg-amber-50',   NULL, true, 60),
        (b_makdjan, cat_minuman, 'Es Soda Gembira',        '', 12000, 'fa-glass-water', 'bg-pink-50',    NULL, true, 61),
        (b_makdjan, cat_minuman, 'Es Batu',                '', 1000,  'fa-glass-water', 'bg-stone-100',  NULL, true, 62),
        (b_makdjan, cat_minuman, 'Coca Cola/Fanta/Sprite', '', 8000,  'fa-glass-water', 'bg-stone-100',  NULL, true, 63),
        (b_makdjan, cat_minuman, 'Aqua',                   '', 5000,  'fa-glass-water', 'bg-stone-100',  NULL, true, 64);
END $$;
