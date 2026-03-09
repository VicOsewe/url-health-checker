package exporter_test

import (
	"bytes"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/VicOsewe/url-health-checker/internal/checker"
	"github.com/VicOsewe/url-health-checker/internal/exporter"
)

func successResult(url string, code int, duration time.Duration) checker.Result {
	return checker.Result{URL: url, StatusCode: code, Duration: duration}
}

func errorResult(url string, err error) checker.Result {
	return checker.Result{URL: url, Error: err}
}

// failWriter is an io.Writer that always fails
type failWriter struct{}

func (f *failWriter) Write(_ []byte) (int, error) {
	return 0, errors.New("write failed")
}

func TestTextExporter(t *testing.T) {
	tests := []struct {
		name          string
		results       []checker.Result
		wantContains  []string
		wantAbsent    []string
		wantLineCount int // 0 means skip line count check
	}{
		{
			name: "successful result shows checkmark and status code",
			results: []checker.Result{
				successResult("https://example.com", 200, 100*time.Millisecond),
			},
			wantContains: []string{"✅", "https://example.com", "200"},
		},
		{
			name: "error result shows cross and error message",
			results: []checker.Result{
				errorResult("https://broken.com", errors.New("connection refused")),
			},
			wantContains: []string{"❌", "connection refused"},
			wantAbsent:   []string{"✅"},
		},
		{
			name: "multiple results produce one line each",
			results: []checker.Result{
				successResult("https://a.com", 200, 50*time.Millisecond),
				errorResult("https://b.com", errors.New("timeout")),
				successResult("https://c.com", 404, 80*time.Millisecond),
			},
			wantLineCount: 3,
		},
		{
			name:    "empty results produces no output",
			results: []checker.Result{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			exp := &exporter.TextExporter{Writer: &buf}

			if err := exp.Export(tt.results); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			out := buf.String()

			for _, want := range tt.wantContains {
				if !strings.Contains(out, want) {
					t.Errorf("expected %q in output, got: %q", want, out)
				}
			}
			for _, absent := range tt.wantAbsent {
				if strings.Contains(out, absent) {
					t.Errorf("did not expect %q in output, got: %q", absent, out)
				}
			}
			if tt.wantLineCount > 0 {
				lines := strings.Split(strings.TrimSpace(out), "\n")
				if len(lines) != tt.wantLineCount {
					t.Errorf("expected %d lines, got %d:\n%s", tt.wantLineCount, len(lines), out)
				}
			}
			if len(tt.results) == 0 && out != "" {
				t.Errorf("expected empty output, got: %q", out)
			}
		})
	}
}

func TestCSVExporter(t *testing.T) {
	tests := []struct {
		name         string
		results      []checker.Result
		wantContains []string
		wantHeader   bool
		wantRowCount int // data rows only, excluding header
	}{
		{
			name:         "always writes header even with no results",
			results:      []checker.Result{},
			wantHeader:   true,
			wantRowCount: 0,
		},
		{
			name: "successful result has correct columns",
			results: []checker.Result{
				successResult("https://example.com", 200, 150*time.Millisecond),
			},
			wantHeader:   true,
			wantRowCount: 1,
			wantContains: []string{"https://example.com", "200"},
		},
		{
			name: "error result includes error message in last column",
			results: []checker.Result{
				errorResult("https://broken.com", errors.New("no route to host")),
			},
			wantHeader:   true,
			wantRowCount: 1,
			wantContains: []string{"no route to host"},
		},
		{
			name: "multiple results produce correct row count",
			results: []checker.Result{
				successResult("https://a.com", 200, 50*time.Millisecond),
				successResult("https://b.com", 404, 80*time.Millisecond),
				errorResult("https://c.com", errors.New("timeout")),
			},
			wantHeader:   true,
			wantRowCount: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			exp := &exporter.CSVExporter{Writer: &buf}

			if err := exp.Export(tt.results); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			out := buf.String()
			lines := strings.Split(strings.TrimSpace(out), "\n")

			if tt.wantHeader {
				if !strings.HasPrefix(lines[0], "url,status_code,duration_ms,error") {
					t.Errorf("expected CSV header in first line, got: %q", lines[0])
				}
			}

			// lines[0] is header, rest are data
			dataRows := lines[1:]
			if tt.wantRowCount == 0 && len(dataRows) == 1 && dataRows[0] == "" {
				dataRows = []string{} // TrimSpace artifact on empty body
			}
			if len(dataRows) != tt.wantRowCount {
				t.Errorf("expected %d data rows, got %d", tt.wantRowCount, len(dataRows))
			}

			for _, want := range tt.wantContains {
				if !strings.Contains(out, want) {
					t.Errorf("expected %q in output, got: %q", want, out)
				}
			}
		})
	}
}

