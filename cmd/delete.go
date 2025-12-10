package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"simplebill/internal/config"
)

func printDeleteHelp() {
	fmt.Println("Usage: simplebill delete <invoice-number> [--confirm]")
	fmt.Println()
	fmt.Println("Delete an invoice (both .yml and .pdf files).")
	fmt.Println()
	fmt.Println("Arguments:")
	fmt.Println("  invoice-number   Invoice number to delete (e.g., INV-2025-0001)")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  -y, --confirm    Skip confirmation prompt")
	fmt.Println("  -h, --help       Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  simplebill delete INV-2025-0001")
	fmt.Println("  simplebill delete INV-2025-0001 --confirm")
}

func RunDelete(args []string) error {
	if len(args) == 0 {
		printDeleteHelp()
		return nil
	}

	var invoiceNumber string
	var confirmed bool

	for _, arg := range args {
		if arg == "--confirm" || arg == "-y" {
			confirmed = true
		} else if arg == "-h" || arg == "--help" {
			printDeleteHelp()
			return nil
		} else {
			invoiceNumber = arg
		}
	}

	if invoiceNumber == "" {
		printDeleteHelp()
		return nil
	}

	dir, err := config.Dir()
	if err != nil {
		return err
	}

	invoicesDir := filepath.Join(dir, "invoices")
	ymlPath := filepath.Join(invoicesDir, invoiceNumber+".yml")
	pdfPath := filepath.Join(invoicesDir, invoiceNumber+".pdf")

	// Check if invoice exists
	if _, err := os.Stat(ymlPath); os.IsNotExist(err) {
		return fmt.Errorf("invoice %s not found", invoiceNumber)
	}

	// Prompt for confirmation if not already confirmed
	if !confirmed {
		fmt.Printf("Delete invoice %s? [y/N] ", invoiceNumber)
		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(strings.ToLower(response))
		if response != "y" && response != "yes" {
			fmt.Println("Cancelled.")
			return nil
		}
	}

	// Delete yml file
	if err := os.Remove(ymlPath); err != nil {
		return fmt.Errorf("deleting invoice: %w", err)
	}

	// Delete pdf file if it exists
	if _, err := os.Stat(pdfPath); err == nil {
		if err := os.Remove(pdfPath); err != nil {
			return fmt.Errorf("deleting PDF: %w", err)
		}
	}

	fmt.Printf("Deleted %s\n", invoiceNumber)
	config.AutoCommit(fmt.Sprintf("simplebill: deleted invoice %s", invoiceNumber))

	return nil
}
