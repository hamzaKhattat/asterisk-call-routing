package router

import (
    "fmt"
    "log"
    "sync"
    "time"
    
    "github.com/asterisk-call-routing/internal/db"
    "github.com/asterisk-call-routing/internal/models"
)

type Router struct {
    db              *db.Database
    mu              sync.RWMutex
    activeCallsMap  map[string]*models.CallRecord
    didToCallMap    map[string]string
}

func NewRouter(database *db.Database) *Router {
    return &Router{
        db:             database,
        activeCallsMap: make(map[string]*models.CallRecord),
        didToCallMap:   make(map[string]string),
    }
}

// ProcessIncomingCall handles initial calls from S1
func (r *Router) ProcessIncomingCall(callID, ani, dnis string) (*models.CallResponse, error) {
    r.mu.Lock()
    defer r.mu.Unlock()
    
    log.Printf("[ROUTER] Processing incoming call: CallID=%s, ANI=%s, DNIS=%s", callID, ani, dnis)
    
    // Get available DID
    did, err := r.db.GetAvailableDID()
    if err != nil {
        log.Printf("[ROUTER] Failed to get available DID: %v", err)
        return nil, err
    }
    
    // Mark DID as in use with DNIS as destination
    if err := r.db.MarkDIDInUse(did.DID, dnis); err != nil {
        log.Printf("[ROUTER] Failed to mark DID in use: %v", err)
        return nil, err
    }
    
    // Create call record
    record := &models.CallRecord{
        CallID:       callID,
        OriginalANI:  ani,
        OriginalDNIS: dnis,
        AssignedDID:  did.DID,
        Status:       "ACTIVE",
        Timestamp:    time.Now(),
    }
    
    // Store in memory maps
    r.activeCallsMap[callID] = record
    r.didToCallMap[did.DID] = callID
    
    // Store in database
    if err := r.db.StoreCallRecord(record); err != nil {
        log.Printf("[ROUTER] Failed to store call record: %v", err)
    }
    
    // According to workflow: ANI-2 = DNIS-1, DID is the new destination
    response := &models.CallResponse{
        Status:      "success",
        DIDAssigned: did.DID,
        NextHop:     "trunk-s3",
        ANIToSend:   dnis,      // DNIS-1 becomes ANI-2
        DNISToSend:  did.DID,   // DID becomes destination
    }
    
    log.Printf("[ROUTER] Call routed: ANI-2=%s (was DNIS-1), DID=%s", dnis, did.DID)
    
    return response, nil
}

// ProcessReturnCall handles calls returning from S3
func (r *Router) ProcessReturnCall(ani2, did string) (*models.CallResponse, error) {
    r.mu.Lock()
    defer r.mu.Unlock()
    
    log.Printf("[ROUTER] Processing return call: ANI-2=%s, DID=%s", ani2, did)
    
    // Get call ID from DID mapping
    callID, exists := r.didToCallMap[did]
    if !exists {
        // Try to get from database
        record, err := r.db.GetCallRecordByDID(did)
        if err != nil {
            log.Printf("[ROUTER] No record found for DID %s: %v", did, err)
            return nil, fmt.Errorf("no active call for DID %s", did)
        }
        callID = record.CallID
    }
    
    // Get call record
    record, exists := r.activeCallsMap[callID]
    if !exists {
        record, err := r.db.GetCallRecordByDID(did)
        if err != nil {
            return nil, fmt.Errorf("call record not found")
        }
    }
    
    // Verify ANI-2 matches original DNIS-1
    if ani2 != record.OriginalDNIS {
        log.Printf("[ROUTER] ANI mismatch: expected %s, got %s", record.OriginalDNIS, ani2)
        return nil, fmt.Errorf("ANI verification failed")
    }
    
    // Release DID
    if err := r.db.ReleaseDID(did); err != nil {
        log.Printf("[ROUTER] Failed to release DID: %v", err)
    }
    
    // Update call status
    r.db.UpdateCallStatus(callID, "COMPLETED", 0)
    
    // Clean up memory maps
    delete(r.activeCallsMap, callID)
    delete(r.didToCallMap, did)
    
    // Return original ANI and DNIS for forwarding to S4
    response := &models.CallResponse{
        Status:     "success",
        NextHop:    "trunk-s4",
        ANIToSend:  record.OriginalANI,  // Restore original ANI
        DNISToSend: record.OriginalDNIS, // Restore original DNIS
    }
    
    log.Printf("[ROUTER] Returning original: ANI=%s, DNIS=%s", record.OriginalANI, record.OriginalDNIS)
    
    return response, nil
}

// GetStatistics returns current router statistics
func (r *Router) GetStatistics() (map[string]interface{}, error) {
    r.mu.RLock()
    activeCalls := len(r.activeCallsMap)
    r.mu.RUnlock()
    
    stats, err := r.db.GetStatistics()
    if err != nil {
        return nil, err
    }
    
    stats["memory_active_calls"] = activeCalls
    stats["timestamp"] = time.Now().Format(time.RFC3339)
    
    return stats, nil
}
