-- 0001_init.sqlite.sql
-- This is for documentation purposes only. The actual schema is created by GORM AutoMigrate.

CREATE TABLE stores (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name VARCHAR(200) NOT NULL,
    slug VARCHAR(100) UNIQUE NOT NULL,
    bot_token VARCHAR(255),
    currency VARCHAR(3) DEFAULT 'BOB',
    qr_image_url TEXT,
    owner_telegram_id BIGINT,
    settings JSON,
    is_active BOOLEAN DEFAULT 1,
    created_at DATETIME,
    updated_at DATETIME
);

CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    store_id INTEGER REFERENCES stores(id),
    name VARCHAR(150) NOT NULL,
    email VARCHAR(150) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(50) DEFAULT 'store_owner',
    phone VARCHAR(20),
    created_at DATETIME,
    updated_at DATETIME
);

CREATE TABLE delivery_zones (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    store_id INTEGER REFERENCES stores(id),
    name VARCHAR(100) NOT NULL,
    price DECIMAL(12, 2) DEFAULT 0,
    estimated_time VARCHAR(50),
    is_active BOOLEAN DEFAULT 1
);

CREATE TABLE products (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    store_id INTEGER REFERENCES stores(id),
    name VARCHAR(200) NOT NULL,
    description TEXT,
    base_price DECIMAL(12, 2) NOT NULL,
    category VARCHAR(100),
    status VARCHAR(20) DEFAULT 'active',
    image_url TEXT,
    created_at DATETIME,
    updated_at DATETIME
);

CREATE TABLE product_variants (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    product_id INTEGER REFERENCES products(id),
    sku VARCHAR(50),
    size VARCHAR(50),
    color VARCHAR(50),
    additional_price DECIMAL(12, 2) DEFAULT 0,
    stock_quantity INTEGER DEFAULT 0,
    created_at DATETIME
);

CREATE TABLE telegram_identities (
    telegram_id BIGINT PRIMARY KEY,
    first_name VARCHAR(100),
    username VARCHAR(100),
    language_code VARCHAR(10),
    created_at DATETIME
);

CREATE TABLE customers (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    store_id INTEGER REFERENCES stores(id),
    telegram_id BIGINT REFERENCES telegram_identities(telegram_id),
    phone VARCHAR(30),
    email VARCHAR(150),
    full_name VARCHAR(200),
    address TEXT,
    ltv DECIMAL(12, 2) DEFAULT 0,
    status VARCHAR(20) DEFAULT 'lead',
    created_at DATETIME
);

CREATE TABLE chat_sessions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    store_id INTEGER REFERENCES stores(id),
    customer_id INTEGER REFERENCES customers(id),
    current_flow VARCHAR(50),
    current_step VARCHAR(50),
    context_data JSON,
    last_interaction_at DATETIME,
    is_active BOOLEAN DEFAULT 1
);

CREATE TABLE orders (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    store_id INTEGER REFERENCES stores(id),
    customer_id INTEGER REFERENCES customers(id),
    status VARCHAR(50) DEFAULT 'draft',
    payment_method VARCHAR(50),
    subtotal DECIMAL(12, 2) NOT NULL DEFAULT 0,
    delivery_cost DECIMAL(12, 2) DEFAULT 0,
    total DECIMAL(12, 2) NOT NULL DEFAULT 0,
    delivery_zone_id INTEGER REFERENCES delivery_zones(id),
    shipping_address TEXT,
    payment_proof_url TEXT,
    notes TEXT,
    created_at DATETIME,
    updated_at DATETIME
);

CREATE TABLE order_items (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    order_id INTEGER REFERENCES orders(id),
    product_variant_id INTEGER REFERENCES product_variants(id),
    product_name VARCHAR(200),
    variant_name VARCHAR(100),
    quantity INTEGER NOT NULL,
    unit_price DECIMAL(12, 2) NOT NULL,
    total_line DECIMAL(12, 2) NOT NULL
);

CREATE TABLE ai_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    store_id INTEGER REFERENCES stores(id),
    chat_session_id INTEGER REFERENCES chat_sessions(id),
    input_tokens INTEGER,
    output_tokens INTEGER,
    model_used VARCHAR(50),
    user_message TEXT,
    ai_response TEXT,
    sentiment_score DECIMAL(3, 2),
    created_at DATETIME
);

CREATE TABLE notifications (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    store_id INTEGER REFERENCES stores(id),
    order_id INTEGER REFERENCES orders(id),
    type VARCHAR(50),
    message TEXT,
    is_read BOOLEAN DEFAULT 0,
    created_at DATETIME
);
