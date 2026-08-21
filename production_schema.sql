-- ============================================================================
-- PRODUCTION MYSQL DATABASE SCHEMA & SEED DATA
-- Stationery Management System (Enterprise Logistics & Requisition)
-- Database Engine: MySQL 8.0+ / MariaDB 10.5+ (InnoDB, utf8mb4)
-- ============================================================================

CREATE DATABASE IF NOT EXISTS stationery_db
    DEFAULT CHARACTER SET utf8mb4
    DEFAULT COLLATE utf8mb4_unicode_ci;

USE stationery_db;

-- ----------------------------------------------------------------------------
-- 1. ROLES TABLE
-- ----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS roles (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(50) NOT NULL UNIQUE,
    description VARCHAR(255)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ----------------------------------------------------------------------------
-- 2. BRANCHES TABLE
-- ----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS branches (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    code VARCHAR(20) NOT NULL UNIQUE,
    address TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'ACTIVE',
    deleted_at TIMESTAMP NULL DEFAULT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_branches_status (status),
    INDEX idx_branches_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ----------------------------------------------------------------------------
-- 3. USERS TABLE
-- ----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS users (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    email VARCHAR(100) NOT NULL UNIQUE,
    mobile VARCHAR(20) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL,
    role_id INT NOT NULL,
    branch_id INT NULL,
    department VARCHAR(50) DEFAULT 'General',
    approver_access_type VARCHAR(20) DEFAULT 'ALL_BRANCHES',
    status VARCHAR(20) NOT NULL DEFAULT 'ACTIVE',
    first_login BOOLEAN NOT NULL DEFAULT TRUE,
    deleted_at TIMESTAMP NULL DEFAULT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE RESTRICT,
    FOREIGN KEY (branch_id) REFERENCES branches(id) ON DELETE SET NULL,
    INDEX idx_users_role_id (role_id),
    INDEX idx_users_branch_id (branch_id),
    INDEX idx_users_email (email),
    INDEX idx_users_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ----------------------------------------------------------------------------
-- 4. PRODUCTS TABLE
-- ----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS products (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(150) NOT NULL,
    category VARCHAR(50) NOT NULL,
    unit VARCHAR(30) NOT NULL,
    unit_price DECIMAL(10,2) NOT NULL DEFAULT 0.00,
    description TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'ACTIVE',
    deleted_at TIMESTAMP NULL DEFAULT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_products_category (category),
    INDEX idx_products_status (status),
    INDEX idx_products_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ----------------------------------------------------------------------------
-- 5. REQUESTS TABLE
-- ----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS requests (
    id INT AUTO_INCREMENT PRIMARY KEY,
    request_no VARCHAR(50) NOT NULL UNIQUE,
    branch_id INT NOT NULL,
    requester_id INT NOT NULL,
    applicant_name VARCHAR(100),
    applicant_mobile VARCHAR(20),
    applicant_email VARCHAR(100),
    department VARCHAR(50) NOT NULL DEFAULT 'General',
    location VARCHAR(255),
    priority VARCHAR(20) NOT NULL DEFAULT 'MEDIUM',
    status VARCHAR(30) NOT NULL DEFAULT 'SUBMITTED',
    notes TEXT,
    rejection_reason TEXT,
    courier_name VARCHAR(100),
    tracking_number VARCHAR(100),
    submitted_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    approved_at TIMESTAMP NULL DEFAULT NULL,
    delivered_at TIMESTAMP NULL DEFAULT NULL,
    completed_at TIMESTAMP NULL DEFAULT NULL,
    deleted_at TIMESTAMP NULL DEFAULT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (branch_id) REFERENCES branches(id) ON DELETE RESTRICT,
    FOREIGN KEY (requester_id) REFERENCES users(id) ON DELETE RESTRICT,
    INDEX idx_requests_status (status),
    INDEX idx_requests_branch_id (branch_id),
    INDEX idx_requests_requester_id (requester_id),
    INDEX idx_requests_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ----------------------------------------------------------------------------
-- 6. REQUEST_ITEMS TABLE
-- ----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS request_items (
    id INT AUTO_INCREMENT PRIMARY KEY,
    request_id INT NOT NULL,
    product_id INT NOT NULL,
    requested_qty INT NOT NULL DEFAULT 1,
    unit_price DECIMAL(10,2) NOT NULL DEFAULT 0.00,
    FOREIGN KEY (request_id) REFERENCES requests(id) ON DELETE CASCADE,
    FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE RESTRICT,
    INDEX idx_request_items_request_id (request_id),
    INDEX idx_request_items_product_id (product_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ----------------------------------------------------------------------------
-- 7. APPROVAL_ITEMS TABLE
-- ----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS approval_items (
    id INT AUTO_INCREMENT PRIMARY KEY,
    request_item_id INT NOT NULL,
    approved_qty INT NOT NULL DEFAULT 0,
    approved_by INT NOT NULL,
    remarks TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (request_item_id) REFERENCES request_items(id) ON DELETE CASCADE,
    FOREIGN KEY (approved_by) REFERENCES users(id) ON DELETE RESTRICT,
    INDEX idx_approval_items_request_item_id (request_item_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ----------------------------------------------------------------------------
-- 8. DELIVERIES TABLE
-- ----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS deliveries (
    id INT AUTO_INCREMENT PRIMARY KEY,
    request_id INT NOT NULL,
    agency_user INT NOT NULL,
    courier_name VARCHAR(100),
    tracking_number VARCHAR(100),
    bill_url LONGTEXT,
    bill_notes TEXT,
    delivered_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    status VARCHAR(30) NOT NULL DEFAULT 'DELIVERED',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (request_id) REFERENCES requests(id) ON DELETE CASCADE,
    FOREIGN KEY (agency_user) REFERENCES users(id) ON DELETE RESTRICT,
    INDEX idx_deliveries_request_id (request_id),
    INDEX idx_deliveries_agency_user (agency_user)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ----------------------------------------------------------------------------
-- 9. DELIVERY_ITEMS TABLE
-- ----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS delivery_items (
    id INT AUTO_INCREMENT PRIMARY KEY,
    delivery_id INT NOT NULL,
    product_id INT NOT NULL,
    approved_qty INT NOT NULL DEFAULT 0,
    delivered_qty INT NOT NULL DEFAULT 0,
    unavailable_qty INT NOT NULL DEFAULT 0,
    unit_price DECIMAL(10,2) NOT NULL DEFAULT 0.00,
    remarks TEXT,
    FOREIGN KEY (delivery_id) REFERENCES deliveries(id) ON DELETE CASCADE,
    FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE RESTRICT,
    INDEX idx_delivery_items_delivery_id (delivery_id),
    INDEX idx_delivery_items_product_id (product_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ----------------------------------------------------------------------------
-- 10. VERIFICATION_ITEMS TABLE
-- ----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS verification_items (
    id INT AUTO_INCREMENT PRIMARY KEY,
    delivery_item_id INT NOT NULL,
    accepted_qty INT NOT NULL DEFAULT 0,
    damaged_qty INT NOT NULL DEFAULT 0,
    not_received_qty INT NOT NULL DEFAULT 0,
    remarks TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (delivery_item_id) REFERENCES delivery_items(id) ON DELETE CASCADE,
    INDEX idx_verification_items_delivery_item_id (delivery_item_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ----------------------------------------------------------------------------
-- 11. CHAT_MESSAGES TABLE
-- ----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS chat_messages (
    id INT AUTO_INCREMENT PRIMARY KEY,
    request_id INT NOT NULL,
    sender_id INT NOT NULL,
    sender_name VARCHAR(100) NOT NULL,
    sender_role VARCHAR(50) NOT NULL,
    target_role VARCHAR(50) NOT NULL,
    message TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (request_id) REFERENCES requests(id) ON DELETE CASCADE,
    INDEX idx_chat_messages_request_id (request_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ----------------------------------------------------------------------------
-- 12. SLA_SETTINGS TABLE
-- ----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS sla_settings (
    id INT AUTO_INCREMENT PRIMARY KEY,
    max_approve_days INT NOT NULL DEFAULT 2,
    max_delivery_days INT NOT NULL DEFAULT 3,
    max_verify_days INT NOT NULL DEFAULT 2,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ----------------------------------------------------------------------------
-- 13. NOTIFICATIONS TABLE
-- ----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS notifications (
    id INT AUTO_INCREMENT PRIMARY KEY,
    user_id INT NOT NULL,
    type VARCHAR(50) NOT NULL,
    subject VARCHAR(150) NOT NULL,
    message TEXT NOT NULL,
    is_read BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    INDEX idx_notifications_user_id (user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ----------------------------------------------------------------------------
-- 14. AUDIT_LOGS TABLE
-- ----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS audit_logs (
    id INT AUTO_INCREMENT PRIMARY KEY,
    user_id INT NULL,
    user_name VARCHAR(100),
    action VARCHAR(100) NOT NULL,
    entity_type VARCHAR(50),
    entity_id VARCHAR(50),
    details TEXT,
    ip_address VARCHAR(45),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_audit_logs_user_id (user_id),
    INDEX idx_audit_logs_action (action)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ----------------------------------------------------------------------------
-- 15. EMAIL_QUEUE TABLE
-- ----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS email_queue (
    id INT AUTO_INCREMENT PRIMARY KEY,
    recipient VARCHAR(100) NOT NULL,
    subject VARCHAR(150) NOT NULL,
    body TEXT NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'PENDING',
    attempts INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_email_queue_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;


-- ============================================================================
-- INITIAL SEED DATA
-- Default System Roles, Admin Users, Sample Branches, and Products
-- ============================================================================

-- Insert Roles
INSERT INTO roles (id, name, description) VALUES 
(1, 'ADMIN', 'System Administrator with full management access'),
(2, 'BRANCH_REQUESTER', 'Branch staff who creates stationery requisition requests'),
(3, 'APPROVER', 'Regional / Executive Manager who approves requests'),
(4, 'AGENCY', 'Logistics & delivery agency partner who dispatches & delivers stationery'),
(5, 'MONITOR', 'Read-only monitoring user for SLA tracking and metrics')
ON DUPLICATE KEY UPDATE name=VALUES(name), description=VALUES(description);

-- Insert Branches
INSERT INTO branches (id, name, code, address, status) VALUES 
(1, 'Headquarters', 'HQ-001', '100 Enterprise Tower, Financial District', 'ACTIVE'),
(2, 'North Region Branch', 'BR-101', '45 North Commercial Boulevard', 'ACTIVE'),
(3, 'South Region Branch', 'BR-102', '88 South Industrial Park', 'ACTIVE'),
(4, 'East Coast Office', 'BR-103', '12 Harbour View Road', 'ACTIVE')
ON DUPLICATE KEY UPDATE name=VALUES(name), address=VALUES(address);

-- Insert Default System Users
-- Password for all default accounts: Admin@123
-- Password hash (Bcrypt cost 10): $2a$10$04tNmsqR9Xw3fJ9J8eR4E.hC0lK8h21/WkZ56bH2.R3P4N5M6Q7R8
INSERT INTO users (id, name, email, mobile, password, role_id, branch_id, department, approver_access_type, status, first_login) VALUES
(1, 'System Administrator', 'admin@stationery.com', '09999999999', '$2a$10$04tNmsqR9Xw3fJ9J8eR4E.hC0lK8h21/WkZ56bH2.R3P4N5M6Q7R8', 1, 1, 'General', 'ALL_BRANCHES', 'ACTIVE', FALSE),
(2, 'John Requester', 'requester@stationery.com', '09888888888', '$2a$10$04tNmsqR9Xw3fJ9J8eR4E.hC0lK8h21/WkZ56bH2.R3P4N5M6Q7R8', 2, 2, 'GOLD LOAN', 'ALL_BRANCHES', 'ACTIVE', TRUE),
(3, 'Sarah Approver', 'approver@stationery.com', '09777777777', '$2a$10$04tNmsqR9Xw3fJ9J8eR4E.hC0lK8h21/WkZ56bH2.R3P4N5M6Q7R8', 3, 2, 'CHIT FUND', 'SINGLE_BRANCH', 'ACTIVE', TRUE),
(4, 'Express Delivery Agency', 'agency@stationery.com', '09666666666', '$2a$10$04tNmsqR9Xw3fJ9J8eR4E.hC0lK8h21/WkZ56bH2.R3P4N5M6Q7R8', 4, NULL, 'Logistics', 'ALL_BRANCHES', 'ACTIVE', TRUE),
(5, 'Michael Monitor', 'monitor@stationery.com', '09555555555', '$2a$10$04tNmsqR9Xw3fJ9J8eR4E.hC0lK8h21/WkZ56bH2.R3P4N5M6Q7R8', 5, NULL, 'Audit', 'ALL_BRANCHES', 'ACTIVE', TRUE)
ON DUPLICATE KEY UPDATE name=VALUES(name), email=VALUES(email);

-- Insert Default Products
INSERT INTO products (id, name, category, unit, unit_price, description, status) VALUES 
(1, 'Ballpoint Pen - Blue (Box of 10)', 'Writing Instruments', 'Box', 12.50, 'High-quality blue ink ballpoint pens 0.7mm', 'ACTIVE'),
(2, 'A4 Printing Paper (80gsm - 500 Sheets)', 'Paper Products', 'Ream', 8.99, 'Premium white multipurpose copy paper', 'ACTIVE'),
(3, 'Permanent Marker - Black', 'Writing Instruments', 'Piece', 2.50, 'Chisel tip waterproof black permanent marker', 'ACTIVE'),
(4, 'Heavy Duty Stapler No. 10', 'Desk Supplies', 'Piece', 15.00, 'Durable metal body desk stapler', 'ACTIVE'),
(5, 'Sticky Notes 3x3 Yellow (100 Sheets)', 'Paper Products', 'Pad', 1.99, 'Standard self-adhesive memo pads', 'ACTIVE'),
(6, 'Expandable File Folder A4', 'Filing & Storage', 'Piece', 6.75, 'Heavy-duty poly expandable document organizer', 'ACTIVE'),
(7, '12-Digit Desk Calculator', 'Electronics', 'Piece', 22.00, 'Dual-power solar and battery desktop calculator', 'ACTIVE'),
(8, 'Paper Clips (100 pcs/box)', 'Desk Supplies', 'Box', 3.25, 'Vinyl coated rust-resistant paper clips', 'ACTIVE')
ON DUPLICATE KEY UPDATE name=VALUES(name), unit_price=VALUES(unit_price);

-- Insert Initial SLA Configuration
INSERT INTO sla_settings (id, max_approve_days, max_delivery_days, max_verify_days) VALUES 
(1, 2, 3, 2)
ON DUPLICATE KEY UPDATE max_approve_days=VALUES(max_approve_days);
