package config

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Company    Company       `yaml:"company"`
	Invoice    InvoiceConfig `yaml:"invoice"`
	AutoCommit bool          `yaml:"auto_commit"`
}

type Company struct {
	Name    string `yaml:"name"`
	Address string `yaml:"address"`
	Email   string `yaml:"email"`
	Phone   string `yaml:"phone"`
	ID      string `yaml:"id"`
}

type InvoiceConfig struct {
	Prefix         string `yaml:"prefix"`
	StartingNumber string `yaml:"starting_number"`
	PaymentTerms   string `yaml:"payment_terms"`
	DueDays        int    `yaml:"due_days"`
	Notes          string `yaml:"notes"`
}

type Customer struct {
	Name    string `yaml:"name"`
	Address string `yaml:"address"`
	Email   string `yaml:"email"`
	Phone   string `yaml:"phone"`
	ID      string `yaml:"id"`
}

type Product struct {
	Name  string  `yaml:"name"`
	SKU   string  `yaml:"sku"`
	Price float64 `yaml:"price"`
}

func Dir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".simplebill"), nil
}

func Load() (*Config, error) {
	dir, err := Dir()
	if err != nil {
		return nil, err
	}

	path := filepath.Join(dir, "config.yml")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("run 'simplebill init' first")
		}
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", path, err)
	}

	return &cfg, nil
}

func LoadCustomers() (map[string]Customer, error) {
	dir, err := Dir()
	if err != nil {
		return nil, err
	}

	path := filepath.Join(dir, "customers.yml")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}

	var customers map[string]Customer
	if err := yaml.Unmarshal(data, &customers); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", path, err)
	}

	return customers, nil
}

func LoadProducts() (map[string]Product, error) {
	dir, err := Dir()
	if err != nil {
		return nil, err
	}

	path := filepath.Join(dir, "products.yml")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}

	var products map[string]Product
	if err := yaml.Unmarshal(data, &products); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", path, err)
	}

	return products, nil
}

// AutoCommit commits all changes if auto_commit is enabled and ~/.simplebill is a git repo
func AutoCommit(message string) error {
	cfg, err := Load()
	if err != nil {
		return nil // silently skip if config can't be loaded
	}

	if !cfg.AutoCommit {
		return nil
	}

	dir, err := Dir()
	if err != nil {
		return nil
	}

	// Check if it's a git repo
	gitDir := filepath.Join(dir, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		return nil // not a git repo, skip
	}

	// git add -A
	addCmd := exec.Command("git", "-C", dir, "add", "-A")
	if err := addCmd.Run(); err != nil {
		return fmt.Errorf("git add failed: %w", err)
	}

	// Check if there are changes to commit
	diffCmd := exec.Command("git", "-C", dir, "diff", "--cached", "--quiet")
	if err := diffCmd.Run(); err == nil {
		return nil // no changes to commit
	}

	// git commit
	commitCmd := exec.Command("git", "-C", dir, "commit", "-m", message)
	if err := commitCmd.Run(); err != nil {
		return fmt.Errorf("git commit failed: %w", err)
	}

	return nil
}
