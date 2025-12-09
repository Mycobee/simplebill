package cmd

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"

	"simplebill/internal/config"
)

//go:embed templates/*
var templates embed.FS

var defaultConfig = `company:
  name: "Your Company Name"
  address: |
    123 Main St
    City, ST 12345
  email: "billing@example.com"
  phone: ""
  id: ""

invoice:
  prefix: "INV"
  starting_number: "0000"  # set to last invoice number (next will be +1)
  payment_terms: "Net 14"
  due_days: 14
  notes: "Thank you for your business!"

# If true and ~/.simplebill is a git repo, auto-commit after changes
auto_commit: false
`

var defaultCustomers = `# Add customers here. The key (e.g., "acme") is used on the command line.
# Example:
#
# acme:
#   name: "Acme Corp"
#   email: "billing@acme.com"
#   phone: "555-123-4567"
#   address: |
#     456 Oak Ave
#     Denver, CO 80202
#   id: "LIC-12345"
`

var defaultProducts = `# Add products here. The key (e.g., "widget") is used on the command line.
# Example:
#
# widget:
#   name: "Standard Widget"
#   sku: "WDG-001"
#   price: 19.99
`

func RunInit() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("could not find home directory: %w", err)
	}

	dir := filepath.Join(home, ".simplebill")

	if _, err := os.Stat(dir); err == nil {
		return fmt.Errorf("%s already exists", dir)
	}

	if err := os.MkdirAll(filepath.Join(dir, "invoices"), 0755); err != nil {
		return fmt.Errorf("could not create directory: %w", err)
	}

	files := map[string]string{
		"config.yml":    defaultConfig,
		"customers.yml": defaultCustomers,
		"products.yml":  defaultProducts,
	}

	for name, content := range files {
		path := filepath.Join(dir, name)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			return fmt.Errorf("could not write %s: %w", name, err)
		}
	}

	templateContent, err := templates.ReadFile("templates/invoice.html")
	if err != nil {
		return fmt.Errorf("could not read embedded template: %w", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "template.html"), templateContent, 0644); err != nil {
		return fmt.Errorf("could not write template.html: %w", err)
	}

	fmt.Printf("Created %s\n", dir)
	fmt.Println("Edit your config files there, then run: simplebill invoice <customer> <product:qty>")

	config.AutoCommit("simplebill: initialized")
	return nil
}
