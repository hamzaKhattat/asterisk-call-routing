package main

import (
    "flag"
    "fmt"
    "log"
    "os"
    "os/signal"
    "syscall"

    "github.com/asterisk-call-routing/internal/ami"
    "github.com/asterisk-call-routing/internal/api"
    "github.com/asterisk-call-routing/internal/config"
    "github.com/asterisk-call-routing/internal/db"
    "github.com/asterisk-call-routing/internal/monitor"
    "github.com/asterisk-call-routing/internal/router"
)

var (
    configPath   string
    httpPort     int
    importDIDs   string
    showStats    bool
    cleanupDIDs  bool
    interactive  bool
)

func init() {
    flag.StringVar(&configPath, "config", "configs/config.json", "Path to configuration file")
    flag.IntVar(&httpPort, "port", 8000, "HTTP server port")
    flag.StringVar(&importDIDs, "import-dids", "", "Import DIDs from CSV file")
    flag.BoolVar(&showStats, "stats", false, "Show statistics")
    flag.BoolVar(&cleanupDIDs, "cleanup", false, "Clean up stuck DIDs")
    flag.BoolVar(&interactive, "i", false, "Interactive mode")
}

func main() {
    flag.Parse()

    // Load configuration
    cfg, err := config.LoadConfig(configPath)
    if err != nil {
        log.Fatalf("Error loading config: %v", err)
    }

    // Initialize database
    if err := db.Initialize(cfg); err != nil {
        log.Fatalf("Error initializing database: %v", err)
    }
    defer db.Close()

    // Handle command-line operations
    if importDIDs != "" {
        count, err := db.ImportDIDsFromCSV(importDIDs)
        if err != nil {
            log.Fatalf("Error importing DIDs: %v", err)
        }
        log.Printf("Successfully imported %d DIDs", count)
        return
    }

    if showStats {
        router.ShowStatistics()
        return
    }

    if cleanupDIDs {
        count, err := db.CleanupStuckDIDs(60) // 60 minutes threshold
        if err != nil {
            log.Fatalf("Error cleaning up DIDs: %v", err)
        }
        log.Printf("Cleaned up %d stuck DIDs", count)
        return
    }

    if interactive {
        router.RunInteractiveMode()
        return
    }

    // Initialize AMI connection
    if err := ami.Initialize(cfg); err != nil {
        log.Fatalf("Error initializing AMI: %v", err)
    }
    defer ami.Close()

    // Start monitoring
    mon := monitor.NewMonitor()
    go mon.Start()

    // Create and start router
    r := router.NewRouter(cfg, mon)
    go r.Start()

    // Start API server
    apiServer := api.NewServer(cfg, r, mon)
    go apiServer.Start(httpPort)

    log.Printf("Call routing system started on port %d", httpPort)

    // Wait for interrupt signal
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
    <-sigChan

    log.Println("Shutting down...")
    r.Stop()
    apiServer.Stop()
    mon.Stop()
}
