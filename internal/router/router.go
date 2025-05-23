func (r *Router) performCleanup() {
   // Clean up stuck DIDs
   count, err := db.CleanupStuckDIDs(60) // 60 minutes threshold
   if err != nil {
       log.Printf("Error cleaning up stuck DIDs: %v", err)
   } else if count > 0 {
       log.Printf("Cleaned up %d stuck DIDs", count)
   }
   
   // Clean up orphaned calls
   r.activeCallsMu.Lock()
   for callID, callState := range r.activeCalls {
       if time.Since(callState.StartTime) > time.Duration(r.config.Router.CallTimeout)*time.Second {
           delete(r.activeCalls, callID)
           
           // Update database
           db.UpdateCallRecord(callID, "timeout", time.Now(), int(time.Since(callState.StartTime).Seconds()))
           
           // Release DID if still in use
           if callState.DIDUsed != "" {
               db.MarkDIDAvailable(callState.DIDUsed)
           }
           
           log.Printf("Cleaned up timed out call: %s", callID)
       }
   }
   r.activeCallsMu.Unlock()
}

// GetActiveCallsCount returns the number of active calls
func (r *Router) GetActiveCallsCount() int {
   r.activeCallsMu.RLock()
   defer r.activeCallsMu.RUnlock()
   return len(r.activeCalls)
}

// GetCallState returns the state of a specific call
func (r *Router) GetCallState(callID string) (*CallState, bool) {
   r.activeCallsMu.RLock()
   defer r.activeCallsMu.RUnlock()
   
   state, exists := r.activeCalls[callID]
   return state, exists
}

type CallResponse struct {
   Success     bool   `json:"success"`
   CallID      string `json:"call_id"`
   DIDAssigned string `json:"did_assigned,omitempty"`
   NextHop     string `json:"next_hop"`
   ANIToSend   string `json:"ani_to_send"`
   DNISToSend  string `json:"dnis_to_send"`
   Error       string `json:"error,omitempty"`
}
func (r *Router) performCleanup() {
   // Clean up stuck DIDs
   count, err := db.CleanupStuckDIDs(60) // 60 minutes threshold
   if err != nil {
       log.Printf("Error cleaning up stuck DIDs: %v", err)
   } else if count > 0 {
       log.Printf("Cleaned up %d stuck DIDs", count)
   }
   
   // Clean up orphaned calls
   r.activeCallsMu.Lock()
   for callID, callState := range r.activeCalls {
       if time.Since(callState.StartTime) > time.Duration(r.config.Router.CallTimeout)*time.Second {
           delete(r.activeCalls, callID)
           
           // Update database
           db.UpdateCallRecord(callID, "timeout", time.Now(), int(time.Since(callState.StartTime).Seconds()))
           
           // Release DID if still in use
           if callState.DIDUsed != "" {
               db.MarkDIDAvailable(callState.DIDUsed)
           }
           
           log.Printf("Cleaned up timed out call: %s", callID)
       }
   }
   r.activeCallsMu.Unlock()
}

// GetActiveCallsCount returns the number of active calls
func (r *Router) GetActiveCallsCount() int {
   r.activeCallsMu.RLock()
   defer r.activeCallsMu.RUnlock()
   return len(r.activeCalls)
}

// GetCallState returns the state of a specific call
func (r *Router) GetCallState(callID string) (*CallState, bool) {
   r.activeCallsMu.RLock()
   defer r.activeCallsMu.RUnlock()
   
   state, exists := r.activeCalls[callID]
   return state, exists
}

type CallResponse struct {
   Success     bool   `json:"success"`
   CallID      string `json:"call_id"`
   DIDAssigned string `json:"did_assigned,omitempty"`
   NextHop     string `json:"next_hop"`
   ANIToSend   string `json:"ani_to_send"`
   DNISToSend  string `json:"dnis_to_send"`
   Error       string `json:"error,omitempty"`
}
