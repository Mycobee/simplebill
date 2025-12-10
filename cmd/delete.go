package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"simplebill/internal/config"
)

func RunDelete(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: simplebill delete <invoice-number> [--confirm]")
	}

	var invoiceNumber string
	var confirmed bool

	for _, arg := range args {
		if arg == "--confirm" || arg == "-y" {
			confirmed = true
		} else {
			invoiceNumber = arg
		}
	}

	if invoiceNumber == "" {
		return fmt.Errorf("usage: simplebill delete <invoice-number> [--confirm]")
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
