package db

import (
    "database/sql"
    "fmt"
    "log"
    "time"
    
    _ "github.com/go-sql-driver/mysql"
    "github.com/asterisk-call-routing/internal/models"
)

type Database struct {
    conn *sql.DB
}

func NewDatabase(dsn string) (*Database, error) {
    conn, err := sql.Open("mysql", dsn)
    if err != nil {
        return nil, err
    }
    
    if err := conn.Ping(); err != nil {
        return nil, err
    }
    
    // Set connection pool settings
    conn.SetMaxOpenConns(25)
    conn.SetMaxIdleConns(5)
    conn.SetConnMaxLifetime(5 * time.Minute)
    
    return &Database{conn: conn}, nil
}

func (db *Database) GetAvailableDID() (*models.DID, error) {
    query := `
        SELECT id, did, country 
        FROM dids 
        WHERE in_use = 0 
        ORDER BY RAND() 
        LIMIT 1
        FOR UPDATE
    `
    
    did := &models.DID{}
    err := db.conn.QueryRow(query).Scan(&did.ID, &did.DID, &did.Country)
    if err != nil {
        return nil, fmt.Errorf("no available DIDs: %v", err)
    }
    
    return did, nil
}

func (db *Database) MarkDIDInUse(did string, destination string) error {
    query := `
        UPDATE dids 
        SET in_use = 1, destination = ?, updated_at = NOW()
        WHERE did = ?
    `
    
    result, err := db.conn.Exec(query, destination, did)
    if err != nil {
        return err
    }
    
    rows, _ := result.RowsAffected()
    if rows == 0 {
        return fmt.Errorf("DID not found: %s", did)
    }
    
    return nil
}

func (db *Database) ReleaseDID(did string) error {
    query := `
        UPDATE dids 
        SET in_use = 0, destination = NULL, updated_at = NOW()
        WHERE did = ?
    `
    
    _, err := db.conn.Exec(query, did)
    return err
}

func (db *Database) StoreCallRecord(record *models.CallRecord) error {
    query := `
        INSERT INTO call_records 
        (call_id, original_ani, original_dnis, assigned_did, status, timestamp)
        VALUES (?, ?, ?, ?, ?, ?)
    `
    
    _, err := db.conn.Exec(query, 
        record.CallID, 
        record.OriginalANI, 
        record.OriginalDNIS,
        record.AssignedDID, 
        record.Status, 
        record.Timestamp,
    )
    
    return err
}

func (db *Database) UpdateCallStatus(callID string, status string, duration int) error {
    query := `
        UPDATE call_records 
        SET status = ?, duration = ?
        WHERE call_id = ?
    `
    
    _, err := db.conn.Exec(query, status, duration, callID)
    return err
}

func (db *Database) GetCallRecordByDID(did string) (*models.CallRecord, error) {
    query := `
        SELECT call_id, original_ani, original_dnis, assigned_did, status, timestamp
        FROM call_records
        WHERE assigned_did = ? AND status IN ('ACTIVE', 'FORWARDED_TO_S3')
        ORDER BY timestamp DESC
        LIMIT 1
    `
    
    record := &models.CallRecord{}
    err := db.conn.QueryRow(query, did).Scan(
        &record.CallID,
        &record.OriginalANI,
        &record.OriginalDNIS,
        &record.AssignedDID,
        &record.Status,
        &record.Timestamp,
    )
    
    return record, err
}

func (db *Database) GetStatistics() (map[string]interface{}, error) {
    stats := make(map[string]interface{})
    
    // Get DID statistics
    var totalDIDs, usedDIDs int
    db.conn.QueryRow("SELECT COUNT(*), SUM(CASE WHEN in_use = 1 THEN 1 ELSE 0 END) FROM dids").Scan(&totalDIDs, &usedDIDs)
    
    stats["total_dids"] = totalDIDs
    stats["used_dids"] = usedDIDs
    stats["available_dids"] = totalDIDs - usedDIDs
    
    // Get call statistics
    var todaysCalls, activeCalls int
    db.conn.QueryRow("SELECT COUNT(*) FROM call_records WHERE DATE(timestamp) = CURDATE()").Scan(&todaysCalls)
    db.conn.QueryRow("SELECT COUNT(*) FROM call_records WHERE status = 'ACTIVE'").Scan(&activeCalls)
    
    stats["calls_today"] = todaysCalls
    stats["active_calls"] = activeCalls
    
    return stats, nil
}

func (db *Database) Close() error {
    return db.conn.Close()
}
