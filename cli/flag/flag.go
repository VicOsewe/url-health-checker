package flag

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/VicOsewe/url-health-checker/internal/checker"
	"github.com/VicOsewe/url-health-checker/internal/exporter"
)

type FlagCLI struct{}

func New() *FlagCLI {
	return &FlagCLI{}
}

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

	ctx := context.Background()

	fmt.Printf("Checking %d URL(s) [timeout: %ds, format: %s]\n", len(urls), *timeout, *format)

	resultCh := make(chan checker.Result, len(urls))

	for _, url := range urls {
		go func(u string) {
			resultCh <- checker.Check(ctx, u, time.Duration(*timeout)*time.Second)
		}(url)
	}

	var results []checker.Result
	for range urls {
		results = append(results, <-resultCh)
	}

	base := exporter.New(*format, os.Stdout)
	exp := &exporter.LoggingExporter{
		Writer:   os.Stderr,
		Exporter: base,
	}

	if err := exp.Export(results); err != nil {
		return err
	}

	return nil
}
