package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"simplebill/internal/config"
	"simplebill/internal/invoice"
)

func RunInvoice(args []string) error {
	// Check for -y flag
	skipPreview := false
	var filteredArgs []string
	for _, arg := range args {
		if arg == "-y" || arg == "--yes" {
			skipPreview = true
		} else {
			filteredArgs = append(filteredArgs, arg)
		}
	}
	args = filteredArgs

	if len(args) < 2 {
		return fmt.Errorf("usage: simplebill invoice <customer> <product:qty> [product:qty...] [-y]")
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	customers, err := config.LoadCustomers()
	if err != nil {
		return err
	}

	products, err := config.LoadProducts()
	if err != nil {
		return err
	}

	// Validate customer
	customerKey := args[0]
	customer, ok := customers[customerKey]
	if !ok {
		return fmt.Errorf("customer '%s' not found in customers.yml", customerKey)
	}

	// Parse product:qty pairs
	var items []invoice.Item
	var total float64

	for _, arg := range args[1:] {
		parts := strings.Split(arg, ":")
		if len(parts) != 2 {
			return fmt.Errorf("invalid format '%s', expected product:quantity", arg)
		}

		productKey := parts[0]
		qty, err := strconv.Atoi(parts[1])
		if err != nil {
			return fmt.Errorf("invalid quantity '%s' for product '%s'", parts[1], productKey)
		}

		product, ok := products[productKey]
		if !ok {
			return fmt.Errorf("product '%s' not found in products.yml", productKey)
		}

		itemTotal := product.Price * float64(qty)
		items = append(items, invoice.Item{
			Product:   productKey,
			Quantity:  qty,
			UnitPrice: product.Price,
			Total:     itemTotal,
		})
		total += itemTotal
	}

	// Generate invoice number
	invNumber, err := invoice.NextNumber(cfg)
	if err != nil {
		return fmt.Errorf("generating invoice number: %w", err)
	}

	// Create invoice
	now := time.Now()
	dueDate := now.AddDate(0, 0, cfg.Invoice.DueDays)

	inv := &invoice.Invoice{
		InvoiceNumber: invNumber,
		Date:          now.Format("2006-01-02"),
		DueDate:       dueDate.Format("2006-01-02"),
		Status:        "draft",
		Customer:      customerKey,
		Items:         items,
		Total:         total,
		CreatedAt:     now,
	}

	dir, _ := config.Dir()

	if skipPreview {
		// Save directly without preview
		if err := inv.Save(); err != nil {
			return err
		}
		if err := RenderPDF(inv, cfg, &customer, products, ""); err != nil {
			return err
		}
		fmt.Printf("Created %s\n", invNumber)
		fmt.Printf("%s/invoices/%s.pdf\n", dir, invNumber)
		config.AutoCommit(fmt.Sprintf("simplebill: created invoice %s", invNumber))
		return nil
	}

	// Preview flow: render to temp, open, prompt
	tempPDF, err := RenderPDFToTemp(inv, cfg, &customer, products)
	if err != nil {
		return err
	}
	defer os.Remove(tempPDF)

	// Open in default viewer
	if err := openFile(tempPDF); err != nil {
		return fmt.Errorf("opening preview: %w", err)
	}

	// Prompt user
	fmt.Print("Save invoice? [y/n]: ")
	reader := bufio.NewReader(os.Stdin)
	response, _ := reader.ReadString('\n')
	response = strings.TrimSpace(strings.ToLower(response))

	if response != "y" && response != "yes" {
		fmt.Println("Invoice cancelled.")
		return nil
	}

	// Save invoice YAML
	if err := inv.Save(); err != nil {
		return err
	}

	// Save final PDF
	if err := RenderPDF(inv, cfg, &customer, products, ""); err != nil {
		return err
	}

	fmt.Printf("Created %s\n", invNumber)
	fmt.Printf("%s/invoices/%s.pdf\n", dir, invNumber)

	// Auto-commit if enabled
	config.AutoCommit(fmt.Sprintf("simplebill: created invoice %s", invNumber))

	return nil
}
