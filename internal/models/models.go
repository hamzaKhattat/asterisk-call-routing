package models

import (
    "database/sql/driver"
    "time"
)

// DID represents a Direct Inward Dialing number
type DID struct {
    ID          int       `db:"id"`
    DID         string    `db:"did"`
    InUse       bool      `db:"in_use"`
    Country     string    `db:"country"`
    Destination string    `db:"destination"`
    LastUsed    time.Time `db:"last_used"`
    CreatedAt   time.Time `db:"created_at"`
}

// CallRecord represents a call routing record
type CallRecord struct {
    ID                int       `db:"id"`
    CallID            string    `db:"call_id"`
    ANIOriginal       string    `db:"ani_original"`
    DNISOriginal      string    `db:"dnis_original"`
    ANIModified       string    `db:"ani_modified"`
    DIDUsed           string    `db:"did_used"`
    StartTime         time.Time `db:"start_time"`
    EndTime           NullTime  `db:"end_time"`
    Duration          int       `db:"duration"`
    Status            string    `db:"status"`
    ServerOrigin      string    `db:"server_origin"`
    ServerDestination string    `db:"server_destination"`
    CallPath          string    `db:"call_path"`
    CreatedAt         time.Time `db:"created_at"`
}

// NullTime handles nullable time fields
type NullTime struct {
    Time  time.Time
    Valid bool
}

func (nt *NullTime) Scan(value interface{}) error {
    if value == nil {
        nt.Time, nt.Valid = time.Time{}, false
        return nil
    }
    nt.Valid = true
    return (&nt.Time).Scan(value)
}

func (nt NullTime) Value() (driver.Value, error) {
    if !nt.Valid {
        return nil, nil
    }
    return nt.Time, nil
}

// CallStatistics represents aggregated call statistics
type CallStatistics struct {
    TotalCalls      int     `db:"total_calls"`
    CompletedCalls  int     `db:"completed_calls"`
    FailedCalls     int     `db:"failed_calls"`
    ActiveCalls     int     `db:"active_calls"`
    AvgDuration     float64 `db:"avg_duration"`
    SuccessRate     float64 `db:"success_rate"`
    TotalDIDs       int     `db:"total_dids"`
    InUseDIDs       int     `db:"in_use_dids"`
    AvailableDIDs   int     `db:"available_dids"`
}
