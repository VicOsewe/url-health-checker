package checker

import (
	"context"
	"net/http"
	"time"
)

type Result struct {
	URL        string
	StatusCode int
	Duration   time.Duration
	Error      error
}

func Check(ctx context.Context, url string, timeout time.Duration) Result {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	start := time.Now()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return Result{URL: url, Error: err, Duration: time.Since(start)}
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return Result{URL: url, Error: err, Duration: time.Since(start)}
	}
	defer resp.Body.Close()

	return Result{
		URL:        url,
		StatusCode: resp.StatusCode,
		Duration:   time.Since(start),
	}
}
