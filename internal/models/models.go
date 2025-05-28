package models

import (
    "time"
)

type DID struct {
    ID          int       `json:"id"`
    DID         string    `json:"did"`
    InUse       bool      `json:"in_use"`
    Country     string    `json:"country"`
    Destination string    `json:"destination"`
    UpdatedAt   time.Time `json:"updated_at"`
}

type CallRecord struct {
    ID           int       `json:"id"`
    CallID       string    `json:"call_id"`
    OriginalANI  string    `json:"original_ani"`
    OriginalDNIS string    `json:"original_dnis"`
    AssignedDID  string    `json:"assigned_did"`
    Status       string    `json:"status"`
    Duration     int       `json:"duration"`
    Timestamp    time.Time `json:"timestamp"`
}

type CallResponse struct {
    Status      string `json:"status"`
    DIDAssigned string `json:"did_assigned"`
    NextHop     string `json:"next_hop"`
    ANIToSend   string `json:"ani_to_send"`
    DNISToSend  string `json:"dnis_to_send"`
}
