-- Create database if not exists
CREATE DATABASE IF NOT EXISTS call_routing CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE call_routing;

-- DIDs table
CREATE TABLE IF NOT EXISTS dids (
   id INT AUTO_INCREMENT PRIMARY KEY,
   did VARCHAR(20) NOT NULL UNIQUE,
   in_use BOOLEAN NOT NULL DEFAULT FALSE,
   country VARCHAR(50),
   destination VARCHAR(20),
   last_used DATETIME,
   created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
   INDEX idx_in_use (in_use),
   INDEX idx_last_used (last_used)
) ENGINE=InnoDB;

-- Call records table
CREATE TABLE IF NOT EXISTS call_records (
   id INT AUTO_INCREMENT PRIMARY KEY,
   call_id VARCHAR(100) NOT NULL UNIQUE,
   ani_original VARCHAR(20) NOT NULL,
   dnis_original VARCHAR(20) NOT NULL,
   ani_modified VARCHAR(20),
   did_used VARCHAR(20),
   start_time DATETIME NOT NULL,
   end_time DATETIME,
   duration INT,
   status VARCHAR(20) NOT NULL,
   server_origin VARCHAR(50),
   server_destination VARCHAR(50),
   call_path TEXT,
   created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
   INDEX idx_call_id (call_id),
   INDEX idx_did_used (did_used),
   INDEX idx_start_time (start_time),
   INDEX idx_status (status),
   INDEX idx_ani_original (ani_original),
   INDEX idx_dnis_original (dnis_original)
) ENGINE=InnoDB;

-- Call events table for detailed tracking
CREATE TABLE IF NOT EXISTS call_events (
   id INT AUTO_INCREMENT PRIMARY KEY,
   call_id VARCHAR(100) NOT NULL,
   event_type VARCHAR(50) NOT NULL,
   event_data JSON,
   created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
   INDEX idx_call_id (call_id),
   INDEX idx_event_type (event_type),
   INDEX idx_created_at (created_at),
   FOREIGN KEY (call_id) REFERENCES call_records(call_id) ON DELETE CASCADE
) ENGINE=InnoDB;

-- System parameters table
CREATE TABLE IF NOT EXISTS system_parameters (
   id INT AUTO_INCREMENT PRIMARY KEY,
   param_name VARCHAR(50) NOT NULL UNIQUE,
   param_value VARCHAR(255),
   description TEXT,
   updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB;

-- Insert default parameters
INSERT INTO system_parameters (param_name, param_value, description) VALUES
('did_cleanup_threshold', '60', 'Minutes after which stuck DIDs are released'),
('call_timeout', '300', 'Default call timeout in seconds'),
('max_concurrent_calls', '1000', 'Maximum concurrent calls allowed'),
('did_selection_mode', 'random', 'DID selection mode: random or sequential')
ON DUPLICATE KEY UPDATE param_value = VALUES(param_value);
