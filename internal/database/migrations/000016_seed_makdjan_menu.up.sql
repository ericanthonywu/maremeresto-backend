-- 000016_seed_makdjan_menu.up.sql
-- Seed complete Mak Djan (branch 33333333-3333-3333-3333-333333333333) menu
-- from the official printed menu brochure (146 items across 10 categories)

DELETE FROM menu_items WHERE branch_id = '33333333-3333-3333-3333-333333333333';

DO $$
DECLARE
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
    -- 1. ANEKA NASI GORENG (30 items)
    INSERT INTO menu_items (branch_id, category_id, name, description, price, icon, icon_bg_class, tag, is_available, sort_order)
    VALUES
        (b_makdjan, cat_nasgor, 'Nasi Goreng Ijo',             '', 24000, 'fa-bowl-rice', 'bg-emerald-50', NULL, true, 1),
        (b_makdjan, cat_nasgor, 'Nasi Goreng Ijo Istimewa',    '', 26000, 'fa-bowl-rice', 'bg-emerald-50', NULL, true, 2),
        (b_makdjan, cat_nasgor, 'Nasi Goreng Ndeso',           '', 24000, 'fa-bowl-rice', 'bg-amber-50',   NULL, true, 3),
        (b_makdjan, cat_nasgor, 'Nasi Goreng Ndeso Istimewa',  '', 26000, 'fa-bowl-rice', 'bg-amber-50',   NULL, true, 4),
        (b_makdjan, cat_nasgor, 'Nasi Goreng Ayam',            '', 24000, 'fa-bowl-rice', 'bg-amber-50',   NULL, true, 5),
        (b_makdjan, cat_nasgor, 'Nasi Goreng Ayam Istimewa',   '', 26000, 'fa-bowl-rice', 'bg-amber-50',   NULL, true, 6),
        (b_makdjan, cat_nasgor, 'Nasi Goreng Kambing',         '', 25000, 'fa-bowl-rice', 'bg-amber-50',   NULL, true, 7),
        (b_makdjan, cat_nasgor, 'Nasi Goreng Kambing Istimewa', '', 27000, 'fa-bowl-rice', 'bg-amber-50',   NULL, true, 8),
        (b_makdjan, cat_nasgor, 'Nasi Goreng Babat',           '', 25000, 'fa-bowl-rice', 'bg-amber-50',   NULL, true, 9),
        (b_makdjan, cat_nasgor, 'Nasi Goreng Babat Istimewa',  '', 27000, 'fa-bowl-rice', 'bg-amber-50',   NULL, true, 10),
        (b_makdjan, cat_nasgor, 'Nasi Goreng Seafood',         '', 26000, 'fa-bowl-rice', 'bg-amber-50',   NULL, true, 11),
        (b_makdjan, cat_nasgor, 'Nasi Goreng Seafood Istimewa', '', 28000, 'fa-bowl-rice', 'bg-amber-50',   NULL, true, 12),
        (b_makdjan, cat_nasgor, 'Nasi Goreng Ijo Seafood',     '', 26000, 'fa-bowl-rice', 'bg-emerald-50', NULL, true, 13),
        (b_makdjan, cat_nasgor, 'Nasi Goreng Ijo Seafood Ist', '', 28000, 'fa-bowl-rice', 'bg-emerald-50', NULL, true, 14),
        (b_makdjan, cat_nasgor, 'Nasi Goreng Udang',           '', 25000, 'fa-bowl-rice', 'bg-amber-50',   NULL, true, 15),
        (b_makdjan, cat_nasgor, 'Nasi Goreng Udang Istimewa',  '', 27000, 'fa-bowl-rice', 'bg-amber-50',   NULL, true, 16),
        (b_makdjan, cat_nasgor, 'Nasi Goreng Ijo Udang',       '', 25000, 'fa-bowl-rice', 'bg-emerald-50', NULL, true, 17),
        (b_makdjan, cat_nasgor, 'Nasi Goreng Ijo Udang Ist',   '', 27000, 'fa-bowl-rice', 'bg-emerald-50', NULL, true, 18),
        (b_makdjan, cat_nasgor, 'Nasi Goreng Sapi',            '', 25000, 'fa-bowl-rice', 'bg-amber-50',   NULL, true, 19),
        (b_makdjan, cat_nasgor, 'Nasi Goreng Sapi Istimewa',   '', 27000, 'fa-bowl-rice', 'bg-amber-50',   NULL, true, 20),
        (b_makdjan, cat_nasgor, 'Nasi Goreng Lada Hitam',      '', 25000, 'fa-bowl-rice', 'bg-stone-100',  NULL, true, 21),
        (b_makdjan, cat_nasgor, 'Nasi Goreng Lada Hitam Ist',  '', 27000, 'fa-bowl-rice', 'bg-stone-100',  NULL, true, 22),
        (b_makdjan, cat_nasgor, 'Nasi Goreng Sosis',           '', 24000, 'fa-bowl-rice', 'bg-amber-50',   NULL, true, 23),
        (b_makdjan, cat_nasgor, 'Nasi Goreng Sosis Istimewa',  '', 26000, 'fa-bowl-rice', 'bg-amber-50',   NULL, true, 24),
        (b_makdjan, cat_nasgor, 'Nasi Goreng Bakso',           '', 24000, 'fa-bowl-rice', 'bg-amber-50',   NULL, true, 25),
        (b_makdjan, cat_nasgor, 'Nasi Goreng Bakso Istimewa',  '', 26000, 'fa-bowl-rice', 'bg-amber-50',   NULL, true, 26),
        (b_makdjan, cat_nasgor, 'Nasi Goreng Nanas',           '', 24000, 'fa-bowl-rice', 'bg-amber-50',   NULL, true, 27),
        (b_makdjan, cat_nasgor, 'Nasi Goreng Nanas Istimewa',  '', 26000, 'fa-bowl-rice', 'bg-amber-50',   NULL, true, 28),
        (b_makdjan, cat_nasgor, 'Nasi Goreng Mawut',           '', 26000, 'fa-bowl-rice', 'bg-amber-50',   NULL, true, 29),
        (b_makdjan, cat_nasgor, 'Nasi Goreng Mawut Istimewa',  '', 28000, 'fa-bowl-rice', 'bg-amber-50',   NULL, true, 30);

    -- 2. MASAKAN DAGING AYAM (23 items)
    INSERT INTO menu_items (branch_id, category_id, name, description, price, icon, icon_bg_class, tag, is_available, sort_order)
    VALUES
        (b_makdjan, cat_ayam, 'Ayam Saos Lemon',            '', 32000, 'fa-drumstick-bite', 'bg-amber-50',   NULL, true, 31),
        (b_makdjan, cat_ayam, 'Ayam Gongso',                '', 29000, 'fa-drumstick-bite', 'bg-amber-50',   NULL, true, 32),
        (b_makdjan, cat_ayam, 'Ayam Lada Hitam',            '', 29000, 'fa-drumstick-bite', 'bg-amber-50',   NULL, true, 33),
        (b_makdjan, cat_ayam, 'Ayam Masak Kecap Pedas',     '', 29000, 'fa-drumstick-bite', 'bg-amber-50',   NULL, true, 34),
        (b_makdjan, cat_ayam, 'Ayam Goreng Mak Djan',       '', 29000, 'fa-drumstick-bite', 'bg-amber-50',   NULL, true, 35),
        (b_makdjan, cat_ayam, 'Ayam Goreng Asam Manis',     '', 29000, 'fa-drumstick-bite', 'bg-amber-50',   NULL, true, 36),
        (b_makdjan, cat_ayam, 'Ayam Goreng Mentega',        '', 29000, 'fa-drumstick-bite', 'bg-amber-50',   NULL, true, 37),
        (b_makdjan, cat_ayam, 'Ayam Goreng Tepung',         '', 29000, 'fa-drumstick-bite', 'bg-amber-50',   NULL, true, 38),
        (b_makdjan, cat_ayam, 'Ayam Goreng Masak Taoco',    '', 29000, 'fa-drumstick-bite', 'bg-amber-50',   NULL, true, 39),
        (b_makdjan, cat_ayam, 'Ayam Goreng Lombok Ijo',     '', 29000, 'fa-drumstick-bite', 'bg-emerald-50', NULL, true, 40),
        (b_makdjan, cat_ayam, 'Ayam Ca Jamur',              '', 29000, 'fa-drumstick-bite', 'bg-amber-50',   NULL, true, 41),
        (b_makdjan, cat_ayam, 'Ayam Ca Kaylan',             '', 29000, 'fa-drumstick-bite', 'bg-emerald-50', NULL, true, 42),
        (b_makdjan, cat_ayam, 'Ayam Ca Kangkung',           '', 29000, 'fa-drumstick-bite', 'bg-emerald-50', NULL, true, 43),
        (b_makdjan, cat_ayam, 'Ayam Ca Brokoli',            '', 29000, 'fa-drumstick-bite', 'bg-emerald-50', NULL, true, 44),
        (b_makdjan, cat_ayam, 'Ayam Sayur Asin Ca/Kuah',    '', 29000, 'fa-drumstick-bite', 'bg-amber-50',   NULL, true, 45),
        (b_makdjan, cat_ayam, 'Ayam Nanking',               '', 29000, 'fa-drumstick-bite', 'bg-amber-50',   NULL, true, 46),
        (b_makdjan, cat_ayam, 'Ayam Rebus',                 '', 29000, 'fa-drumstick-bite', 'bg-stone-100',  NULL, true, 47),
        (b_makdjan, cat_ayam, 'Sayap Masak Kecap',          '', 29000, 'fa-drumstick-bite', 'bg-amber-50',   NULL, true, 48),
        (b_makdjan, cat_ayam, 'Sayap Masak Madu',           '', 29000, 'fa-drumstick-bite', 'bg-amber-50',   NULL, true, 49),
        (b_makdjan, cat_ayam, 'Bestik Ayam',                '', 29000, 'fa-drumstick-bite', 'bg-amber-50',   NULL, true, 50),
        (b_makdjan, cat_ayam, 'Ngoyang Kee',                '', 29000, 'fa-drumstick-bite', 'bg-amber-50',   NULL, true, 51),
        (b_makdjan, cat_ayam, 'Kulit Goreng',               '', 27000, 'fa-drumstick-bite', 'bg-amber-50',   NULL, true, 52),
        (b_makdjan, cat_ayam, 'Ayam Teriyaki',              '', 29000, 'fa-drumstick-bite', 'bg-amber-50',   NULL, true, 53);

    -- 3. ANEKA MASAKAN (29 items)
    INSERT INTO menu_items (branch_id, category_id, name, description, price, icon, icon_bg_class, tag, is_available, sort_order)
    VALUES
        (b_makdjan, cat_masakan, 'Bakmi Goreng',            '', 24000, 'fa-bowl-food', 'bg-amber-50',   NULL, true, 54),
        (b_makdjan, cat_masakan, 'Bakmi Godog',             '', 24000, 'fa-bowl-food', 'bg-amber-50',   NULL, true, 55),
        (b_makdjan, cat_masakan, 'Bihun Goreng',            '', 24000, 'fa-bowl-food', 'bg-amber-50',   NULL, true, 56),
        (b_makdjan, cat_masakan, 'Bihun Godog',             '', 24000, 'fa-bowl-food', 'bg-amber-50',   NULL, true, 57),
        (b_makdjan, cat_masakan, 'Cap Cay Grng/Godog',      '', 25000, 'fa-bowl-food', 'bg-emerald-50', NULL, true, 58),
        (b_makdjan, cat_masakan, 'Nasi Ayam Ca',            '', 26000, 'fa-bowl-rice', 'bg-amber-50',   NULL, true, 59),
        (b_makdjan, cat_masakan, 'Nasi Campur',             '', 25000, 'fa-bowl-rice', 'bg-amber-50',   NULL, true, 60),
        (b_makdjan, cat_masakan, 'Nasi Putih',              '', 6000,  'fa-bowl-rice', 'bg-stone-100',  NULL, true, 61),
        (b_makdjan, cat_masakan, 'Kambing Teriyaki',        '', 30000, 'fa-bowl-food', 'bg-amber-50',   NULL, true, 62),
        (b_makdjan, cat_masakan, 'Masak Tahu',              '', 27000, 'fa-bowl-food', 'bg-amber-50',   NULL, true, 63),
        (b_makdjan, cat_masakan, 'Tahu Masak Kecap',        '', 27000, 'fa-bowl-food', 'bg-amber-50',   NULL, true, 64),
        (b_makdjan, cat_masakan, 'Misoa Kuah',              '', 26000, 'fa-bowl-food', 'bg-amber-50',   NULL, true, 65),
        (b_makdjan, cat_masakan, 'Mun/Angsio Tahu',         '', 27000, 'fa-bowl-food', 'bg-amber-50',   NULL, true, 66),
        (b_makdjan, cat_masakan, 'Sapo Tahu',               '', 32000, 'fa-bowl-food', 'bg-amber-50',   NULL, true, 67),
        (b_makdjan, cat_masakan, 'Soon Goreng',             '', 26000, 'fa-bowl-food', 'bg-amber-50',   NULL, true, 68),
        (b_makdjan, cat_masakan, 'Tamie Goreng',            '', 25000, 'fa-bowl-food', 'bg-amber-50',   NULL, true, 69),
        (b_makdjan, cat_masakan, 'Yam Yien',                '', 26000, 'fa-bowl-food', 'bg-amber-50',   NULL, true, 70),
        (b_makdjan, cat_masakan, 'Aneka Sayur',             '', 23000, 'fa-leaf',      'bg-emerald-50', NULL, true, 71),
        (b_makdjan, cat_masakan, 'Fu Yung Hay',             '', 26000, 'fa-bowl-food', 'bg-amber-50',   NULL, true, 72),
        (b_makdjan, cat_masakan, 'Kay Lan Ca Polos',        '', 23000, 'fa-leaf',      'bg-emerald-50', NULL, true, 73),
        (b_makdjan, cat_masakan, 'Kangkung Ca Polos',       '', 23000, 'fa-leaf',      'bg-emerald-50', NULL, true, 74),
        (b_makdjan, cat_masakan, 'Kweetiauw Goreng/Kuah',   '', 25000, 'fa-bowl-food', 'bg-amber-50',   NULL, true, 75),
        (b_makdjan, cat_masakan, 'Kweetiauw Nyemek',        '', 25000, 'fa-bowl-food', 'bg-amber-50',   NULL, true, 76),
        (b_makdjan, cat_masakan, 'Kamar Bola',              '', 26000, 'fa-bowl-food', 'bg-amber-50',   NULL, true, 77),
        (b_makdjan, cat_masakan, 'Loo Mie',                 '', 26000, 'fa-bowl-food', 'bg-amber-50',   NULL, true, 78),
        (b_makdjan, cat_masakan, 'Brokoli Polos',           '', 24000, 'fa-leaf',      'bg-emerald-50', NULL, true, 79),
        (b_makdjan, cat_masakan, 'Orak - Arik',             '', 25000, 'fa-bowl-food', 'bg-amber-50',   NULL, true, 80),
        (b_makdjan, cat_masakan, 'Koloo Kee',               '', 28000, 'fa-bowl-food', 'bg-amber-50',   NULL, true, 81),
        (b_makdjan, cat_masakan, 'Paklay Ca',               '', 26000, 'fa-bowl-food', 'bg-amber-50',   NULL, true, 82);

    -- 4. ANEKA SOUP (6 items)
    INSERT INTO menu_items (branch_id, category_id, name, description, price, icon, icon_bg_class, tag, is_available, sort_order)
    VALUES
        (b_makdjan, cat_soup, 'Soup Asparagus Kepiting',       '', 42000, 'fa-bowl-food', 'bg-amber-50', NULL, true, 83),
        (b_makdjan, cat_soup, 'Soup Asparagus Kepiting (1/2)', '', 25000, 'fa-bowl-food', 'bg-amber-50', NULL, true, 84),
        (b_makdjan, cat_soup, 'Soup Shanghay',                 '', 33000, 'fa-bowl-food', 'bg-amber-50', NULL, true, 85),
        (b_makdjan, cat_soup, 'Soup Ayam',                     '', 29000, 'fa-bowl-food', 'bg-amber-50', NULL, true, 86),
        (b_makdjan, cat_soup, 'Soup Tahu Sayur Asin',          '', 29000, 'fa-bowl-food', 'bg-amber-50', NULL, true, 87),
        (b_makdjan, cat_soup, 'Soup Sapi',                     '', 30000, 'fa-bowl-food', 'bg-amber-50', NULL, true, 88);

    -- 5. MASAKAN TELUR ASIN (4 items)
    INSERT INTO menu_items (branch_id, category_id, name, description, price, icon, icon_bg_class, tag, is_available, sort_order)
    VALUES
        (b_makdjan, cat_telurasin, 'Gurame Grng Telor Asin', '', 36000, 'fa-fish',           'bg-amber-50', NULL, true, 89),
        (b_makdjan, cat_telurasin, 'Ayam Grng Telor Asin',   '', 36000, 'fa-drumstick-bite', 'bg-amber-50', NULL, true, 90),
        (b_makdjan, cat_telurasin, 'Udang Grng Telor Asin',  '', 36000, 'fa-shrimp',         'bg-amber-50', NULL, true, 91),
        (b_makdjan, cat_telurasin, 'Cumi Grng Telor Asin',   '', 36000, 'fa-bowl-food',      'bg-amber-50', NULL, true, 92);

    -- 6. MASAKAN GURAME (9 items)
    INSERT INTO menu_items (branch_id, category_id, name, description, price, icon, icon_bg_class, tag, is_available, sort_order)
    VALUES
        (b_makdjan, cat_gurame, 'Sup Gurame Kuah Sayur Asin', '', 33000, 'fa-fish', 'bg-amber-50', NULL, true, 93),
        (b_makdjan, cat_gurame, 'Gurame Oven Garam',          '', 32000, 'fa-fish', 'bg-amber-50', NULL, true, 94),
        (b_makdjan, cat_gurame, 'Gurame Goreng Mak Djan',     '', 29000, 'fa-fish', 'bg-amber-50', NULL, true, 95),
        (b_makdjan, cat_gurame, 'Gurame Goreng Asam Manis',   '', 29000, 'fa-fish', 'bg-amber-50', NULL, true, 96),
        (b_makdjan, cat_gurame, 'Gurame Goreng Mentega',      '', 29000, 'fa-fish', 'bg-amber-50', NULL, true, 97),
        (b_makdjan, cat_gurame, 'Gurame Goreng Tepung',       '', 29000, 'fa-fish', 'bg-amber-50', NULL, true, 98),
        (b_makdjan, cat_gurame, 'Gurame Masak Taoco',         '', 29000, 'fa-fish', 'bg-amber-50', NULL, true, 99),
        (b_makdjan, cat_gurame, 'Gurame Masak Kecap',         '', 29000, 'fa-fish', 'bg-amber-50', NULL, true, 100),
        (b_makdjan, cat_gurame, 'Gurame Gongso',              '', 29000, 'fa-fish', 'bg-amber-50', NULL, true, 101);

    -- 7. MASAKAN UDANG (11 items)
    INSERT INTO menu_items (branch_id, category_id, name, description, price, icon, icon_bg_class, tag, is_available, sort_order)
    VALUES
        (b_makdjan, cat_udang, 'Udang Bago Spesial',       '', 32000, 'fa-shrimp', 'bg-amber-50',   NULL, true, 102),
        (b_makdjan, cat_udang, 'Udang Bago Grng Mentega',  '', 30000, 'fa-shrimp', 'bg-amber-50',   NULL, true, 103),
        (b_makdjan, cat_udang, 'Udang Goreng Mak Djan',    '', 29000, 'fa-shrimp', 'bg-amber-50',   NULL, true, 104),
        (b_makdjan, cat_udang, 'Udang Goreng Asam Manis',  '', 29000, 'fa-shrimp', 'bg-amber-50',   NULL, true, 105),
        (b_makdjan, cat_udang, 'Udang Goreng Mentega',     '', 29000, 'fa-shrimp', 'bg-amber-50',   NULL, true, 106),
        (b_makdjan, cat_udang, 'Udang Goreng Tepung',      '', 29000, 'fa-shrimp', 'bg-amber-50',   NULL, true, 107),
        (b_makdjan, cat_udang, 'Udang Goreng Masak Taoco', '', 29000, 'fa-shrimp', 'bg-amber-50',   NULL, true, 108),
        (b_makdjan, cat_udang, 'Udang Ca Sayur',           '', 29000, 'fa-shrimp', 'bg-emerald-50', NULL, true, 109),
        (b_makdjan, cat_udang, 'Udang Ca Jamur',           '', 29000, 'fa-shrimp', 'bg-amber-50',   NULL, true, 110),
        (b_makdjan, cat_udang, 'Udang Ca Kaylan',          '', 29000, 'fa-shrimp', 'bg-emerald-50', NULL, true, 111),
        (b_makdjan, cat_udang, 'Udang Ca Kangkung',        '', 29000, 'fa-shrimp', 'bg-emerald-50', NULL, true, 112);

    -- 8. MASAKAN DAGING SAPI (10 items)
    INSERT INTO menu_items (branch_id, category_id, name, description, price, icon, icon_bg_class, tag, is_available, sort_order)
    VALUES
        (b_makdjan, cat_sapi, 'Sapi Ca Sayur',    '', 30000, 'fa-bowl-food', 'bg-emerald-50', NULL, true, 113),
        (b_makdjan, cat_sapi, 'Sapi Ca Jamur',    '', 30000, 'fa-bowl-food', 'bg-amber-50',   NULL, true, 114),
        (b_makdjan, cat_sapi, 'Sapi Ca Kaylan',   '', 30000, 'fa-bowl-food', 'bg-emerald-50', NULL, true, 115),
        (b_makdjan, cat_sapi, 'Sapi Ca Kangkung', '', 30000, 'fa-bowl-food', 'bg-emerald-50', NULL, true, 116),
        (b_makdjan, cat_sapi, 'Sapi Ca Brokoli',  '', 30000, 'fa-bowl-food', 'bg-emerald-50', NULL, true, 117),
        (b_makdjan, cat_sapi, 'Sapi Lada Hitam',  '', 30000, 'fa-bowl-food', 'bg-stone-100',  NULL, true, 118),
        (b_makdjan, cat_sapi, 'Sapi Masak Jahe',  '', 32000, 'fa-bowl-food', 'bg-amber-50',   NULL, true, 119),
        (b_makdjan, cat_sapi, 'Sapi Teriyaki',    '', 30000, 'fa-bowl-food', 'bg-amber-50',   NULL, true, 120),
        (b_makdjan, cat_sapi, 'Bestik Sapi',      '', 30000, 'fa-bowl-food', 'bg-amber-50',   NULL, true, 121),
        (b_makdjan, cat_sapi, 'Babat Gongso',     '', 25500, 'fa-bowl-food', 'bg-amber-50',   NULL, true, 122);

    -- 9. MASAKAN CUMI (6 items)
    INSERT INTO menu_items (branch_id, category_id, name, description, price, icon, icon_bg_class, tag, is_available, sort_order)
    VALUES
        (b_makdjan, cat_cumi, 'Cumi Goreng Mak Djan',   '', 29000, 'fa-bowl-food', 'bg-amber-50',   NULL, true, 123),
        (b_makdjan, cat_cumi, 'Cumi Goreng Asam Manis', '', 29000, 'fa-bowl-food', 'bg-amber-50',   NULL, true, 124),
        (b_makdjan, cat_cumi, 'Cumi Goreng Mentega',    '', 29000, 'fa-bowl-food', 'bg-amber-50',   NULL, true, 125),
        (b_makdjan, cat_cumi, 'Cumi Goreng Tepung',     '', 29000, 'fa-bowl-food', 'bg-amber-50',   NULL, true, 126),
        (b_makdjan, cat_cumi, 'Cumi Ca Sayur',          '', 29000, 'fa-bowl-food', 'bg-emerald-50', NULL, true, 127),
        (b_makdjan, cat_cumi, 'Cumi Ca Jamur',          '', 29000, 'fa-bowl-food', 'bg-amber-50',   NULL, true, 128);

    -- 10. ANEKA MINUMAN (18 items)
    INSERT INTO menu_items (branch_id, category_id, name, description, price, icon, icon_bg_class, tag, is_available, sort_order)
    VALUES
        (b_makdjan, cat_minuman, 'Es Pelangi',             '', 8000,  'fa-glass-water', 'bg-pink-50',    NULL, true, 129),
        (b_makdjan, cat_minuman, 'Lemon Tea',              '', 6000,  'fa-mug-hot',     'bg-amber-50',   NULL, true, 130),
        (b_makdjan, cat_minuman, 'Wedang Teh Tawar',       '', 3000,  'fa-mug-hot',     'bg-amber-50',   NULL, true, 131),
        (b_makdjan, cat_minuman, 'Wedang Teh Manis',       '', 5000,  'fa-mug-hot',     'bg-amber-50',   NULL, true, 132),
        (b_makdjan, cat_minuman, 'Wedang Jeruk',           '', 7000,  'fa-mug-hot',     'bg-amber-50',   NULL, true, 133),
        (b_makdjan, cat_minuman, 'Wedang Kopi',            '', 6000,  'fa-mug-hot',     'bg-stone-100',  NULL, true, 134),
        (b_makdjan, cat_minuman, 'Wedang Kopi Susu',       '', 8000,  'fa-mug-saucer',  'bg-amber-50',   NULL, true, 135),
        (b_makdjan, cat_minuman, 'Es Teh Tawar',           '', 4000,  'fa-glass-water', 'bg-amber-50',   NULL, true, 136),
        (b_makdjan, cat_minuman, 'Es Teh Manis',           '', 5000,  'fa-glass-water', 'bg-amber-50',   NULL, true, 137),
        (b_makdjan, cat_minuman, 'Es Hijau',               '', 8000,  'fa-glass-water', 'bg-emerald-50', NULL, true, 138),
        (b_makdjan, cat_minuman, 'Es Kopi',                '', 7000,  'fa-glass-water', 'bg-stone-100',  NULL, true, 139),
        (b_makdjan, cat_minuman, 'Es Jeruk',               '', 8000,  'fa-glass-water', 'bg-amber-50',   NULL, true, 140),
        (b_makdjan, cat_minuman, 'Es Sirup',               '', 6000,  'fa-glass-water', 'bg-pink-50',    NULL, true, 141),
        (b_makdjan, cat_minuman, 'Es Lemon Tea',           '', 7000,  'fa-glass-water', 'bg-amber-50',   NULL, true, 142),
        (b_makdjan, cat_minuman, 'Es Soda Gembira',        '', 12000, 'fa-glass-water', 'bg-pink-50',    NULL, true, 143),
        (b_makdjan, cat_minuman, 'Es Batu',                '', 1000,  'fa-glass-water', 'bg-stone-100',  NULL, true, 144),
        (b_makdjan, cat_minuman, 'Coca Cola/Fanta/Sprite', '', 8000,  'fa-glass-water', 'bg-stone-100',  NULL, true, 145),
        (b_makdjan, cat_minuman, 'Aqua',                   '', 5000,  'fa-glass-water', 'bg-stone-100',  NULL, true, 146);
END $$;
