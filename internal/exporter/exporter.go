package exporter

import (
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/VicOsewe/url-health-checker/internal/checker"
)

type Exporter interface {
	Export(results []checker.Result) error
}

// --- Text Exporter ---

type TextExporter struct {
	Writer io.Writer
}

func (t *TextExporter) Export(results []checker.Result) error {
	for _, r := range results {
		if r.Error != nil {
			if _, err := fmt.Fprintf(t.Writer, "❌ %s — ERROR: %v\n", r.URL, r.Error); err != nil {
				return err
			}
		} else {
			if _, err := fmt.Fprintf(t.Writer, "✅ %s — %d (%s)\n", r.URL, r.StatusCode, r.Duration); err != nil {
				return err
			}
		}
	}
	return nil
}

// --- JSON Exporter ---

type jsonResult struct {
	URL        string `json:"url"`
	StatusCode int    `json:"status_code,omitempty"`
	DurationMs int64  `json:"duration_ms"`
	Error      string `json:"error,omitempty"`
}

type JSONExporter struct {
	Writer io.Writer
}

func (j *JSONExporter) Export(results []checker.Result) error {
	var output []jsonResult
	for _, r := range results {
		jr := jsonResult{
			URL:        r.URL,
			StatusCode: r.StatusCode,
			DurationMs: r.Duration.Milliseconds(),
		}
		if r.Error != nil {
			jr.Error = r.Error.Error()
		}
		output = append(output, jr)
	}

	enc := json.NewEncoder(j.Writer)
	enc.SetIndent("", "  ")
	return enc.Encode(output)
}

// --- CSV Exporter ---

type CSVExporter struct {
	Writer io.Writer
}

func (c *CSVExporter) Export(results []checker.Result) error {
	if _, err := fmt.Fprintln(c.Writer, "url,status_code,duration_ms,error"); err != nil {
		return err
	}
	for _, r := range results {
		errStr := ""
		if r.Error != nil {
			errStr = r.Error.Error()
		}
		if _, err := fmt.Fprintf(c.Writer, "%s,%d,%d,%s\n",
			r.URL,
			r.StatusCode,
			r.Duration/time.Millisecond,
			errStr,
		); err != nil {
			return err
		}
	}
	return nil
}

// --- Logging Middleware ---

type LoggingExporter struct {
	Writer   io.Writer
	Exporter Exporter
}

func (l *LoggingExporter) Export(results []checker.Result) error {
	if _, err := fmt.Fprintf(l.Writer, "→ exporting %d results...\n", len(results)); err != nil {
		return err
	}

	start := time.Now()
	err := l.Exporter.Export(results)
	elapsed := time.Since(start)

	if err != nil {
		if _, werr := fmt.Fprintf(l.Writer, "→ export failed after %s: %v\n", elapsed, err); werr != nil {
			return werr
		}
	} else {
		if _, werr := fmt.Fprintf(l.Writer, "→ export completed in %s\n", elapsed); werr != nil {
			return werr
		}
	}

	return err
}

func New(format string, w io.Writer) Exporter {
	switch format {
	case "json":
		return &JSONExporter{Writer: w}
	case "csv":
		return &CSVExporter{Writer: w}
	default:
		return &TextExporter{Writer: w}
	}
}
