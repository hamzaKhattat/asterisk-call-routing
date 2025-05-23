package monitor

import (
   "context"
   "log"
   "runtime"
   "sync"
   "time"
)

type Monitor struct {
   startTime time.Time
   events    []Event
   eventsMu  sync.RWMutex
   ctx       context.Context
   cancel    context.CancelFunc
}

type Event struct {
   Timestamp time.Time
   Type      string
   Data      map[string]interface{}
}

func NewMonitor() *Monitor {
   ctx, cancel := context.WithCancel(context.Background())
   
   return &Monitor{
       startTime: time.Now(),
       events:    make([]Event, 0),
       ctx:       ctx,
       cancel:    cancel,
   }
}

func (m *Monitor) Start() {
   go m.monitorRoutine()
}

func (m *Monitor) Stop() {
   m.cancel()
}

func (m *Monitor) RecordEvent(eventType string, data map[string]interface{}) {
   m.eventsMu.Lock()
   defer m.eventsMu.Unlock()
   
   event := Event{
       Timestamp: time.Now(),
       Type:      eventType,
       Data:      data,
   }
   
   m.events = append(m.events, event)
   
   // Keep only last 1000 events
   if len(m.events) > 1000 {
       m.events = m.events[len(m.events)-1000:]
   }
}

func (m *Monitor) GetEvents(limit int) []Event {
   m.eventsMu.RLock()
   defer m.eventsMu.RUnlock()
   
   if limit > len(m.events) {
       limit = len(m.events)
   }
   
   return m.events[len(m.events)-limit:]
}

func (m *Monitor) GetUptime() string {
   return time.Since(m.startTime).String()
}

func (m *Monitor) GetSystemStats() map[string]interface{} {
   var memStats runtime.MemStats
   runtime.ReadMemStats(&memStats)
   
   return map[string]interface{}{
       "uptime":          m.GetUptime(),
       "goroutines":      runtime.NumGoroutine(),
       "memory_alloc_mb": memStats.Alloc / 1024 / 1024,
       "memory_sys_mb":   memStats.Sys / 1024 / 1024,
       "gc_runs":         memStats.NumGC,
   }
}

func (m *Monitor) monitorRoutine() {
   ticker := time.NewTicker(1 * time.Minute)
   defer ticker.Stop()
   
   for {
       select {
       case <-m.ctx.Done():
           return
       case <-ticker.C:
           stats := m.GetSystemStats()
           log.Printf("System stats: %+v", stats)
       }
   }
}
