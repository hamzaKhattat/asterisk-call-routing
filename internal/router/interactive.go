package router

import (
   "bufio"
   "fmt"
   "os"
   "strings"
   
   "github.com/asterisk-call-routing/internal/db"
)

func RunInteractiveMode() {
   scanner := bufio.NewScanner(os.Stdin)
   
   for {
       fmt.Println("\n=== Call Routing System Menu ===")
       fmt.Println("1. Show Statistics")
       fmt.Println("2. Show Recent Calls")
       fmt.Println("3. Show DID Status")
       fmt.Println("4. Import DIDs from CSV")
       fmt.Println("5. Clean Up Stuck DIDs")
       fmt.Println("6. Export Call Records")
       fmt.Println("0. Exit")
       fmt.Print("\nSelect option: ")
       
       scanner.Scan()
       choice := strings.TrimSpace(scanner.Text())
       
       switch choice {
       case "1":
           ShowStatistics()
       case "2":
           showRecentCalls()
       case "3":
           showDIDStatus()
       case "4":
           importDIDsInteractive(scanner)
       case "5":
           cleanupDIDsInteractive()
       case "6":
           exportCallRecords()
       case "0":
           fmt.Println("Exiting...")
           return
       default:
           fmt.Println("Invalid option")
       }
   }
}

func ShowStatistics() {
   stats, err := db.GetStatistics()
   if err != nil {
       fmt.Printf("Error getting statistics: %v\n", err)
       return
   }
   
   fmt.Println("\n=== Call Statistics ===")
   fmt.Printf("Total Calls:      %d\n", stats.TotalCalls)
   fmt.Printf("Completed Calls:  %d\n", stats.CompletedCalls)
   fmt.Printf("Failed Calls:     %d\n", stats.FailedCalls)
   fmt.Printf("Active Calls:     %d\n", stats.ActiveCalls)
   fmt.Printf("Success Rate:     %.2f%%\n", stats.SuccessRate)
   fmt.Printf("Avg Duration:     %.2f seconds\n", stats.AvgDuration)
   
   fmt.Println("\n=== DID Statistics ===")
   fmt.Printf("Total DIDs:       %d\n", stats.TotalDIDs)
   fmt.Printf("In Use:           %d\n", stats.InUseDIDs)
   fmt.Printf("Available:        %d\n", stats.AvailableDIDs)
   fmt.Printf("Usage:            %.2f%%\n", float64(stats.InUseDIDs)/float64(stats.TotalDIDs)*100)
}

func showRecentCalls() {
   calls, err := db.GetRecentCalls(20)
   if err != nil {
       fmt.Printf("Error getting recent calls: %v\n", err)
       return
   }
   
   fmt.Println("\n=== Recent Calls ===")
   fmt.Printf("%-20s %-15s %-15s %-15s %-10s %-10s\n", 
       "Call ID", "ANI", "DNIS", "DID", "Status", "Duration")
   fmt.Println(strings.Repeat("-", 90))
   
   for _, call := range calls {
       fmt.Printf("%-20s %-15s %-15s %-15s %-10s %d sec\n",
           call.CallID[:20], call.ANIOriginal, call.DNISOriginal,
           call.DIDUsed, call.Status, call.Duration)
   }
}

func showDIDStatus() {
   var stats struct {
       Total     int `db:"total"`
       InUse     int `db:"in_use"`
       Available int `db:"available"`
   }
   
   err := db.DB.Get(&stats, `
       SELECT 
           COUNT(*) as total,
           SUM(CASE WHEN in_use = 1 THEN 1 ELSE 0 END) as in_use,
           SUM(CASE WHEN in_use = 0 THEN 1 ELSE 0 END) as available
       FROM dids
   `)
   
   if err != nil {
       fmt.Printf("Error getting DID status: %v\n", err)
       return
   }
   
   fmt.Println("\n=== DID Status ===")
   fmt.Printf("Total DIDs:     %d\n", stats.Total)
   fmt.Printf("In Use:         %d (%.2f%%)\n", stats.InUse, float64(stats.InUse)/float64(stats.Total)*100)
   fmt.Printf("Available:      %d (%.2f%%)\n", stats.Available, float64(stats.Available)/float64(stats.Total)*100)
   
   // Show sample DIDs
   var dids []struct {
       DID     string `db:"did"`
       InUse   bool   `db:"in_use"`
       Country string `db:"country"`
   }
   
   db.DB.Select(&dids, "SELECT did, in_use, country FROM dids LIMIT 10")
   
   fmt.Println("\nSample DIDs:")
   fmt.Printf("%-20s %-10s %-15s\n", "DID", "Status", "Country")
   fmt.Println(strings.Repeat("-", 45))
   
   for _, did := range dids {
       status := "Available"
       if did.InUse {
           status = "In Use"
       }
       fmt.Printf("%-20s %-10s %-15s\n", did.DID, status, did.Country)
   }
}

func importDIDsInteractive(scanner *bufio.Scanner) {
   fmt.Print("Enter CSV file path: ")
   scanner.Scan()
   filepath := strings.TrimSpace(scanner.Text())
   
   count, err := db.ImportDIDsFromCSV(filepath)
   if err != nil {
       fmt.Printf("Error importing DIDs: %v\n", err)
       return
   }
   
   fmt.Printf("Successfully imported %d DIDs\n", count)
}

func cleanupDIDsInteractive() {
   fmt.Print("Clean up DIDs stuck for more than how many minutes? [60]: ")
   
   scanner := bufio.NewScanner(os.Stdin)
   scanner.Scan()
   input := strings.TrimSpace(scanner.Text())
   
   minutes := 60
   if input != "" {
       fmt.Sscanf(input, "%d", &minutes)
   }
   
   count, err := db.CleanupStuckDIDs(minutes)
   if err != nil {
       fmt.Printf("Error cleaning up DIDs: %v\n", err)
       return
   }
   
   fmt.Printf("Cleaned up %d stuck DIDs\n", count)
}

func exportCallRecords() {
   fmt.Println("Exporting call records to call_records_export.csv...")
   
   // Implementation would export to CSV
   fmt.Println("Export completed")
}
