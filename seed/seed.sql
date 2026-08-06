-- Seed Data for Stationery Management System

INSERT INTO roles (id, name, description) VALUES 
(1, 'ADMIN', 'System Administrator with full access'),
(2, 'BRANCH_REQUESTER', 'Branch user who creates stationery requests'),
(3, 'APPROVER', 'Manager who approves stationery requests'),
(4, 'AGENCY', 'Delivery agency user who delivers approved stationery'),
(5, 'MONITOR', 'Read-only monitoring user who can trigger reminder emails');

INSERT INTO branches (id, name, code, address, status) VALUES 
(1, 'Headquarters', 'HQ-001', '100 Enterprise Tower, Financial District', 'ACTIVE'),
(2, 'North Region Branch', 'BR-101', '45 North Commercial Boulevard', 'ACTIVE'),
(3, 'South Region Branch', 'BR-102', '88 South Industrial Park', 'ACTIVE'),
(4, 'East Coast Office', 'BR-103', '12 Harbour View Road', 'ACTIVE');

-- Default Password for all initial users: Admin@123
-- Hash below generated with Bcrypt cost 10: $2a$10$hK.fN.w1n65mC90D14g2x.l0p5z.2G8W8T9eK.g7Y0f56z8rW
INSERT INTO users (id, name, email, mobile, password, role_id, branch_id, approver_access_type, status, first_login) VALUES
(1, 'System Administrator', 'admin@stationery.com', '09999999999', '$2a$10$04tNmsqR9Xw3fJ9J8eR4E.hC0lK8h21/WkZ56bH2.R3P4N5M6Q7R8', 1, 1, 'ALL_BRANCHES', 'ACTIVE', FALSE),
(2, 'John Requester', 'requester@stationery.com', '09888888888', '$2a$10$04tNmsqR9Xw3fJ9J8eR4E.hC0lK8h21/WkZ56bH2.R3P4N5M6Q7R8', 2, 2, 'ALL_BRANCHES', 'ACTIVE', TRUE),
(3, 'Sarah Approver', 'approver@stationery.com', '09777777777', '$2a$10$04tNmsqR9Xw3fJ9J8eR4E.hC0lK8h21/WkZ56bH2.R3P4N5M6Q7R8', 3, 2, 'SINGLE_BRANCH', 'ACTIVE', TRUE),
(4, 'Express Delivery Agency', 'agency@stationery.com', '09666666666', '$2a$10$04tNmsqR9Xw3fJ9J8eR4E.hC0lK8h21/WkZ56bH2.R3P4N5M6Q7R8', 4, NULL, 'ALL_BRANCHES', 'ACTIVE', TRUE),
(5, 'Michael Monitor', 'monitor@stationery.com', '09555555555', '$2a$10$04tNmsqR9Xw3fJ9J8eR4E.hC0lK8h21/WkZ56bH2.R3P4N5M6Q7R8', 5, NULL, 'ALL_BRANCHES', 'ACTIVE', TRUE);

INSERT INTO products (id, name, category, unit, description, status) VALUES 
(1, 'Ballpoint Pen - Blue (Box of 10)', 'Writing Instruments', 'Box', 'High-quality blue ink ballpoint pens 0.7mm', 'ACTIVE'),
(2, 'A4 Printing Paper (80gsm - 500 Sheets)', 'Paper Products', 'Ream', 'Premium white multipurpose copy paper', 'ACTIVE'),
(3, 'Permanent Marker - Black', 'Writing Instruments', 'Piece', 'Chisel tip waterproof black permanent marker', 'ACTIVE'),
(4, 'Heavy Duty Stapler No. 10', 'Desk Supplies', 'Piece', 'Durable metal body desk stapler', 'ACTIVE'),
(5, 'Sticky Notes 3x3 Yellow (100 Sheets)', 'Paper Products', 'Pad', 'Standard self-adhesive memo pads', 'ACTIVE'),
(6, 'Expandable File Folder A4', 'Filing & Storage', 'Piece', 'Heavy-duty poly expandable document organizer', 'ACTIVE'),
(7, '12-Digit Desk Calculator', 'Electronics', 'Piece', 'Dual-power solar and battery desktop calculator', 'ACTIVE'),
(8, 'Paper Clips (100 pcs/box)', 'Desk Supplies', 'Box', 'Vinyl coated rust-resistant paper clips', 'ACTIVE');
