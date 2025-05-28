-- Create database
CREATE DATABASE IF NOT EXISTS call_routing;
USE call_routing;

-- Create DIDs table
CREATE TABLE IF NOT EXISTS dids (
    id INT AUTO_INCREMENT PRIMARY KEY,
    did VARCHAR(20) NOT NULL UNIQUE,
    in_use BOOLEAN DEFAULT FALSE,
    country VARCHAR(50),
    destination VARCHAR(20),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_in_use (in_use),
    INDEX idx_did (did)
);

-- Create call_records table
CREATE TABLE IF NOT EXISTS call_records (
    id INT AUTO_INCREMENT PRIMARY KEY,
    call_id VARCHAR(50) NOT NULL,
    original_ani VARCHAR(20) NOT NULL,
    original_dnis VARCHAR(20) NOT NULL,
    assigned_did VARCHAR(20),
    status VARCHAR(50),
    duration INT DEFAULT 0,
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_call_id (call_id),
    INDEX idx_did (assigned_did),
    INDEX idx_timestamp (timestamp),
    INDEX idx_status (status)
);

-- Insert sample DIDs
INSERT IGNORE INTO dids (did, country) VALUES
    ('584148757547', 'Venezuela'),
    ('584249726299', 'Venezuela'),
    ('584167000000', 'Venezuela'),
    ('584267000011', 'Venezuela'),
    ('584148757548', 'Venezuela'),
    ('584249726300', 'Venezuela'),
    ('584167000001', 'Venezuela'),
    ('584267000012', 'Venezuela'),
    ('584148757549', 'Venezuela'),
    ('584249726301', 'Venezuela');
