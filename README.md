# simplebill

I got tired of copy-pasting Excel invoices like a caveman. I didn't want to pay for invoicing software or self-host some "freemium" product masquerading as free and open source. I just wanted to send a bill. So I built this.

`simplebill` turns YAML files + an HTML template into clean PDF invoices. No cloud apps, no subscriptions, no vendor lock-in. Just your terminal.

- **One command:** `simplebill invoice acme widget:10` → done
- **YAML everything:** customers, products, config - all human-readable
- **Your design:** it's just HTML/CSS, make it look however you want
- **Git-friendly:** auto-commit every invoice if you want a paper trail

## Install

TODO

## Setup

```bash
simplebill init
```

Creates `~/.simplebill/` with config files. Edit them:

- `config.yml` - your company info, invoice settings
- `customers.yml` - customer list (key: name, address, email, etc.)
- `products.yml` - product catalog (key: name, sku, price)
- `template.html` - invoice HTML template

## Usage

### Create an invoice

```bash
simplebill invoice <customer> <product:qty> [product:qty...]
```

Opens preview, prompts to save. Use `-y` to skip preview.

```bash
simplebill invoice acme widget:10 gizmo:5
simplebill invoice acme widget:10 -y
```

### List data

```bash
simplebill list              # invoices (default)
simplebill list customers
simplebill list products
simplebill list config
```

## Dependencies

Requires [wkhtmltopdf](https://wkhtmltopdf.org/) for PDF generation:

```bash
# macOS
brew install wkhtmltopdf

# Debian/Ubuntu
sudo apt install wkhtmltopdf

# Fedora
sudo dnf install wkhtmltopdf
```

## License

GPL-3.0
