package api

import (
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "time"
    
    "github.com/asterisk-call-routing/internal/router"
)

type Server struct {
    router *router.Router
    port   int
}

func NewServer(r *router.Router, port int) *Server {
    return &Server{
        router: r,
        port:   port,
    }
}

func (s *Server) Start() error {
    mux := http.NewServeMux()
    
    // API endpoints
    mux.HandleFunc("/api/processIncoming", s.handleProcessIncoming)
    mux.HandleFunc("/api/processReturn", s.handleProcessReturn)
    mux.HandleFunc("/api/stats", s.handleStats)
    mux.HandleFunc("/api/health", s.handleHealth)
    
    // Asterisk AGI endpoints
    mux.HandleFunc("/getdid", s.handleGetDID)
    mux.HandleFunc("/releasedid", s.handleReleaseDID)
    
    server := &http.Server{
        Addr:         fmt.Sprintf(":%d", s.port),
        Handler:      mux,
        ReadTimeout:  10 * time.Second,
        WriteTimeout: 10 * time.Second,
    }
    
    log.Printf("[API] Server starting on port %d", s.port)
    return server.ListenAndServe()
}

func (s *Server) handleProcessIncoming(w http.ResponseWriter, r *http.Request) {
    callID := r.URL.Query().Get("callid")
    ani := r.URL.Query().Get("ani")
    dnis := r.URL.Query().Get("dnis")
    
    if callID == "" || ani == "" || dnis == "" {
        http.Error(w, "Missing parameters", http.StatusBadRequest)
        return
    }
    
    resp, err := s.router.ProcessIncomingCall(callID, ani, dnis)
    if err != nil {
        log.Printf("[API] ProcessIncoming error: %v", err)
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(resp)
}

func (s *Server) handleProcessReturn(w http.ResponseWriter, r *http.Request) {
    ani2 := r.URL.Query().Get("ani2")
    did := r.URL.Query().Get("did")
    
    if ani2 == "" || did == "" {
        http.Error(w, "Missing parameters", http.StatusBadRequest)
        return
    }
    
    resp, err := s.router.ProcessReturnCall(ani2, did)
    if err != nil {
        log.Printf("[API] ProcessReturn error: %v", err)
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(resp)
}

// Legacy endpoint for compatibility
func (s *Server) handleGetDID(w http.ResponseWriter, r *http.Request) {
    callID := r.URL.Query().Get("callid")
    ani := r.URL.Query().Get("ani")
    dnis := r.URL.Query().Get("dnis")
    
    resp, err := s.router.ProcessIncomingCall(callID, ani, dnis)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    // Return in legacy format
    fmt.Fprintf(w, "DID:%s,ANI2:%s", resp.DIDAssigned, resp.ANIToSend)
}

// Legacy endpoint for compatibility
func (s *Server) handleReleaseDID(w http.ResponseWriter, r *http.Request) {
    did := r.URL.Query().Get("did")
    ani2 := r.URL.Query().Get("ani2")
    
    if ani2 == "" {
        // Try to infer from DID
        ani2 = did
    }
    
    resp, err := s.router.ProcessReturnCall(ani2, did)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    // Return in legacy format
    fmt.Fprintf(w, "ANI1:%s,DNIS1:%s", resp.ANIToSend, resp.DNISToSend)
}

func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
    stats, err := s.router.GetStatistics()
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(stats)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("OK"))
}
