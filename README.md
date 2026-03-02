# URL Health Checker

A fast, lightweight CLI tool for checking website health, measuring response times, and exporting results in multiple formats.

## Features

- Check HTTP/HTTPS health status for one or multiple URLs
- Measure response times with configurable timeouts
- Export results as **text**, **JSON**, or **CSV**

## Installation

```bash
git clone https://github.com/user/url-health-checker.git
cd url-health-checker
go build -o url-health-checker .
```

## Usage

```bash
# Check a single URL
url-health-checker check https://example.com

# Check multiple URLs
url-health-checker check https://example.com https://google.com

# Set a custom timeout (seconds)
url-health-checker check --timeout=5 https://example.com

# Export results as JSON
url-health-checker check --format=json https://example.com

# Export results as CSV
url-health-checker check --format=csv https://example.com

# Show help
url-health-checker help
url-health-checker check --help
```

## Project Structure

```
url-health-checker/
├── main.go                        # Entry point - wires CLI to business logic
├── cli/
│   ├── cli.go                     # CLI interface definition
│   ├── flag/
│   │   └── flag.go                # flag implementation (active)
│   └── cobra/
│       └── cobra.go               # cobra implementation (future swap)
├── internal/
│   ├── checker/
│   │   └── checker.go             # HTTP health check logic
│   └── exporter/
│       └── exporter.go            # Output formatters (text/json/csv)
├── .gitignore
├── go.mod
└── README.md
```



## Requirements

- Go 1.22+
