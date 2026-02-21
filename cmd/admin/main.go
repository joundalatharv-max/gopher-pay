package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"gopherpay/internal/billing"
	"gopherpay/internal/config"
	"gopherpay/internal/db"
)

func main() {

	ctx := context.Background()

	// Load configuration from .env
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	// Connect DB
	database, err := db.NewPostgresConnection(ctx, cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close()

	// Billing
	reportRepo := billing.NewPostgresReportRepository(database)
	reportService := billing.NewReportService(reportRepo)

	// Check command
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {

	// ========================================
	// REPORT
	// ========================================

	case "report":

		reportCmd := flag.NewFlagSet("report", flag.ExitOnError)

		user := reportCmd.String("user", "", "Account number")
		output := reportCmd.String("output", "", "Output CSV filename")

		reportCmd.Parse(os.Args[2:])

		if *user == "" {
			fmt.Println("Usage:")
			fmt.Println("  report --user=ACC1001")
			os.Exit(1)
		}

		// Ensure Reports directory exists and write file there
		outDir := "Reports"
		if err := os.MkdirAll(outDir, 0o755); err != nil {
			fmt.Println("failed to create Reports dir:", err)
			os.Exit(1)
		}

		filename := *output
		if filename == "" {
			filename = fmt.Sprintf("%s_report.csv", *user)
		}

		// sanitize and place inside Reports directory
		filename = filepath.Base(filename)
		fullpath := filepath.Join(outDir, filename)

		err := reportService.GenerateReport(
			ctx,
			*user,
			fullpath,
		)

		if err != nil {
			fmt.Println("Report failed:", err)
			os.Exit(1)
		}

		fmt.Println("Report generated:", fullpath)

	// ========================================
	// UNKNOWN
	// ========================================

	default:

		fmt.Println("Unknown command:", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {

	fmt.Println("Usage:")
	fmt.Println("")
	fmt.Println("Transfer money:")
	fmt.Println("  gopherpay transfer ACC1001 ACC1002 500")
	fmt.Println("")
	fmt.Println("Generate report:")
	fmt.Println("  gopherpay report --user=ACC1001")
	fmt.Println("")
	fmt.Println("Generate report with custom filename:")
	fmt.Println("  gopherpay report --user=ACC1001 --output=myreport.csv")
}
