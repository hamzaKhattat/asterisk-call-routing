package db

import (
    "database/sql"
    "encoding/csv"
    "fmt"
    "log"
    "math/rand"
    "os"
    "time"

    _ "github.com/go-sql-driver/mysql"
    "github.com/jmoiron/sqlx"

    "github.com/asterisk-call-routing/internal/config"
    "github.com/asterisk-call-routing/internal/models"
)

var DB *sqlx.DB

func Initialize(cfg *config.Config) error {
    dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true",
        cfg.Database.Username,
        cfg.Database.Password,
        cfg.Database.Host,
        cfg.Database.Port,
        cfg.Database.Name,
    )

    var err error
    DB, err = sqlx.Connect("mysql", dsn)
    if err != nil {
        return err
    }

    // Configure connection pool
    DB.SetMaxOpenConns(25)
    DB.SetMaxIdleConns(5)
    DB.SetConnMaxLifetime(5 * time.Minute)

    // Test connection
    if err := DB.Ping(); err != nil {
        return err
    }

    log.Println("Database connection established")
    return nil
}

func Close() {
    if DB != nil {
        DB.Close()
    }
}

// GetAvailableDID retrieves a random available DID
func GetAvailableDID() (*models.DID, error) {
    tx, err := DB.Beginx()
    if err != nil {
        return nil, err
    }
    defer tx.Rollback()

    // Lock the table to prevent race conditions
    var did models.DID
    query := `
        SELECT * FROM dids 
        WHERE in_use = 0 
        ORDER BY RAND() 
        LIMIT 1 
        FOR UPDATE
    `
    
    err = tx.Get(&did, query)
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, fmt.Errorf("no available DIDs")
        }
        return nil, err
    }

    // Mark as in use
    _, err = tx.Exec("UPDATE dids SET in_use = 1, last_used = NOW() WHERE id = ?", did.ID)
    if err != nil {
        return nil, err
    }

    if err := tx.Commit(); err != nil {
        return nil, err
    }

    return &did, nil
}

// MarkDIDInUse marks a DID as in use and associates it with a destination
func MarkDIDInUse(did string, destination string) error {
    query := `
        UPDATE dids 
        SET in_use = 1, destination = ?, last_used = NOW() 
        WHERE did = ?
    `
    _, err := DB.Exec(query, destination, did)
    return err
}

// MarkDIDAvailable marks a DID as available
func MarkDIDAvailable(did string) error {
    query := `
        UPDATE dids 
        SET in_use = 0, destination = NULL 
        WHERE did = ?
    `
    _, err := DB.Exec(query, did)
    return err
}

// GetDestinationForDID retrieves the original destination for a DID
func GetDestinationForDID(did string) (string, error) {
    var destination sql.NullString
    err := DB.Get(&destination, "SELECT destination FROM dids WHERE did = ?", did)
    if err != nil {
        return "", err
    }
    if !destination.Valid {
        return "", fmt.Errorf("no destination found for DID %s", did)
    }
    return destination.String, nil
}

// CreateCallRecord creates a new call record
func CreateCallRecord(record *models.CallRecord) error {
    query := `
        INSERT INTO call_records (
            call_id, ani_original, dnis_original, ani_modified, 
            did_used, start_time, status, server_origin, 
            server_destination, call_path
        ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
    `
    _, err := DB.Exec(query,
        record.CallID, record.ANIOriginal, record.DNISOriginal,
        record.ANIModified, record.DIDUsed, record.StartTime,
        record.Status, record.ServerOrigin, record.ServerDestination,
        record.CallPath,
    )
    return err
}

// UpdateCallRecord updates an existing call record
func UpdateCallRecord(callID, status string, endTime time.Time, duration int) error {
    query := `
        UPDATE call_records 
        SET status = ?, end_time = ?, duration = ? 
        WHERE call_id = ?
    `
    _, err := DB.Exec(query, status, endTime, duration, callID)
    return err
}

// GetCallRecord retrieves a call record by call ID
func GetCallRecord(callID string) (*models.CallRecord, error) {
    var record models.CallRecord
    err := DB.Get(&record, "SELECT * FROM call_records WHERE call_id = ?", callID)
    if err != nil {
        return nil, err
    }
    return &record, nil
}

// ImportDIDsFromCSV imports DIDs from a CSV file
func ImportDIDsFromCSV(filename string) (int, error) {
    file, err := os.Open(filename)
    if err != nil {
        return 0, err
    }
    defer file.Close()

    reader := csv.NewReader(file)
    records, err := reader.ReadAll()
    if err != nil {
        return 0, err
    }

    count := 0
    for i, record := range records {
        if i == 0 && record[0] == "DID" {
            continue // Skip header
        }

        if len(record) >= 2 {
            did := record[0]
            country := record[1]

            query := `
                INSERT INTO dids (did, country, in_use) 
                VALUES (?, ?, 0) 
                ON DUPLICATE KEY UPDATE country = VALUES(country)
            `
            _, err := DB.Exec(query, did, country)
            if err != nil {
                log.Printf("Error importing DID %s: %v", did, err)
                continue
            }
            count++
        }
    }

    return count, nil
}

// CleanupStuckDIDs releases DIDs that have been stuck in use
func CleanupStuckDIDs(thresholdMinutes int) (int, error) {
    threshold := time.Now().Add(-time.Duration(thresholdMinutes) * time.Minute)
    
    result, err := DB.Exec(`
        UPDATE dids 
        SET in_use = 0, destination = NULL 
        WHERE in_use = 1 AND last_used < ?
    `, threshold)
    
    if err != nil {
        return 0, err
    }
    
    rows, _ := result.RowsAffected()
    return int(rows), nil
}

// GetStatistics retrieves call and DID statistics
func GetStatistics() (*models.CallStatistics, error) {
    stats := &models.CallStatistics{}
    
    // Get call statistics
    err := DB.Get(stats, `
        SELECT 
            COUNT(*) as total_calls,
            SUM(CASE WHEN status = 'completed' THEN 1 ELSE 0 END) as completed_calls,
            SUM(CASE WHEN status IN ('failed', 'timeout') THEN 1 ELSE 0 END) as failed_calls,
            SUM(CASE WHEN status IN ('started', 'in_progress', 'returning') THEN 1 ELSE 0 END) as active_calls,
            COALESCE(AVG(CASE WHEN status = 'completed' THEN duration END), 0) as avg_duration
        FROM call_records
    `)
    if err != nil {
        return nil, err
    }
    
    // Calculate success rate
    if stats.TotalCalls > 0 {
        stats.SuccessRate = float64(stats.CompletedCalls) / float64(stats.TotalCalls) * 100
    }
    
    // Get DID statistics
    err = DB.Get(&stats.TotalDIDs, "SELECT COUNT(*) FROM dids")
    if err != nil {
        return nil, err
    }
    
    err = DB.Get(&stats.InUseDIDs, "SELECT COUNT(*) FROM dids WHERE in_use = 1")
    if err != nil {
        return nil, err
    }
    
    stats.AvailableDIDs = stats.TotalDIDs - stats.InUseDIDs
    
    return stats, nil
}

// GetRecentCalls retrieves recent call records
func GetRecentCalls(limit int) ([]models.CallRecord, error) {
    var calls []models.CallRecord
    query := `
        SELECT * FROM call_records 
        ORDER BY start_time DESC 
        LIMIT ?
    `
    err := DB.Select(&calls, query, limit)
    return calls, err
}
