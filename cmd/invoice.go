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

func printInvoiceHelp() {
	fmt.Println("Usage: simplebill invoice <customer> <product:qty[:discount[:@price]]>... [-y]")
	fmt.Println()
	fmt.Println("Generate a PDF invoice for a customer.")
	fmt.Println()
	fmt.Println("Arguments:")
	fmt.Println("  customer                     Customer key from customers.yml")
	fmt.Println("  product:qty                  Product key and quantity (e.g., widget:10)")
	fmt.Println("  product:qty:discount         Percentage discount (e.g., widget:1:25 for 25% off)")
	fmt.Println("  product:qty:discount:@price  Custom price with optional discount (e.g., widget:1:0:@15.00)")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  -y, --yes    Skip preview and save immediately")
	fmt.Println("  -h, --help   Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  simplebill invoice acme widget:10")
	fmt.Println("  simplebill invoice acme widget:5 widget:1:25 gadget:3 -y")
	fmt.Println("  simplebill invoice acme widget:10:0:@15.00")
}

func RunInvoice(args []string) error {
	// Check for flags
	skipPreview := false
	var filteredArgs []string
	for _, arg := range args {
		if arg == "-y" || arg == "--yes" {
			skipPreview = true
		} else if arg == "-h" || arg == "--help" {
			printInvoiceHelp()
			return nil
		} else {
			filteredArgs = append(filteredArgs, arg)
		}
	}
	args = filteredArgs

	if len(args) < 2 {
		printInvoiceHelp()
		return nil
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
		if len(parts) < 2 || len(parts) > 4 {
			return fmt.Errorf("invalid format '%s', expected product:qty or product:qty:discount or product:qty:discount:@price", arg)
		}

		productKey := parts[0]
		qty, err := strconv.Atoi(parts[1])
		if err != nil {
			return fmt.Errorf("invalid quantity '%s' for product '%s'", parts[1], productKey)
		}

		var discount int
		if len(parts) >= 3 {
			discount, err = strconv.Atoi(parts[2])
			if err != nil || discount < 0 || discount > 100 {
				return fmt.Errorf("invalid discount '%s' for product '%s', expected 0-100", parts[2], productKey)
			}
		}

		var customPrice float64
		if len(parts) == 4 {
			priceStr := parts[3]
			if !strings.HasPrefix(priceStr, "@") {
				return fmt.Errorf("invalid price '%s' for product '%s', expected @price (e.g., @15.00)", priceStr, productKey)
			}
			customPrice, err = strconv.ParseFloat(priceStr[1:], 64)
			if err != nil || customPrice < 0 {
				return fmt.Errorf("invalid price '%s' for product '%s'", priceStr, productKey)
			}
		}

		product, ok := products[productKey]
		if !ok {
			return fmt.Errorf("product '%s' not found in products.yml", productKey)
		}

		basePrice := product.Price
		if customPrice > 0 {
			basePrice = customPrice
		}
		unitPrice := basePrice
		if discount > 0 {
			unitPrice = basePrice * (1 - float64(discount)/100)
		}
		itemTotal := unitPrice * float64(qty)
		items = append(items, invoice.Item{
			Product:   productKey,
			Quantity:  qty,
			UnitPrice: unitPrice,
			Total:     itemTotal,
			Discount:  discount,
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
