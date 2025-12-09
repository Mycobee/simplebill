package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "init":
		runInit()
	case "invoice":
		runInvoice(os.Args[2:])
	case "list":
		runList()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Usage: simplebill <command>")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  init                              Initialize ~/.simplebill/ directory")
	fmt.Println("  invoice <customer> <product:qty>  Generate an invoice")
	fmt.Println("  list                              List all invoices")
}

func runInit() {
	fmt.Println("init command not implemented")
}

func runInvoice(args []string) {
	fmt.Println("invoice command not implemented")
}

func runList() {
	fmt.Println("list command not implemented")
}
