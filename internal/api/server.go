package api

import (
   "context"
   "encoding/json"
   "fmt"
   "log"
   "net/http"
   "time"

   "github.com/gorilla/mux"
   
   "github.com/asterisk-call-routing/internal/config"
   "github.com/asterisk-call-routing/internal/db"
   "github.com/asterisk-call-routing/internal/monitor"
   "github.com/asterisk-call-routing/internal/router"
)

type Server struct {
   config   *config.Config
   router   *router.Router
   monitor  *monitor.Monitor
   server   *http.Server
}

func NewServer(cfg *config.Config, r *router.Router, m *monitor.Monitor) *Server {
   return &Server{
       config:  cfg,
       router:  r,
       monitor: m,
   }
}

func (s *Server) Start(port int) {
   r := mux.NewRouter()
   
   // AGI endpoints for Asterisk
   r.HandleFunc("/process-incoming", s.handleProcessIncoming).Methods("POST")
   r.HandleFunc("/process-return", s.handleProcessReturn).Methods("POST")
   r.HandleFunc("/call-ended", s.handleCallEnded).Methods("POST")
   
   // API endpoints
   r.HandleFunc("/api/stats", s.handleGetStats).Methods("GET")
   r.HandleFunc("/api/calls", s.handleGetCalls).Methods("GET")
   r.HandleFunc("/api/dids", s.handleGetDIDs).Methods("GET")
   r.HandleFunc("/api/health", s.handleHealthCheck).Methods("GET")
   
   // Monitoring endpoints
   r.HandleFunc("/stats", s.handleStats).Methods("GET")
   r.HandleFunc("/health", s.handleHealth).Methods("GET")
   
   // Web interface
   r.PathPrefix("/").Handler(http.FileServer(http.Dir("./web/")))
   
   s.server = &http.Server{
       Addr:         fmt.Sprintf(":%d", port),
       Handler:      r,
       ReadTimeout:  15 * time.Second,
       WriteTimeout: 15 * time.Second,
   }
   
   log.Printf("API server listening on port %d", port)
   if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
       log.Fatalf("API server error: %v", err)
   }
}

func (s *Server) Stop() {
   ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
   defer cancel()
   
   if err := s.server.Shutdown(ctx); err != nil {
       log.Printf("API server shutdown error: %v", err)
   }
}

// AGI Handlers

func (s *Server) handleProcessIncoming(w http.ResponseWriter, r *http.Request) {
   var req struct {
       UniqueID  string `json:"uniqueid"`
       CallerID  string `json:"callerid"`
       Extension string `json:"extension"`
   }
   
   // Parse form data (AGI sends form data, not JSON)
   if err := r.ParseForm(); err != nil {
       http.Error(w, "Invalid request", http.StatusBadRequest)
       return
   }
   
   req.UniqueID = r.FormValue("uniqueid")
   req.CallerID = r.FormValue("callerid")
   req.Extension = r.FormValue("extension")
   
   // Process the call
   resp, err := s.router.ProcessIncomingCall(req.UniqueID, req.CallerID, req.Extension)
   if err != nil {
       log.Printf("Error processing incoming call: %v", err)
       fmt.Fprintf(w, "EXEC Hangup 503\n")
       return
   }
   
   // Return AGI response
   fmt.Fprintf(w, "SET VARIABLE CALL_ID %s\n", resp.CallID)
   fmt.Fprintf(w, "SET VARIABLE DID_ASSIGNED %s\n", resp.DIDAssigned)
   fmt.Fprintf(w, "SET CALLERID \"%s\" <%s>\n", resp.ANIToSend, resp.ANIToSend)
   fmt.Fprintf(w, "EXEC Dial SIP/%s@%s\n", resp.DNISToSend, resp.NextHop)
}

func (s *Server) handleProcessReturn(w http.ResponseWriter, r *http.Request) {
   var req struct {
       UniqueID  string `json:"uniqueid"`
       CallerID  string `json:"callerid"`
       Extension string `json:"extension"`
   }
   
   if err := r.ParseForm(); err != nil {
       http.Error(w, "Invalid request", http.StatusBadRequest)
       return
   }
   
   req.UniqueID = r.FormValue("uniqueid")
   req.CallerID = r.FormValue("callerid")
   req.Extension = r.FormValue("extension")
   
   // Process the return call
   resp, err := s.router.ProcessReturnCall(req.UniqueID, req.CallerID, req.Extension)
   if err != nil {
       log.Printf("Error processing return call: %v", err)
       fmt.Fprintf(w, "EXEC Hangup 404\n")
       return
   }
   
   // Return AGI response
   fmt.Fprintf(w, "SET CALLERID \"%s\" <%s>\n", resp.ANIToSend, resp.ANIToSend)
   fmt.Fprintf(w, "EXEC Dial SIP/%s@%s\n", resp.DNISToSend, resp.NextHop)
}

