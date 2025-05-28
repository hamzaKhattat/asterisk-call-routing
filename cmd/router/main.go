package main

import (
    "flag"
    "fmt"
    "log"
    "os"
    "os/signal"
    "syscall"
    
    "github.com/asterisk-call-routing/internal/api"
    "github.com/asterisk-call-routing/internal/db"
    "github.com/asterisk-call-routing/internal/router"
)

func main() {
    // Command line flags
    var (
        httpPort = flag.Int("port", 8001, "HTTP server port")
        dbHost   = flag.String("dbhost", "localhost", "MySQL host")
        dbPort   = flag.Int("dbport", 3306, "MySQL port")
        dbUser   = flag.String("dbuser", "root", "MySQL user")
        dbPass   = flag.String("dbpass", "temppass", "MySQL password")
        dbName   = flag.String("dbname", "call_routing", "MySQL database name")
    )
    flag.Parse()
    
    // Setup logging
    log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)
    log.Printf("Starting S2 Dynamic Call Router...")
    
    // Build database DSN
    dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true",
        *dbUser, *dbPass, *dbHost, *dbPort, *dbName)
    
    // Initialize database
    database, err := db.NewDatabase(dsn)
    if err != nil {
        log.Fatalf("Failed to connect to database: %v", err)
    }
    defer database.Close()
    
    // Initialize router
    r := router.NewRouter(database)
    
    // Start API server
    apiServer := api.NewServer(r, *httpPort)
    go func() {
        if err := apiServer.Start(); err != nil {
            log.Fatalf("API server failed: %v", err)
        }
    }()
    
    log.Printf("S2 Router started successfully on port %d", *httpPort)
    
    // Wait for interrupt signal
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
    <-sigChan
    
    log.Println("Shutting down...")
}
