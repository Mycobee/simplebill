package cmd

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"simplebill/internal/config"
	"simplebill/internal/invoice"
)

// TemplateData holds all data passed to the HTML template
type TemplateData struct {
	InvoiceNumber string
	Date          string
	DueDate       string
	Company       config.Company
	Customer      config.Customer
	PaymentTerms  string
	Notes         string
	Items         []TemplateItem
	Total         float64
}

// TemplateItem holds item data for the template
type TemplateItem struct {
	Name     string
	SKU      string
	Quantity int
	Price    float64
	Total    float64
}

func buildTemplateData(inv *invoice.Invoice, cfg *config.Config, customer *config.Customer, products map[string]config.Product) TemplateData {
	var items []TemplateItem
	for _, item := range inv.Items {
		prod := products[item.Product]
		items = append(items, TemplateItem{
			Name:     prod.Name,
			SKU:      prod.SKU,
			Quantity: item.Quantity,
			Price:    item.UnitPrice,
			Total:    item.Total,
		})
	}

	return TemplateData{
		InvoiceNumber: inv.InvoiceNumber,
		Date:          inv.Date,
		DueDate:       inv.DueDate,
		Company:       cfg.Company,
		Customer:      *customer,
		PaymentTerms:  cfg.Invoice.PaymentTerms,
		Notes:         cfg.Invoice.Notes,
		Items:         items,
		Total:         inv.Total,
	}
}

func renderHTML(data TemplateData) ([]byte, error) {
	dir, err := config.Dir()
	if err != nil {
		return nil, err
	}

	tmplPath := filepath.Join(dir, "template.html")
	tmplContent, err := os.ReadFile(tmplPath)
	if err != nil {
		return nil, fmt.Errorf("reading template: %w", err)
	}

	tmpl, err := template.New("invoice").Parse(string(tmplContent))
	if err != nil {
		return nil, fmt.Errorf("parsing template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("executing template: %w", err)
	}

	return buf.Bytes(), nil
}

func runWkhtmltopdf(htmlPath, pdfPath string) error {
	if _, err := exec.LookPath("wkhtmltopdf"); err != nil {
		return fmt.Errorf("wkhtmltopdf not installed\n\nInstall it with:\n  macOS: brew install wkhtmltopdf\n  Ubuntu/Debian: sudo apt install wkhtmltopdf\n  Fedora: sudo dnf install wkhtmltopdf")
	}

	args := []string{
		"--page-size", "Letter",
		"--margin-top", "10mm",
		"--margin-bottom", "10mm",
		"--margin-left", "10mm",
		"--margin-right", "10mm",
	}

	if runtime.GOOS != "windows" {
		args = append(args, "--quiet")
	}

	args = append(args, htmlPath, pdfPath)

	cmd := exec.Command("wkhtmltopdf", args...)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("wkhtmltopdf failed: %s\n%s", err, output)
	}

	return nil
}

// RenderPDF renders invoice to final PDF location. If outputPath is empty, uses default location.
func RenderPDF(inv *invoice.Invoice, cfg *config.Config, customer *config.Customer, products map[string]config.Product, outputPath string) error {
	data := buildTemplateData(inv, cfg, customer, products)
	html, err := renderHTML(data)
	if err != nil {
		return err
	}

	// Write HTML to temp file
	tmpFile, err := os.CreateTemp("", "simplebill-*.html")
	if err != nil {
		return fmt.Errorf("creating temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write(html); err != nil {
		tmpFile.Close()
		return fmt.Errorf("writing temp file: %w", err)
	}
	tmpFile.Close()

	// Determine output path
	if outputPath == "" {
		dir, err := config.Dir()
		if err != nil {
			return err
		}
		outputPath = filepath.Join(dir, "invoices", inv.InvoiceNumber+".pdf")
	}

	return runWkhtmltopdf(tmpFile.Name(), outputPath)
}

// RenderPDFToTemp renders invoice to a temp PDF file and returns the path
func RenderPDFToTemp(inv *invoice.Invoice, cfg *config.Config, customer *config.Customer, products map[string]config.Product) (string, error) {
	data := buildTemplateData(inv, cfg, customer, products)
	html, err := renderHTML(data)
	if err != nil {
		return "", err
	}

	// Write HTML to temp file
	tmpHTML, err := os.CreateTemp("", "simplebill-*.html")
	if err != nil {
		return "", fmt.Errorf("creating temp file: %w", err)
	}
	defer os.Remove(tmpHTML.Name())

	if _, err := tmpHTML.Write(html); err != nil {
		tmpHTML.Close()
		return "", fmt.Errorf("writing temp file: %w", err)
	}
	tmpHTML.Close()

	// Create temp PDF file
	tmpPDF, err := os.CreateTemp("", "simplebill-preview-*.pdf")
	if err != nil {
		return "", fmt.Errorf("creating temp PDF: %w", err)
	}
	tmpPDF.Close()

	if err := runWkhtmltopdf(tmpHTML.Name(), tmpPDF.Name()); err != nil {
		os.Remove(tmpPDF.Name())
		return "", err
	}

	return tmpPDF.Name(), nil
}

// openFile opens a file with the default system application
func openFile(path string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", path)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", path)
	default: // linux, freebsd, etc.
		cmd = exec.Command("xdg-open", path)
	}
	return cmd.Start()
}
