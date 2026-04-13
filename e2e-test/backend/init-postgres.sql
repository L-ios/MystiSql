-- MystiSql E2E Test Database Initialization Script for PostgreSQL 14

-- Create users table
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) NOT NULL UNIQUE,
    email VARCHAR(100) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    full_name VARCHAR(100),
    age INTEGER,
    balance DECIMAL(10, 2) DEFAULT 0.00,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    metadata JSONB,
    CONSTRAINT chk_age_positive CHECK (age >= 0 AND age <= 150)
);

CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_created_at ON users(created_at);

-- Create products table
CREATE TABLE IF NOT EXISTS products (
    id SERIAL PRIMARY KEY,
    name VARCHAR(200) NOT NULL,
    description TEXT,
    price DECIMAL(10, 2) NOT NULL,
    stock INTEGER DEFAULT 0,
    category VARCHAR(50),
    is_available BOOLEAN DEFAULT TRUE,
    tags TEXT[],
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_price_positive CHECK (price >= 0),
    CONSTRAINT chk_stock_non_negative CHECK (stock >= 0)
);

CREATE INDEX idx_products_name ON products(name);
CREATE INDEX idx_products_category ON products(category);
CREATE INDEX idx_products_price ON products(price);

-- Create orders table
CREATE TABLE IF NOT EXISTS orders (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    order_number VARCHAR(50) NOT NULL UNIQUE,
    total_amount DECIMAL(10, 2) NOT NULL,
    status VARCHAR(20) DEFAULT 'pending' CHECK (status IN ('pending', 'confirmed', 'shipped', 'delivered', 'cancelled')),
    notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_orders_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX idx_orders_user_id ON orders(user_id);
CREATE INDEX idx_orders_order_number ON orders(order_number);
CREATE INDEX idx_orders_status ON orders(status);
CREATE INDEX idx_orders_created_at ON orders(created_at);

-- Create audit_log table
CREATE TABLE IF NOT EXISTS audit_log (
    id SERIAL PRIMARY KEY,
    table_name VARCHAR(50) NOT NULL,
    record_id BIGINT NOT NULL,
    action VARCHAR(10) NOT NULL CHECK (action IN ('INSERT', 'UPDATE', 'DELETE')),
    old_values JSONB,
    new_values JSONB,
    user_id INTEGER,
    ip_address INET,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_audit_log_table_name ON audit_log(table_name);
CREATE INDEX idx_audit_log_record_id ON audit_log(record_id);
CREATE INDEX idx_audit_log_action ON audit_log(action);
CREATE INDEX idx_audit_log_created_at ON audit_log(created_at);

-- Create a trigger for updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_products_updated_at BEFORE UPDATE ON products
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_orders_updated_at BEFORE UPDATE ON orders
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Insert test users
INSERT INTO users (username, email, password_hash, full_name, age, balance, is_active, metadata) VALUES
('alice', 'alice@example.com', '$2a$10$XOPbrlUPQdwdJUpSrIF6X.LbE14qsMmKGhM1A8W9iqaG3vv1BD7WC', 'Alice Johnson', 28, 1500.50, TRUE, '{"role": "admin", "department": "IT"}'::jsonb),
('bob', 'bob@example.com', '$2a$10$XOPbrlUPQdwdJUpSrIF6X.LbE14qsMmKGhM1A8W9iqaG3vv1BD7WC', 'Bob Smith', 35, 2350.75, TRUE, '{"role": "user", "department": "Sales"}'::jsonb),
('charlie', 'charlie@example.com', '$2a$10$XOPbrlUPQdwdJUpSrIF6X.LbE14qsMmKGhM1A8W9iqaG3vv1BD7WC', 'Charlie Brown', 42, 890.00, TRUE, '{"role": "user", "department": "Marketing"}'::jsonb),
('diana', 'diana@example.com', '$2a$10$XOPbrlUPQdwdJUpSrIF6X.LbE14qsMmKGhM1A8W9iqaG3vv1BD7WC', 'Diana Prince', 31, 4200.00, TRUE, '{"role": "admin", "department": "HR"}'::jsonb),
('eve', 'eve@example.com', '$2a$10$XOPbrlUPQdwdJUpSrIF6X.LbE14qsMmKGhM1A8W9iqaG3vv1BD7WC', 'Eve Wilson', 26, 1100.25, FALSE, '{"role": "user", "department": "Support"}'::jsonb),
('frank', 'frank@example.com', '$2a$10$XOPbrlUPQdwdJUpSrIF6X.LbE14qsMmKGhM1A8W9iqaG3vv1BD7WC', 'Frank Miller', 38, 5500.90, TRUE, '{"role": "manager", "department": "Engineering"}'::jsonb),
('grace', 'grace@example.com', '$2a$10$XOPbrlUPQdwdJUpSrIF6X.LbE14qsMmKGhM1A8W9iqaG3vv1BD7WC', 'Grace Lee', 29, 3200.00, TRUE, '{"role": "user", "department": "Finance"}'::jsonb),
('henry', 'henry@example.com', '$2a$10$XOPbrlUPQdwdJUpSrIF6X.LbE14qsMmKGhM1A8W9iqaG3vv1BD7WC', 'Henry Davis', 45, 7800.50, TRUE, '{"role": "admin", "department": "Operations"}'::jsonb),
('ivy', 'ivy@example.com', '$2a$10$XOPbrlUPQdwdJUpSrIF6X.LbE14qsMmKGhM1A8W9iqaG3vv1BD7WC', 'Ivy Chen', 33, 2900.75, TRUE, '{"role": "user", "department": "Design"}'::jsonb),
('jack', 'jack@example.com', '$2a$10$XOPbrlUPQdwdJUpSrIF6X.LbE14qsMmKGhM1A8W9iqaG3vv1BD7WC', 'Jack Taylor', 27, 1800.00, TRUE, '{"role": "user", "department": "IT"}'::jsonb);

-- Insert test products with array tags
INSERT INTO products (name, description, price, stock, category, is_available, tags) VALUES
('Laptop Pro 15', 'High-performance laptop for professionals', 1299.99, 50, 'Electronics', TRUE, ARRAY['computer', 'portable', 'work']),
('Wireless Mouse', 'Ergonomic wireless mouse with precision tracking', 49.99, 200, 'Accessories', TRUE, ARRAY['input', 'wireless']),
('USB-C Hub', 'Multi-port USB-C hub with HDMI and SD card reader', 79.99, 150, 'Accessories', TRUE, ARRAY['adapter', 'usb-c']),
('Mechanical Keyboard', 'RGB mechanical keyboard with blue switches', 129.99, 100, 'Accessories', TRUE, ARRAY['input', 'gaming', 'rgb']),
('4K Monitor', '27-inch 4K UHD monitor with HDR support', 449.99, 75, 'Electronics', TRUE, ARRAY['display', '4k']),
('Webcam HD', '1080p HD webcam with built-in microphone', 89.99, 180, 'Accessories', TRUE, ARRAY['video', 'streaming']),
('Wireless Earbuds', 'True wireless earbuds with active noise cancellation', 199.99, 120, 'Audio', TRUE, ARRAY['audio', 'wireless', 'bluetooth']),
('Bluetooth Speaker', 'Portable Bluetooth speaker with 360° sound', 149.99, 90, 'Audio', TRUE, ARRAY['audio', 'bluetooth', 'portable']),
('Gaming Chair', 'Ergonomic gaming chair with lumbar support', 349.99, 60, 'Furniture', TRUE, ARRAY['furniture', 'gaming', 'ergonomic']),
('Desk Lamp', 'LED desk lamp with adjustable brightness', 39.99, 220, 'Furniture', TRUE, ARRAY['lighting', 'led']),
('External SSD 1TB', 'Portable external SSD with USB 3.2', 119.99, 140, 'Storage', TRUE, ARRAY['storage', 'ssd', 'portable']),
('Graphics Tablet', 'Digital drawing tablet with stylus', 249.99, 85, 'Accessories', TRUE, ARRAY['input', 'design', 'digital-art']);

-- Insert test orders
INSERT INTO orders (user_id, order_number, total_amount, status, notes) VALUES
(1, 'ORD-2024-001', 1379.98, 'delivered', 'Express delivery requested'),
(2, 'ORD-2024-002', 229.98, 'shipped', NULL),
(3, 'ORD-2024-003', 449.99, 'confirmed', 'Gift wrap needed'),
(1, 'ORD-2024-004', 199.99, 'pending', NULL),
(4, 'ORD-2024-005', 599.98, 'delivered', 'Leave at door'),
(5, 'ORD-2024-006', 89.99, 'cancelled', 'Customer request'),
(6, 'ORD-2024-007', 1299.99, 'shipped', 'Insurance included'),
(7, 'ORD-2024-008', 179.98, 'confirmed', NULL),
(8, 'ORD-2024-009', 349.99, 'pending', 'Delivery before weekend'),
(9, 'ORD-2024-010', 149.99, 'delivered', NULL);

-- Insert some audit log entries
INSERT INTO audit_log (table_name, record_id, action, old_values, new_values, user_id, ip_address) VALUES
('users', 1, 'INSERT', NULL, '{"username": "alice", "email": "alice@example.com"}'::jsonb, 1, '192.168.1.100'::inet),
('users', 2, 'UPDATE', '{"balance": 1500.00}'::jsonb, '{"balance": 2350.75}'::jsonb, 2, '192.168.1.101'::inet),
('products', 1, 'INSERT', NULL, '{"name": "Laptop Pro 15", "price": 1299.99}'::jsonb, 1, '192.168.1.100'::inet),
('orders', 1, 'UPDATE', '{"status": "pending"}'::jsonb, '{"status": "confirmed"}'::jsonb, 1, '192.168.1.102'::inet),
('users', 5, 'UPDATE', '{"is_active": true}'::jsonb, '{"is_active": false}'::jsonb, 4, '192.168.1.105'::inet);

-- Create a view for testing
CREATE OR REPLACE VIEW active_users AS
SELECT id, username, email, full_name, balance
FROM users
WHERE is_active = TRUE;

-- Create a function for testing
CREATE OR REPLACE FUNCTION get_user_orders(p_user_id INTEGER)
RETURNS TABLE (
    id INTEGER,
    order_number VARCHAR(50),
    total_amount DECIMAL(10, 2),
    status VARCHAR(20),
    created_at TIMESTAMP
) AS $$
BEGIN
    RETURN QUERY
    SELECT o.id, o.order_number, o.total_amount, o.status, o.created_at
    FROM orders o
    WHERE o.user_id = p_user_id
    ORDER BY o.created_at DESC;
END;
$$ LANGUAGE plpgsql;

-- Create a CTE example view
CREATE OR REPLACE VIEW order_summary AS
WITH user_order_stats AS (
    SELECT 
        u.id AS user_id,
        u.username,
        COUNT(o.id) AS total_orders,
        COALESCE(SUM(o.total_amount), 0) AS total_spent,
        COALESCE(AVG(o.total_amount), 0) AS avg_order_value
    FROM users u
    LEFT JOIN orders o ON u.id = o.user_id
    GROUP BY u.id, u.username
)
SELECT * FROM user_order_stats;
