-- MystiSql E2E Test Database Initialization Script for MySQL 8

-- Create users table
CREATE TABLE IF NOT EXISTS users (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    username VARCHAR(50) NOT NULL UNIQUE,
    email VARCHAR(100) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    full_name VARCHAR(100),
    age INT,
    balance DECIMAL(10, 2) DEFAULT 0.00,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    metadata JSON,
    INDEX idx_username (username),
    INDEX idx_email (email),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Create products table
CREATE TABLE IF NOT EXISTS products (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(200) NOT NULL,
    description TEXT,
    price DECIMAL(10, 2) NOT NULL,
    stock INT DEFAULT 0,
    category VARCHAR(50),
    is_available BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_name (name),
    INDEX idx_category (category),
    INDEX idx_price (price)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Create orders table
CREATE TABLE IF NOT EXISTS orders (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id BIGINT NOT NULL,
    order_number VARCHAR(50) NOT NULL UNIQUE,
    total_amount DECIMAL(10, 2) NOT NULL,
    status ENUM('pending', 'confirmed', 'shipped', 'delivered', 'cancelled') DEFAULT 'pending',
    notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    INDEX idx_user_id (user_id),
    INDEX idx_order_number (order_number),
    INDEX idx_status (status),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Create audit_log table
CREATE TABLE IF NOT EXISTS audit_log (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    table_name VARCHAR(50) NOT NULL,
    record_id BIGINT NOT NULL,
    action ENUM('INSERT', 'UPDATE', 'DELETE') NOT NULL,
    old_values JSON,
    new_values JSON,
    user_id BIGINT,
    ip_address VARCHAR(45),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_table_name (table_name),
    INDEX idx_record_id (record_id),
    INDEX idx_action (action),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Insert test users
INSERT INTO users (username, email, password_hash, full_name, age, balance, is_active, metadata) VALUES
('alice', 'alice@example.com', '$2a$10$XOPbrlUPQdwdJUpSrIF6X.LbE14qsMmKGhM1A8W9iqaG3vv1BD7WC', 'Alice Johnson', 28, 1500.50, TRUE, '{"role": "admin", "department": "IT"}'),
('bob', 'bob@example.com', '$2a$10$XOPbrlUPQdwdJUpSrIF6X.LbE14qsMmKGhM1A8W9iqaG3vv1BD7WC', 'Bob Smith', 35, 2350.75, TRUE, '{"role": "user", "department": "Sales"}'),
('charlie', 'charlie@example.com', '$2a$10$XOPbrlUPQdwdJUpSrIF6X.LbE14qsMmKGhM1A8W9iqaG3vv1BD7WC', 'Charlie Brown', 42, 890.00, TRUE, '{"role": "user", "department": "Marketing"}'),
('diana', 'diana@example.com', '$2a$10$XOPbrlUPQdwdJUpSrIF6X.LbE14qsMmKGhM1A8W9iqaG3vv1BD7WC', 'Diana Prince', 31, 4200.00, TRUE, '{"role": "admin", "department": "HR"}'),
('eve', 'eve@example.com', '$2a$10$XOPbrlUPQdwdJUpSrIF6X.LbE14qsMmKGhM1A8W9iqaG3vv1BD7WC', 'Eve Wilson', 26, 1100.25, FALSE, '{"role": "user", "department": "Support"}'),
('frank', 'frank@example.com', '$2a$10$XOPbrlUPQdwdJUpSrIF6X.LbE14qsMmKGhM1A8W9iqaG3vv1BD7WC', 'Frank Miller', 38, 5500.90, TRUE, '{"role": "manager", "department": "Engineering"}'),
('grace', 'grace@example.com', '$2a$10$XOPbrlUPQdwdJUpSrIF6X.LbE14qsMmKGhM1A8W9iqaG3vv1BD7WC', 'Grace Lee', 29, 3200.00, TRUE, '{"role": "user", "department": "Finance"}'),
('henry', 'henry@example.com', '$2a$10$XOPbrlUPQdwdJUpSrIF6X.LbE14qsMmKGhM1A8W9iqaG3vv1BD7WC', 'Henry Davis', 45, 7800.50, TRUE, '{"role": "admin", "department": "Operations"}'),
('ivy', 'ivy@example.com', '$2a$10$XOPbrlUPQdwdJUpSrIF6X.LbE14qsMmKGhM1A8W9iqaG3vv1BD7WC', 'Ivy Chen', 33, 2900.75, TRUE, '{"role": "user", "department": "Design"}'),
('jack', 'jack@example.com', '$2a$10$XOPbrlUPQdwdJUpSrIF6X.LbE14qsMmKGhM1A8W9iqaG3vv1BD7WC', 'Jack Taylor', 27, 1800.00, TRUE, '{"role": "user", "department": "IT"}');

-- Insert test products
INSERT INTO products (name, description, price, stock, category, is_available) VALUES
('Laptop Pro 15', 'High-performance laptop for professionals', 1299.99, 50, 'Electronics', TRUE),
('Wireless Mouse', 'Ergonomic wireless mouse with precision tracking', 49.99, 200, 'Accessories', TRUE),
('USB-C Hub', 'Multi-port USB-C hub with HDMI and SD card reader', 79.99, 150, 'Accessories', TRUE),
('Mechanical Keyboard', 'RGB mechanical keyboard with blue switches', 129.99, 100, 'Accessories', TRUE),
('4K Monitor', '27-inch 4K UHD monitor with HDR support', 449.99, 75, 'Electronics', TRUE),
('Webcam HD', '1080p HD webcam with built-in microphone', 89.99, 180, 'Accessories', TRUE),
('Wireless Earbuds', 'True wireless earbuds with active noise cancellation', 199.99, 120, 'Audio', TRUE),
('Bluetooth Speaker', 'Portable Bluetooth speaker with 360° sound', 149.99, 90, 'Audio', TRUE),
('Gaming Chair', 'Ergonomic gaming chair with lumbar support', 349.99, 60, 'Furniture', TRUE),
('Desk Lamp', 'LED desk lamp with adjustable brightness', 39.99, 220, 'Furniture', TRUE),
('External SSD 1TB', 'Portable external SSD with USB 3.2', 119.99, 140, 'Storage', TRUE),
('Graphics Tablet', 'Digital drawing tablet with stylus', 249.99, 85, 'Accessories', TRUE);

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
('users', 1, 'INSERT', NULL, '{"username": "alice", "email": "alice@example.com"}', 1, '192.168.1.100'),
('users', 2, 'UPDATE', '{"balance": 1500.00}', '{"balance": 2350.75}', 2, '192.168.1.101'),
('products', 1, 'INSERT', NULL, '{"name": "Laptop Pro 15", "price": 1299.99}', 1, '192.168.1.100'),
('orders', 1, 'UPDATE', '{"status": "pending"}', '{"status": "confirmed"}', 1, '192.168.1.102'),
('users', 5, 'UPDATE', '{"is_active": true}', '{"is_active": false}', 4, '192.168.1.105');

-- Create a view for testing
CREATE OR REPLACE VIEW active_users AS
SELECT id, username, email, full_name, balance
FROM users
WHERE is_active = TRUE;

-- Create a stored procedure for testing
DELIMITER //
CREATE PROCEDURE get_user_orders(IN p_user_id BIGINT)
BEGIN
    SELECT o.id, o.order_number, o.total_amount, o.status, o.created_at
    FROM orders o
    WHERE o.user_id = p_user_id
    ORDER BY o.created_at DESC;
END //
DELIMITER ;
