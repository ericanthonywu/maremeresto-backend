-- 000013_seed_samratulangi_menu.up.sql
-- Clear all previously seeded dummy menu items and categories, and seed the
-- official Mareme Samratulangi (branch 11111111-1111-1111-1111-111111111111) menu.

DELETE FROM menu_items;
DELETE FROM categories;

INSERT INTO categories (id, name, slug, emoji, sort_order)
VALUES
    ('c1111111-1111-1111-1111-111111111111', 'Aneka Nasi Goreng',   'aneka-nasi-goreng',   '🍛', 1),
    ('c2222222-2222-2222-2222-222222222222', 'Masakan Daging Ayam', 'masakan-daging-ayam', '🍗', 2),
    ('c3333333-3333-3333-3333-333333333333', 'Aneka Masakan',       'aneka-masakan',       '🍜', 3),
    ('c4444444-4444-4444-4444-444444444444', 'Aneka Soup',          'aneka-soup',          '🍲', 4),
    ('c5555555-5555-5555-5555-555555555555', 'Masakan Telur Asin',  'masakan-telur-asin',  '🥚', 5)
ON CONFLICT (slug) DO UPDATE SET
    name = EXCLUDED.name,
    emoji = EXCLUDED.emoji,
    sort_order = EXCLUDED.sort_order;

DO $$
DECLARE
    b_samratulangi UUID := '11111111-1111-1111-1111-111111111111';
    cat_nasgor     UUID := 'c1111111-1111-1111-1111-111111111111';
    cat_ayam       UUID := 'c2222222-2222-2222-2222-222222222222';
    cat_masakan    UUID := 'c3333333-3333-3333-3333-333333333333';
    cat_soup       UUID := 'c4444444-4444-4444-4444-444444444444';
    cat_telurasin  UUID := 'c5555555-5555-5555-5555-555555555555';
