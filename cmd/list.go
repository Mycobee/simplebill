package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
	"simplebill/internal/config"
	"simplebill/internal/invoice"
)

func RunList(args []string) error {
	if len(args) == 0 {
		return listInvoices()
	}

	switch args[0] {
	case "invoices":
		return listInvoices()
	case "customers":
		return listCustomers()
	case "products":
		return listProducts()
	case "config":
		return listConfig()
	default:
		return fmt.Errorf("unknown list type '%s'. Use: invoices, customers, products, config", args[0])
	}
}

func listInvoices() error {
	dir, err := config.Dir()
	if err != nil {
		return err
	}

	invoicesDir := filepath.Join(dir, "invoices")
	entries, err := os.ReadDir(invoicesDir)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("No invoices yet.")
			return nil
		}
		return err
	}

	customers, err := config.LoadCustomers()
	if err != nil {
		return err
	}

	var invoices []invoice.Invoice
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".yml") {
			continue
		}

		path := filepath.Join(invoicesDir, entry.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		var inv invoice.Invoice
		if err := yaml.Unmarshal(data, &inv); err != nil {
			continue
		}
		invoices = append(invoices, inv)
	}

	if len(invoices) == 0 {
		fmt.Println("No invoices yet.")
		return nil
	}

	// Sort by date descending (newest first)
	sort.Slice(invoices, func(i, j int) bool {
		return invoices[i].Date > invoices[j].Date
	})

	for _, inv := range invoices {
		customerName := inv.Customer
		if c, ok := customers[inv.Customer]; ok {
			customerName = c.Name
		}
		fmt.Printf("%-15s  %s  %-30s  $%7.2f\n",
			inv.InvoiceNumber, inv.Date, customerName, inv.Total)
	}

	return nil
}

func listCustomers() error {
	customers, err := config.LoadCustomers()
	if err != nil {
		return err
	}

	if len(customers) == 0 {
		fmt.Println("No customers defined.")
		return nil
	}

	// Sort keys for consistent output
	var keys []string
	for k := range customers {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		c := customers[k]
		fmt.Printf("%-15s  %s\n", k, c.Name)
	}

	return nil
}

func listProducts() error {
	products, err := config.LoadProducts()
	if err != nil {
		return err
	}

	if len(products) == 0 {
		fmt.Println("No products defined.")
		return nil
	}

	// Sort keys for consistent output
	var keys []string
	for k := range products {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		p := products[k]
		fmt.Printf("%-15s  %-40s  $%.2f\n", k, p.Name, p.Price)
	}

	return nil
}

func listConfig() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	fmt.Println("Company:")
	fmt.Printf("  Name:    %s\n", cfg.Company.Name)
	fmt.Printf("  Address: %s\n", strings.ReplaceAll(strings.TrimSpace(cfg.Company.Address), "\n", ", "))
	fmt.Printf("  Email:   %s\n", cfg.Company.Email)
	if cfg.Company.Phone != "" {
		fmt.Printf("  Phone:   %s\n", cfg.Company.Phone)
	}
	if cfg.Company.ID != "" {
		fmt.Printf("  ID:      %s\n", cfg.Company.ID)
	}

	fmt.Println()
	fmt.Println("Invoice Settings:")
	fmt.Printf("  Prefix:        %s\n", cfg.Invoice.Prefix)
	fmt.Printf("  Starting #:    %s\n", cfg.Invoice.StartingNumber)
	fmt.Printf("  Payment Terms: %s\n", cfg.Invoice.PaymentTerms)
	fmt.Printf("  Due Days:      %d\n", cfg.Invoice.DueDays)
	fmt.Printf("  Notes:         %s\n", cfg.Invoice.Notes)

	fmt.Println()
	fmt.Printf("Auto-commit: %v\n", cfg.AutoCommit)

	return nil
}
