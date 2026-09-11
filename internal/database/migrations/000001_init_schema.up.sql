-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Sequence for order numbers (daily or sequential)
CREATE SEQUENCE IF NOT EXISTS order_number_seq START 1;

-- Users table
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    phone VARCHAR(30) UNIQUE NOT NULL,
    name VARCHAR(100) NOT NULL,
    role VARCHAR(20) NOT NULL DEFAULT 'customer', -- customer, branch_admin, owner
    branch_id UUID,
    password_hash VARCHAR(255), -- only for admin/owner
    address TEXT,
    latitude DOUBLE PRECISION,
    longitude DOUBLE PRECISION,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Branches table
CREATE TABLE IF NOT EXISTS branches (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slug VARCHAR(50) UNIQUE NOT NULL,
    name VARCHAR(100) NOT NULL,
    address TEXT NOT NULL,
    phone VARCHAR(30) NOT NULL,
    latitude DOUBLE PRECISION NOT NULL,
    longitude DOUBLE PRECISION NOT NULL,
    gradient_theme VARCHAR(50) NOT NULL DEFAULT 'brand', -- brand, emerald, indigo
    icon VARCHAR(50) NOT NULL DEFAULT 'building',
    facility_tags JSONB NOT NULL DEFAULT '[]'::jsonb,
    rating DOUBLE PRECISION NOT NULL DEFAULT 4.8,
    is_open BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Branch settings table
CREATE TABLE IF NOT EXISTS branch_settings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    branch_id UUID UNIQUE NOT NULL REFERENCES branches(id) ON DELETE CASCADE,
    operating_hours JSONB NOT NULL DEFAULT '{"weekday": {"open": "08:00", "close": "22:00"}, "weekend": {"open": "09:00", "close": "21:00"}}'::jsonb,
    max_delivery_radius_km INT NOT NULL DEFAULT 10,
    base_delivery_fee_near INT NOT NULL DEFAULT 8000,
    base_delivery_fee_mid INT NOT NULL DEFAULT 12000,
    base_delivery_fee_far INT NOT NULL DEFAULT 18000,
    near_threshold_km INT NOT NULL DEFAULT 3,
    mid_threshold_km INT NOT NULL DEFAULT 7,
    service_fee INT NOT NULL DEFAULT 2000,
    min_order_amount INT NOT NULL DEFAULT 20000,
    free_delivery_threshold INT NOT NULL DEFAULT 150000,
    whatsapp_number VARCHAR(30) NOT NULL DEFAULT '081234567890',
    description TEXT,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Categories table
CREATE TABLE IF NOT EXISTS categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    slug VARCHAR(50) UNIQUE NOT NULL,
    emoji VARCHAR(20) NOT NULL,
    sort_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Menu items table
CREATE TABLE IF NOT EXISTS menu_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    branch_id UUID NOT NULL REFERENCES branches(id) ON DELETE CASCADE,
    category_id UUID NOT NULL REFERENCES categories(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    description TEXT NOT NULL,
    price INT NOT NULL, -- in Rupiah
    icon VARCHAR(50) NOT NULL DEFAULT 'fa-mug-hot',
    icon_bg_class VARCHAR(50) NOT NULL DEFAULT 'bg-amber-50',
    image_url TEXT,
    tag VARCHAR(50), -- e.g. "Best", "Favorit"
    is_available BOOLEAN NOT NULL DEFAULT true,
    sort_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Orders table
CREATE TABLE IF NOT EXISTS orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_number VARCHAR(50) UNIQUE NOT NULL,
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    branch_id UUID NOT NULL REFERENCES branches(id) ON DELETE RESTRICT,
    order_type VARCHAR(20) NOT NULL DEFAULT 'delivery', -- delivery, pickup, scheduled
    status VARCHAR(30) NOT NULL DEFAULT 'pending', -- pending, accepted, preparing, ready, on_the_way, delivered, completed, rejected, cancelled
    customer_name VARCHAR(100) NOT NULL,
    customer_phone VARCHAR(30) NOT NULL,
    delivery_address TEXT,
    delivery_notes TEXT,
    delivery_lat DOUBLE PRECISION,
    delivery_lon DOUBLE PRECISION,
    delivery_distance_km DOUBLE PRECISION DEFAULT 0,
    subtotal INT NOT NULL DEFAULT 0,
    delivery_fee INT NOT NULL DEFAULT 0,
    service_fee INT NOT NULL DEFAULT 2000,
    discount INT NOT NULL DEFAULT 0,
    grand_total INT NOT NULL DEFAULT 0,
    promo_code VARCHAR(50),
    scheduled_at TIMESTAMP WITH TIME ZONE,
    driver_name VARCHAR(100) DEFAULT 'Andi Pratama',
    driver_phone VARCHAR(30) DEFAULT '081234567890',
    driver_vehicle VARCHAR(100) DEFAULT 'Honda Vario 160 Hitam',
    driver_plate VARCHAR(30) DEFAULT 'B 1234 ABC',
    driver_rating DOUBLE PRECISION DEFAULT 4.95,
    rejection_reason TEXT,
    version INT NOT NULL DEFAULT 1,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Order items table
CREATE TABLE IF NOT EXISTS order_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    menu_item_id UUID REFERENCES menu_items(id) ON DELETE SET NULL,
    item_name VARCHAR(100) NOT NULL,
    item_price INT NOT NULL,
    item_icon VARCHAR(50) NOT NULL,
    quantity INT NOT NULL DEFAULT 1,
    notes TEXT,
    line_total INT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Payments table
CREATE TABLE IF NOT EXISTS payments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID UNIQUE NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    midtrans_order_id VARCHAR(100) UNIQUE NOT NULL,
    payment_method VARCHAR(50) NOT NULL, -- qris, gopay, shopeepay
    payment_type VARCHAR(50) NOT NULL DEFAULT 'qris',
    status VARCHAR(30) NOT NULL DEFAULT 'pending', -- pending, settlement, expire, cancel, deny, refund
    amount INT NOT NULL,
    idempotency_key VARCHAR(100) UNIQUE NOT NULL,
    snap_token VARCHAR(255),
    snap_redirect_url TEXT,
    qr_string TEXT,
    midtrans_transaction_id VARCHAR(100),
    midtrans_response JSONB,
    paid_at TIMESTAMP WITH TIME ZONE,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Promos table
CREATE TABLE IF NOT EXISTS promos (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code VARCHAR(50) UNIQUE NOT NULL,
    type VARCHAR(30) NOT NULL, -- fixed, free_delivery
    discount_amount INT DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT true,
    min_spend INT DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Order status history
CREATE TABLE IF NOT EXISTS order_status_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    from_status VARCHAR(30),
    to_status VARCHAR(30) NOT NULL,
    changed_by VARCHAR(50) DEFAULT 'system',
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for lightning fast queries
CREATE INDEX IF NOT EXISTS idx_orders_branch_id ON orders(branch_id);
CREATE INDEX IF NOT EXISTS idx_orders_user_id ON orders(user_id);
CREATE INDEX IF NOT EXISTS idx_orders_status ON orders(status);
CREATE INDEX IF NOT EXISTS idx_orders_created_at ON orders(created_at);
CREATE INDEX IF NOT EXISTS idx_menu_items_branch ON menu_items(branch_id);
CREATE INDEX IF NOT EXISTS idx_menu_items_category ON menu_items(category_id);
CREATE INDEX IF NOT EXISTS idx_payments_order_id ON payments(order_id);
CREATE INDEX IF NOT EXISTS idx_payments_status ON payments(status);
