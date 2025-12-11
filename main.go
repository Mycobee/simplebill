package main

import (
	"fmt"
	"os"

	"simplebill/cmd"
)

const Version = "0.1.2"

func main() {
	cmd.CheckForUpdate(Version)

	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	var err error
	switch os.Args[1] {
	case "-h", "--help", "help":
		printUsage()
		return
	case "-v", "--version", "version":
		fmt.Printf("simplebill %s\n", Version)
		return
	case "init":
		err = cmd.RunInit()
	case "invoice":
		err = cmd.RunInvoice(os.Args[2:])
	case "list":
		err = cmd.RunList(os.Args[2:])
	case "delete":
		err = cmd.RunDelete(os.Args[2:])
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Usage: simplebill <command>")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  init                              Initialize ~/.simplebill/ directory")
	fmt.Println("  invoice <customer> <product:qty>  Generate an invoice")
	fmt.Println("  list [type]                       List data (default: invoices)")
	fmt.Println("  delete <invoice-number>           Delete an invoice")
}

