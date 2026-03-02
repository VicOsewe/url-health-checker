// Package flag implements the CLI interface using Go's standard flag library.
package flag

import (
	"flag"
	"fmt"
	"os"
)

// FlagCLI is the flag-based implementation of the CLI interface.
type FlagCLI struct{}

// New returns a new FlagCLI instance.
func New() *FlagCLI {
	return &FlagCLI{}
}

// Run parses arguments and routes to the correct command.
func (f *FlagCLI) Run(args []string) error {
	if len(args) < 1 {
		printUsage()
		return nil
	}

	switch args[0] {
	case "check":
		return runCheck(args[1:])
	case "help", "--help", "-h":
		printUsage()
		return nil
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", args[0])
		printUsage()
		os.Exit(1)
	}
	return nil
}

func printUsage() {
	fmt.Fprintf(os.Stderr, `URL Health Checker - Check website health and measure response times

USAGE:
  url-health-checker <command> [options]

COMMANDS:
  check     Check the health of one or more URLs
  help      Show this help message

Run 'url-health-checker check --help' for check options.
`)
}

func runCheck(args []string) error {
	fs := flag.NewFlagSet("check", flag.ExitOnError)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, `Usage: url-health-checker check [options] <url> [url2 ...]

Options:
  --timeout int    Request timeout in seconds (default 10)
  --format string  Output format: text, json, csv (default "text")

Examples:
  url-health-checker check https://example.com
  url-health-checker check --format=json https://example.com https://google.com
`)
	}

	timeout := fs.Int("timeout", 10, "Request timeout in seconds")
	format := fs.String("format", "text", "Output format: text, json, csv")

	if err := fs.Parse(args); err != nil {
		return err
	}

	urls := fs.Args()
	if len(urls) == 0 {
		fs.Usage()
		os.Exit(1)
	}

	fmt.Printf("Checking %d URL(s) [timeout: %ds, format: %s]\n", len(urls), *timeout, *format)
	for _, url := range urls {
		fmt.Printf("  → %s\n", url)
	}

	return nil
}