func TestJSONExporter(t *testing.T) {
	tests := []struct {
		name         string
		results      []checker.Result
		wantContains []string
		wantAbsent   []string
		wantPrefix   string
	}{
		{
			name: "successful result produces valid JSON array with expected fields",
			results: []checker.Result{
				successResult("https://example.com", 200, 100*time.Millisecond),
			},
			wantPrefix:   "[",
			wantContains: []string{`"url"`, `"status_code"`, `"duration_ms"`},
		},
		{
			name: "error result includes error field and omits zero status_code",
			results: []checker.Result{
				errorResult("https://broken.com", errors.New("refused")),
			},
			wantContains: []string{`"error"`},
			wantAbsent:   []string{`"status_code": 0`},
		},
		{
			name:    "empty results does not error",
			results: []checker.Result{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			exp := &exporter.JSONExporter{Writer: &buf}

			if err := exp.Export(tt.results); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			out := buf.String()

			if tt.wantPrefix != "" && !strings.HasPrefix(strings.TrimSpace(out), tt.wantPrefix) {
				t.Errorf("expected output to start with %q, got: %q", tt.wantPrefix, out)
			}
			for _, want := range tt.wantContains {
				if !strings.Contains(out, want) {
					t.Errorf("expected %q in output, got: %q", want, out)
				}
			}
			for _, absent := range tt.wantAbsent {
				if strings.Contains(out, absent) {
					t.Errorf("did not expect %q in output, got: %q", absent, out)
				}
			}
		})
	}
}

func TestNew(t *testing.T) {
	tests := []struct {
		format  string
		wantNil bool
	}{
		{format: "text"},
		{format: "json"},
		{format: "csv"},
		{format: "unknown"}, // defaults to text
		{format: ""},        // defaults to text
	}

	for _, tt := range tests {
		t.Run(tt.format, func(t *testing.T) {
			exp := exporter.New(tt.format, &bytes.Buffer{})
			if exp == nil {
				t.Errorf("New(%q) returned nil", tt.format)
			}
		})
	}
}

func TestLoggingExporter(t *testing.T) {
	tests := []struct {
		name         string
		results      []checker.Result
		innerWriter  func() exporter.Exporter
		wantLogLines []string
		wantErr      bool
	}{
		{
			name: "writes start and completion lines to log",
			results: []checker.Result{
				successResult("https://example.com", 200, 10*time.Millisecond),
			},
			innerWriter: func() exporter.Exporter {
				return &exporter.TextExporter{Writer: &bytes.Buffer{}}
			},
			wantLogLines: []string{"exporting 1 results", "export completed"},
		},
		{
			name: "logs correct count for multiple results",
			results: []checker.Result{
				successResult("https://a.com", 200, 10*time.Millisecond),
				successResult("https://b.com", 200, 10*time.Millisecond),
			},
			innerWriter: func() exporter.Exporter {
				return &exporter.TextExporter{Writer: &bytes.Buffer{}}
			},
			wantLogLines: []string{"exporting 2 results"},
		},
		{
			name: "propagates error from inner exporter",
			results: []checker.Result{
				successResult("https://example.com", 200, 10*time.Millisecond),
			},
			innerWriter: func() exporter.Exporter {
				return &exporter.TextExporter{Writer: &failWriter{}}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var logBuf bytes.Buffer
			logged := &exporter.LoggingExporter{
				Writer:   &logBuf,
				Exporter: tt.innerWriter(),
			}

			err := logged.Export(tt.results)

			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			log := logBuf.String()
			for _, want := range tt.wantLogLines {
				if !strings.Contains(log, want) {
					t.Errorf("expected %q in log, got: %q", want, log)
				}
			}
		})
	}
}
