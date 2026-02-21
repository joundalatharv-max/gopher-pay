package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/google/uuid"

	"gopherpay/internal/billing"
	"gopherpay/internal/config"
	"gopherpay/internal/db"
	"gopherpay/internal/wallet"
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

	// Wallet
	walletRepo := wallet.NewPostgresRepository(database)
	walletService := wallet.NewWalletService(database, walletRepo)

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
	// TRANSFER
	// ========================================

	case "transfer":

		if len(os.Args) != 5 {
			fmt.Println("Usage:")
			fmt.Println("  transfer <fromAccountNumber> <toAccountNumber> <amount>")
			os.Exit(1)
		}

		fromAccount := os.Args[2]
		toAccount := os.Args[3]

		amount, err := strconv.ParseInt(os.Args[4], 10, 64)
		if err != nil {
			log.Fatal("Invalid amount")
		}

		// Validate amount is positive integer (cents)
		if amount <= 0 {
			log.Fatal("Amount must be a positive integer (cents)")
		}

		// Generate request ID
		requestID := uuid.New().String()

		err = walletService.Transfer(
			ctx,
			fromAccount,
			toAccount,
			amount,
			requestID,
		)

		if err != nil {
			fmt.Println("Transfer failed:", err)
			os.Exit(1)
		}

		fmt.Println("Transfer successful")
		fmt.Println("RequestID:", requestID)

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

		filename := *output

		if filename == "" {
			filename = fmt.Sprintf("%s_report.csv", *user)
		}

		err := reportService.GenerateReport(
			ctx,
			*user,
			filename,
		)

		if err != nil {
			fmt.Println("Report failed:", err)
			os.Exit(1)
		}

		fmt.Println("Report generated:", filename)

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