func (s *Server) handleCallEnded(w http.ResponseWriter, r *http.Request) {
   if err := r.ParseForm(); err != nil {
       http.Error(w, "Invalid request", http.StatusBadRequest)
       return
   }
   
   callID := r.FormValue("uniqueid")
   
   if err := s.router.CompleteCall(callID); err != nil {
       log.Printf("Error completing call %s: %v", callID, err)
   }
   
   fmt.Fprintf(w, "EXEC NoOp Call completed\n")
}

// API Handlers

func (s *Server) handleGetStats(w http.ResponseWriter, r *http.Request) {
   stats, err := db.GetStatistics()
   if err != nil {
       http.Error(w, "Error getting statistics", http.StatusInternalServerError)
       return
   }
   
   // Add router statistics
   response := map[string]interface{}{
       "calls": map[string]interface{}{
           "total":      stats.TotalCalls,
           "completed":  stats.CompletedCalls,
           "failed":     stats.FailedCalls,
           "active":     s.router.GetActiveCallsCount(),
           "success_rate": fmt.Sprintf("%.2f%%", stats.SuccessRate),
           "avg_duration": fmt.Sprintf("%.2f seconds", stats.AvgDuration),
       },
       "dids": map[string]interface{}{
           "total":     stats.TotalDIDs,
           "in_use":    stats.InUseDIDs,
           "available": stats.AvailableDIDs,
           "usage_percent": fmt.Sprintf("%.2f%%", float64(stats.InUseDIDs)/float64(stats.TotalDIDs)*100),
       },
       "system": s.monitor.GetSystemStats(),
       "timestamp": time.Now().Format(time.RFC3339),
   }
   
   w.Header().Set("Content-Type", "application/json")
   json.NewEncoder(w).Encode(response)
}

func (s *Server) handleGetCalls(w http.ResponseWriter, r *http.Request) {
   limit := 100
   calls, err := db.GetRecentCalls(limit)
   if err != nil {
       http.Error(w, "Error getting calls", http.StatusInternalServerError)
       return
   }
   
   w.Header().Set("Content-Type", "application/json")
   json.NewEncoder(w).Encode(calls)
}

func (s *Server) handleGetDIDs(w http.ResponseWriter, r *http.Request) {
   var dids []struct {
       DID      string    `db:"did" json:"did"`
       InUse    bool      `db:"in_use" json:"in_use"`
       Country  string    `db:"country" json:"country"`
       LastUsed time.Time `db:"last_used" json:"last_used"`
   }
   
   err := db.DB.Select(&dids, "SELECT did, in_use, country, last_used FROM dids ORDER BY did")
   if err != nil {
       http.Error(w, "Error getting DIDs", http.StatusInternalServerError)
       return
   }
   
   w.Header().Set("Content-Type", "application/json")
   json.NewEncoder(w).Encode(dids)
}

func (s *Server) handleHealthCheck(w http.ResponseWriter, r *http.Request) {
   // Check database connection
   if err := db.DB.Ping(); err != nil {
       http.Error(w, "Database connection error", http.StatusServiceUnavailable)
       return
   }
   
   // Check DID availability
   var availableDIDs int
   db.DB.Get(&availableDIDs, "SELECT COUNT(*) FROM dids WHERE in_use = 0")
   
   if availableDIDs == 0 {
       http.Error(w, "No DIDs available", http.StatusServiceUnavailable)
       return
   }
   
   response := map[string]interface{}{
       "status": "healthy",
       "checks": map[string]string{
           "database": "ok",
           "dids":     "ok",
           "router":   "ok",
       },
       "timestamp": time.Now().Format(time.RFC3339),
   }
   
   w.Header().Set("Content-Type", "application/json")
   json.NewEncoder(w).Encode(response)
}

// Simple monitoring endpoints
func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
   stats, _ := db.GetStatistics()
   
   response := map[string]interface{}{
       "incoming_calls":   stats.TotalCalls,
       "return_calls":     stats.CompletedCalls,
       "completed_calls":  stats.CompletedCalls,
       "rejected_calls":   0,
       "failed_calls":     stats.FailedCalls,
       "active_calls":     s.router.GetActiveCallsCount(),
       "total_dids":       stats.TotalDIDs,
       "in_use_dids":      stats.InUseDIDs,
       "avg_processing_ms": 50,
       "uptime":           s.monitor.GetUptime(),
   }
   
   w.Header().Set("Content-Type", "application/json")
   json.NewEncoder(w).Encode(response)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
   if err := db.DB.Ping(); err != nil {
       http.Error(w, "Database error", http.StatusServiceUnavailable)
       return
   }
   
   fmt.Fprintf(w, "OK")
}