BEGIN
    -- 1. ANEKA NASI GORENG (23 items)
    INSERT INTO menu_items (branch_id, category_id, name, description, price, icon, icon_bg_class, tag, is_available, sort_order)
    VALUES
        (b_samratulangi, cat_nasgor, 'Nasi Goreng Ijo',             'Nasi goreng cabai hijau khas Mareme yang harum dan gurih.', 25500, 'fa-bowl-rice', 'bg-emerald-50', NULL,      true, 1),
        (b_samratulangi, cat_nasgor, 'Nasi Goreng Ijo Istimewa',    'Nasi goreng cabai hijau porsi spesial dengan topping lengkap.', 27000, 'fa-bowl-rice', 'bg-emerald-50', 'Favorit', true, 2),
        (b_samratulangi, cat_nasgor, 'Nasi Goreng Ijo Udang',       'Nasi goreng cabai hijau dipadu udang segar pilihan.', 26500, 'fa-bowl-rice', 'bg-emerald-50', NULL,      true, 3),
        (b_samratulangi, cat_nasgor, 'Nasi Goreng Ijo Udang Ist',   'Nasi goreng cabai hijau udang porsi istimewa lengkap.', 28000, 'fa-bowl-rice', 'bg-emerald-50', NULL,      true, 4),
        (b_samratulangi, cat_nasgor, 'Nasi Goreng Ijo Seafood',     'Nasi goreng cabai hijau dengan aneka seafood segar.', 27500, 'fa-bowl-rice', 'bg-emerald-50', NULL,      true, 5),
        (b_samratulangi, cat_nasgor, 'Nasi Goreng Ijo Seafood Ist', 'Nasi goreng cabai hijau seafood porsi istimewa lengkap.', 29000, 'fa-bowl-rice', 'bg-emerald-50', 'Best',    true, 6),
        (b_samratulangi, cat_nasgor, 'Nasi Goreng Sosis',           'Nasi goreng gurih dengan irisan sosis pilihan.', 25500, 'fa-bowl-rice', 'bg-amber-50',   NULL,      true, 7),
        (b_samratulangi, cat_nasgor, 'Nasi Goreng Bakso',           'Nasi goreng bumbu rempah khas dengan irisan bakso.', 25500, 'fa-bowl-rice', 'bg-amber-50',   NULL,      true, 8),
        (b_samratulangi, cat_nasgor, 'Nasi Goreng Mawut',           'Perpaduan nasi dan bakmi goreng khas dengan bumbu spesial.', 29000, 'fa-bowl-rice', 'bg-amber-50',   'Favorit', true, 9),
        (b_samratulangi, cat_nasgor, 'Nasi Goreng Ayam',            'Nasi goreng klasik dengan suwiran daging ayam gurih.', 25500, 'fa-bowl-rice', 'bg-amber-50',   NULL,      true, 10),
        (b_samratulangi, cat_nasgor, 'Nasi Goreng Ayam Ist',        'Nasi goreng ayam porsi istimewa dengan pelengkap spesial.', 27000, 'fa-bowl-rice', 'bg-amber-50',   'Best',    true, 11),
        (b_samratulangi, cat_nasgor, 'Nasi Goreng Kambing',         'Nasi goreng rempah harum dengan potongan daging kambing empuk.', 25500, 'fa-bowl-rice', 'bg-amber-50',   NULL,      true, 12),
        (b_samratulangi, cat_nasgor, 'Nasi Goreng Kambing Ist',     'Nasi goreng kambing porsi istimewa lengkap.', 27000, 'fa-bowl-rice', 'bg-amber-50',   NULL,      true, 13),
        (b_samratulangi, cat_nasgor, 'Nasi Goreng Babat',           'Nasi goreng gurih dengan potongan babat empuk khas Solo.', 25500, 'fa-bowl-rice', 'bg-amber-50',   NULL,      true, 14),
        (b_samratulangi, cat_nasgor, 'Nasi Goreng Babat Ist',       'Nasi goreng babat porsi istimewa lengkap.', 27000, 'fa-bowl-rice', 'bg-amber-50',   'Favorit', true, 15),
        (b_samratulangi, cat_nasgor, 'Nasi Goreng Seafood',         'Nasi goreng spesial dengan campuran udang dan cumi segar.', 27500, 'fa-bowl-rice', 'bg-amber-50',   NULL,      true, 16),
        (b_samratulangi, cat_nasgor, 'Nasi Goreng Seafood Ist',     'Nasi goreng seafood porsi istimewa lengkap.', 29000, 'fa-bowl-rice', 'bg-amber-50',   NULL,      true, 17),
        (b_samratulangi, cat_nasgor, 'Nasi Goreng Udang',           'Nasi goreng gurih dengan udang segar.', 26500, 'fa-bowl-rice', 'bg-amber-50',   NULL,      true, 18),
        (b_samratulangi, cat_nasgor, 'Nasi Goreng Udang Ist',       'Nasi goreng udang porsi istimewa lengkap.', 28000, 'fa-bowl-rice', 'bg-amber-50',   NULL,      true, 19),
        (b_samratulangi, cat_nasgor, 'Nasi Goreng Sapi',            'Nasi goreng dengan irisan daging sapi empuk.', 26500, 'fa-bowl-rice', 'bg-amber-50',   NULL,      true, 20),
        (b_samratulangi, cat_nasgor, 'Nasi Goreng Sapi Ist',        'Nasi goreng daging sapi porsi istimewa lengkap.', 28000, 'fa-bowl-rice', 'bg-amber-50',   NULL,      true, 21),
        (b_samratulangi, cat_nasgor, 'Nasi Goreng Lada Hitam',      'Nasi goreng dengan bumbu saus lada hitam hangat dan harum.', 27000, 'fa-bowl-rice', 'bg-stone-100',  NULL,      true, 22),
        (b_samratulangi, cat_nasgor, 'Nasi Goreng Lada Hitam Ist',  'Nasi goreng lada hitam porsi istimewa lengkap.', 28500, 'fa-bowl-rice', 'bg-stone-100',  NULL,      true, 23);

    -- 2. MASAKAN DAGING AYAM (22 items)
    INSERT INTO menu_items (branch_id, category_id, name, description, price, icon, icon_bg_class, tag, is_available, sort_order)
    VALUES
        (b_samratulangi, cat_ayam, 'Ayam Teriyaki',            'Olahan daging ayam lembut dengan saus teriyaki manis gurih.', 30500, 'fa-drumstick-bite', 'bg-amber-50', NULL,      true, 24),
        (b_samratulangi, cat_ayam, 'Ayam Gongso',              'Ayam tumis bumbu gongso kecap pedas manis khas Solo.', 30500, 'fa-drumstick-bite', 'bg-amber-50', 'Favorit', true, 25),
        (b_samratulangi, cat_ayam, 'Ayam Lada Hitam',          'Potongan daging ayam tumis saus lada hitam pekat.', 30500, 'fa-drumstick-bite', 'bg-amber-50', NULL,      true, 26),
        (b_samratulangi, cat_ayam, 'Ayam Masak Kecap Pedas',   'Ayam masak bumbu kecap manis dengan sensasi pedas segar.', 30500, 'fa-drumstick-bite', 'bg-amber-50', NULL,      true, 27),
        (b_samratulangi, cat_ayam, 'Ayam Goreng Mareme',       'Ayam goreng spesial racikan khas Mareme yang renyah dan gurih.', 30500, 'fa-drumstick-bite', 'bg-amber-50', 'Best',    true, 28),
        (b_samratulangi, cat_ayam, 'Ayam Goreng Asam Manis',   'Ayam goreng tepung renyah disiram saus asam manis segar.', 30500, 'fa-drumstick-bite', 'bg-amber-50', 'Favorit', true, 29),
        (b_samratulangi, cat_ayam, 'Ayam Goreng Mentega',      'Ayam goreng dimasak dengan saus mentega bawang putih harum.', 30500, 'fa-drumstick-bite', 'bg-amber-50', 'Best',    true, 30),
        (b_samratulangi, cat_ayam, 'Ayam Goreng Tepung',       'Potongan ayam goreng krispi berbalut tepung bumbu gurih.', 30500, 'fa-drumstick-bite', 'bg-amber-50', NULL,      true, 31),
        (b_samratulangi, cat_ayam, 'Ayam Goreng Masak Tauco',  'Ayam goreng dimasak dengan bumbu tauco gurih khas.', 30500, 'fa-drumstick-bite', 'bg-amber-50', NULL,      true, 32),
        (b_samratulangi, cat_ayam, 'Ayam Ca Jamur',            'Tumisan daging ayam lembut dengan jamur segar dan saus tiram.', 30500, 'fa-drumstick-bite', 'bg-amber-50', NULL,      true, 33),
        (b_samratulangi, cat_ayam, 'Ayam Ca Kaylan',           'Tumisan daging ayam dengan sayur kaylan hijau segar.', 30500, 'fa-drumstick-bite', 'bg-emerald-50', NULL,    true, 34),
        (b_samratulangi, cat_ayam, 'Ayam Ca Kangkung',         'Tumisan daging ayam dengan kangkung segar bumbu bawang putih.', 30500, 'fa-drumstick-bite', 'bg-emerald-50', NULL,    true, 35),
        (b_samratulangi, cat_ayam, 'Ayam Ca Brokoli',          'Tumisan daging ayam dengan kuntum brokoli hijau renyah.', 30500, 'fa-drumstick-bite', 'bg-emerald-50', NULL,    true, 36),
        (b_samratulangi, cat_ayam, 'Ayam Sayur Asin Ca/Kuah',  'Olahan ayam dengan sawi asin segar, tersedia pilihan ca atau kuah.', 30500, 'fa-drumstick-bite', 'bg-amber-50', NULL,  true, 37),
        (b_samratulangi, cat_ayam, 'Ayam Nanking',             'Hidangan ayam lapis klasik dengan siraman saus kental gurih.', 30500, 'fa-drumstick-bite', 'bg-amber-50', NULL,      true, 38),
        (b_samratulangi, cat_ayam, 'Ayam Rebus',               'Ayam rebus lembut dengan minyak wijen dan bawang putih harum.', 30500, 'fa-drumstick-bite', 'bg-stone-100', NULL,     true, 39),
        (b_samratulangi, cat_ayam, 'Sayap Masak Kecap',        'Sayap ayam empuk dimasak bumbu kecap rempah meresap.', 30500, 'fa-drumstick-bite', 'bg-amber-50', NULL,      true, 40),
        (b_samratulangi, cat_ayam, 'Bestik Ayam',              'Bestik daging ayam dengan kuah saus manis gurih khas Solo.', 30500, 'fa-drumstick-bite', 'bg-amber-50', NULL,      true, 41),
        (b_samratulangi, cat_ayam, 'Ko Loo Kee',               'Ayam goreng tepung krispi dengan saus asam manis nanas.', 30500, 'fa-drumstick-bite', 'bg-amber-50', 'Favorit', true, 42),
        (b_samratulangi, cat_ayam, 'Ngo Yang Kee',             'Ayam olahan bumbu ngohiong rempah lima rasa yang khas.', 30500, 'fa-drumstick-bite', 'bg-amber-50', NULL,      true, 43),
        (b_samratulangi, cat_ayam, 'Pak Lay Ca',               'Tumisan jeroan/ayam dan sayuran segar bumbu gurih.', 30000, 'fa-bowl-food',      'bg-amber-50', NULL,      true, 44),
        (b_samratulangi, cat_ayam, 'Kulit Goreng',             'Kulit ayam goreng krispi dan gurih.', 30000, 'fa-drumstick-bite', 'bg-amber-50', 'Favorit', true, 45);

    -- 3. ANEKA MASAKAN (30 items)
    INSERT INTO menu_items (branch_id, category_id, name, description, price, icon, icon_bg_class, tag, is_available, sort_order)
    VALUES
        (b_samratulangi, cat_masakan, 'Bakmi Goreng',           'Bakmi goreng bumbu khas Mareme dengan sayuran dan daging.', 25000, 'fa-bowl-food', 'bg-amber-50',   'Best',    true, 46),
        (b_samratulangi, cat_masakan, 'Bakmi Godog',            'Bakmi kuah kaldu gurih hangat dengan telur dan sayuran.', 25000, 'fa-bowl-food', 'bg-amber-50',   NULL,      true, 47),
        (b_samratulangi, cat_masakan, 'Bihun Goreng',           'Bihun goreng lembut berbumbu bawang putih dan kecap gurih.', 25000, 'fa-bowl-food', 'bg-amber-50',   NULL,      true, 48),
        (b_samratulangi, cat_masakan, 'Bihun Godog',            'Bihun kuah kaldu hangat dengan sayuran segar.', 25000, 'fa-bowl-food', 'bg-amber-50',   NULL,      true, 49),
        (b_samratulangi, cat_masakan, 'Capjay Goreng',          'Tumisan aneka sayuran segar dengan saus kental gurih.', 26500, 'fa-bowl-food', 'bg-emerald-50', 'Favorit', true, 50),
        (b_samratulangi, cat_masakan, 'Capjay Godog',           'Capjay kuah hangat dengan aneka sayuran segar.', 26500, 'fa-bowl-food', 'bg-emerald-50', NULL,      true, 51),
        (b_samratulangi, cat_masakan, 'Nasi Ayam Ca',           'Nasi putih hangat disajikan dengan tumisan ayam ca gurih.', 27500, 'fa-bowl-rice', 'bg-amber-50',   NULL,      true, 52),
        (b_samratulangi, cat_masakan, 'Nasi Campur',            'Nasi dengan aneka lauk pendamping spesial khas Mareme.', 28000, 'fa-bowl-rice', 'bg-amber-50',   NULL,      true, 53),
        (b_samratulangi, cat_masakan, 'Nasi Putih',             'Nasi putih pulen hangat per porsi.', 5000,  'fa-bowl-rice', 'bg-stone-100',  NULL,      true, 54),
        (b_samratulangi, cat_masakan, 'Masak Tahu',             'Olahan tahu lembut dengan sayuran dan bumbu gurih.', 28000, 'fa-bowl-food', 'bg-amber-50',   NULL,      true, 55),
        (b_samratulangi, cat_masakan, 'Tahu Masak Kecap',       'Tahu goreng lembut dimasak dengan bumbu kecap manis gurih.', 28500, 'fa-bowl-food', 'bg-amber-50',   NULL,      true, 56),
        (b_samratulangi, cat_masakan, 'Misoa Kuah',             'Misoa lembut dalam kuah kaldu hangat yang menenangkan.', 27500, 'fa-bowl-food', 'bg-amber-50',   NULL,      true, 57),
        (b_samratulangi, cat_masakan, 'Mun Tahu',               'Tahu sutra cincang dalam kuah kental gurih hangat.', 28500, 'fa-bowl-food', 'bg-amber-50',   NULL,      true, 58),
        (b_samratulangi, cat_masakan, 'Angsio Tahu',            'Tahu Jepang lembut dimasak saus angsio kecokelatan yang lezat.', 28500, 'fa-bowl-food', 'bg-amber-50',   NULL,      true, 59),
        (b_samratulangi, cat_masakan, 'Sapo Tahu',              'Tahu Jepang, sayuran, dan seafood dalam kuah sapo kental hangat.', 33000, 'fa-bowl-food', 'bg-amber-50',   'Best',    true, 60),
        (b_samratulangi, cat_masakan, 'Soon Goreng',            'Soun goreng kenyal berbumbu gurih dengan sayuran.', 27000, 'fa-bowl-food', 'bg-amber-50',   NULL,      true, 61),
        (b_samratulangi, cat_masakan, 'Ta Mie Goreng',          'Mie kering renyah disiram kuah kental sayuran dan daging.', 26500, 'fa-bowl-food', 'bg-amber-50',   NULL,      true, 62),
        (b_samratulangi, cat_masakan, 'Yam Yien',               'Mie olahan khas dengan bumbu gurih.', 27000, 'fa-bowl-food', 'bg-amber-50',   NULL,      true, 63),
        (b_samratulangi, cat_masakan, 'Aneka Sayur',            'Tumisan aneka sayuran hijau segar.', 24000, 'fa-leaf',      'bg-emerald-50', NULL,      true, 64),
        (b_samratulangi, cat_masakan, 'Fuyung Hay',             'Telur dadar tebal berisi sayuran dan daging dengan saus asam manis.', 27500, 'fa-bowl-food', 'bg-amber-50',   'Favorit', true, 65),
        (b_samratulangi, cat_masakan, 'Kay Lan Ca Polos',       'Tumis sayur kaylan segar dengan bawang putih.', 24000, 'fa-leaf',      'bg-emerald-50', NULL,      true, 66),
        (b_samratulangi, cat_masakan, 'Kangkung Ca Polos',      'Tumis kangkung segar bawang putih dengan aroma wok.', 24000, 'fa-leaf',      'bg-emerald-50', NULL,      true, 67),
        (b_samratulangi, cat_masakan, 'Kwee Tiauw Goreng/Kuah', 'Kwetiau lembut dimasak goreng atau kuah sesuai selera.', 25500, 'fa-bowl-food', 'bg-amber-50',   NULL,      true, 68),
        (b_samratulangi, cat_masakan, 'Kwee Tiauw Nyemek',      'Kwetiau dengan kuah nyemek kental gurih khas.', 25500, 'fa-bowl-food', 'bg-amber-50',   'Favorit', true, 69),
        (b_samratulangi, cat_masakan, 'Kamar Bola',             'Olahan bola daging/udang goreng dengan saus khas.', 27000, 'fa-bowl-food', 'bg-amber-50',   NULL,      true, 70),
        (b_samratulangi, cat_masakan, 'Loo Mie',                'Mie tebal dalam kuah kental gurih kaya rasa.', 27000, 'fa-bowl-food', 'bg-amber-50',   NULL,      true, 71),
        (b_samratulangi, cat_masakan, 'Brokoli Polos',          'Tumis brokoli hijau segar dengan saus bawang putih.', 24500, 'fa-leaf',      'bg-emerald-50', NULL,      true, 72),
        (b_samratulangi, cat_masakan, 'Babat Gongso',           'Babat empuk dimasak bumbu gongso kecap pedas manis.', 28500, 'fa-bowl-food', 'bg-amber-50',   'Favorit', true, 73),
        (b_samratulangi, cat_masakan, 'Orak Arik',              'Tumisan telur orak-arik dengan sayuran segar gurih.', 28500, 'fa-bowl-food', 'bg-amber-50',   NULL,      true, 74),
        (b_samratulangi, cat_masakan, 'Nasi Capjay',            'Nasi putih hangat disajikan bersama capjay sayuran segar.', 25500, 'fa-bowl-rice', 'bg-emerald-50', NULL,      true, 75);

    -- 4. ANEKA SOUP (5 items)
    INSERT INTO menu_items (branch_id, category_id, name, description, price, icon, icon_bg_class, tag, is_available, sort_order)
    VALUES
        (b_samratulangi, cat_soup, 'Soup Asparagus Kepiting', 'Sup kental hangat perpaduan asparagus lembut dan daging kepiting.', 43500, 'fa-bowl-food', 'bg-amber-50', 'Best', true, 76),
        (b_samratulangi, cat_soup, 'Soup ShangHay',           'Sup hangat ala Shanghai dengan isian lengkap dan kuah gurih.', 34500, 'fa-bowl-food', 'bg-amber-50', NULL,   true, 77),
        (b_samratulangi, cat_soup, 'Soup Ayam',               'Sup kaldu ayam bening hangat yang gurih dan menyegarkan.', 30500, 'fa-bowl-food', 'bg-amber-50', NULL,   true, 78),
        (b_samratulangi, cat_soup, 'Soup Tahu Sayur Asin',    'Sup segar berisi tahu lembut dan sawi asin pilihan.', 30500, 'fa-bowl-food', 'bg-amber-50', NULL,   true, 79),
        (b_samratulangi, cat_soup, 'Soup Sapi',               'Sup irisan daging sapi empuk dalam kuah kaldu hangat.', 31000, 'fa-bowl-food', 'bg-amber-50', NULL,   true, 80);

    -- 5. MASAKAN TELUR ASIN (4 items)
    INSERT INTO menu_items (branch_id, category_id, name, description, price, icon, icon_bg_class, tag, is_available, sort_order)
    VALUES
        (b_samratulangi, cat_telurasin, 'Ayam Telur Asin',   'Ayam goreng tepung renyah dibalut saus telur asin creamy gurih.', 37500, 'fa-drumstick-bite', 'bg-amber-50', 'Best',    true, 81),
        (b_samratulangi, cat_telurasin, 'Cumi Telur Asin',   'Cumi goreng krispi dengan balutan saus telur asin gurih.', 37500, 'fa-bowl-food',      'bg-amber-50', 'Favorit', true, 82),
        (b_samratulangi, cat_telurasin, 'Udang Telur Asin',  'Udang goreng renyah dimasak dengan saus kuning telur asin.', 37500, 'fa-bowl-food',      'bg-amber-50', NULL,      true, 83),
        (b_samratulangi, cat_telurasin, 'Gurame Telur Asin', 'Fillet ikan gurame krispi berbalut saus telur asin spesial.', 37500, 'fa-fish',           'bg-amber-50', NULL,      true, 84);
END $$;
