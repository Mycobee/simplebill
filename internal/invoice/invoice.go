package invoice

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"time"

	"gopkg.in/yaml.v3"
	"simplebill/internal/config"
)

type Invoice struct {
	InvoiceNumber string    `yaml:"invoice_number"`
	Date          string    `yaml:"date"`
	DueDate       string    `yaml:"due_date"`
	Customer      string    `yaml:"customer"`
	Items         []Item    `yaml:"items"`
	Total         float64   `yaml:"total"`
	CreatedAt     time.Time `yaml:"created_at"`
}

type Item struct {
	Product   string  `yaml:"product"`
	Quantity  int     `yaml:"quantity"`
	UnitPrice float64 `yaml:"unit_price"`
	Total     float64 `yaml:"total"`
}

// NextNumber determines the next invoice number based on existing invoices or starting_number
func NextNumber(cfg *config.Config) (string, error) {
	dir, err := config.Dir()
	if err != nil {
		return "", err
	}

	invoicesDir := filepath.Join(dir, "invoices")
	year := time.Now().Year()
	prefix := cfg.Invoice.Prefix

	// Pattern to match invoice files: PREFIX-YEAR-NNNN.yml
	pattern := regexp.MustCompile(fmt.Sprintf(`^%s-%d-(\d{4})\.yml$`, regexp.QuoteMeta(prefix), year))

	startingNum, _ := strconv.Atoi(cfg.Invoice.StartingNumber)

	entries, err := os.ReadDir(invoicesDir)
	if err != nil {
		if os.IsNotExist(err) {
			// No invoices directory, use starting number
			return fmt.Sprintf("%s-%d-%04d", prefix, year, startingNum+1), nil
		}
		return "", err
	}

	maxSeq := startingNum
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		matches := pattern.FindStringSubmatch(entry.Name())
		if matches != nil {
			seq, _ := strconv.Atoi(matches[1])
			if seq > maxSeq {
				maxSeq = seq
			}
		}
	}

	return fmt.Sprintf("%s-%d-%04d", prefix, year, maxSeq+1), nil
}

// Save writes the invoice to a YAML file
func (inv *Invoice) Save() error {
	dir, err := config.Dir()
	if err != nil {
		return err
	}

	path := filepath.Join(dir, "invoices", inv.InvoiceNumber+".yml")
	data, err := yaml.Marshal(inv)
	if err != nil {
		return fmt.Errorf("marshaling invoice: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("writing invoice: %w", err)
	}

	return nil
}
