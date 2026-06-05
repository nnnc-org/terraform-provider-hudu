package client

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"golang.org/x/time/rate"
)

const (
	defaultTimeout = 30 * time.Second
	maxRetries     = 3
)

// Client is a thin wrapper around net/http for the Hudu REST API.
type Client struct {
	BaseURL    string
	APIKey     string
	UserAgent  string
	HTTPClient *http.Client
	limiter    *rate.Limiter
}

// Config carries everything needed to construct a Client.
type Config struct {
	Host      string // e.g. https://hudu.example.com — /api/v1 is appended if missing
	APIKey    string
	UserAgent string
	Insecure  bool
	Timeout   time.Duration
}

// New constructs a Client from cfg.
func New(cfg Config) (*Client, error) {
	if cfg.Host == "" {
		return nil, errors.New("hudu: host is required")
	}
	if cfg.APIKey == "" {
		return nil, errors.New("hudu: api_key is required")
	}

	base := strings.TrimRight(cfg.Host, "/")
	if !strings.HasSuffix(base, "/api/v1") {
		base = base + "/api/v1"
	}
	if _, err := url.Parse(base); err != nil {
		return nil, fmt.Errorf("hudu: invalid host %q: %w", cfg.Host, err)
	}

	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = defaultTimeout
	}

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: cfg.Insecure}, //nolint:gosec
	}

	return &Client{
		BaseURL:   base,
		APIKey:    cfg.APIKey,
		UserAgent: cfg.UserAgent,
		HTTPClient: &http.Client{
			Timeout:   timeout,
			Transport: transport,
		},
		// 300/min ~= 5/s, burst 10
		limiter: rate.NewLimiter(rate.Limit(5), 10),
	}, nil
}

// do performs an HTTP request and decodes the response into out (if non-nil).
// body, if non-nil, is JSON-encoded before sending. Returns ErrNotFound for 404.
func (c *Client) do(ctx context.Context, method, path string, query url.Values, body, out any) error {
	var reqBody io.Reader
	if body != nil {
		buf, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("hudu: marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(buf)
	}

	full := c.BaseURL + path
	if len(query) > 0 {
		full += "?" + query.Encode()
	}

	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if err := c.limiter.Wait(ctx); err != nil {
			return fmt.Errorf("hudu: rate limiter: %w", err)
		}

		// Rewind body for retries.
		if body != nil && attempt > 0 {
			reqBody = bytes.NewReader(mustMarshal(body))
		}

		req, err := http.NewRequestWithContext(ctx, method, full, reqBody)
		if err != nil {
			return fmt.Errorf("hudu: new request: %w", err)
		}
		req.Header.Set("x-api-key", c.APIKey)
		req.Header.Set("Accept", "application/json")
		if body != nil {
			req.Header.Set("Content-Type", "application/json")
		}
		if c.UserAgent != "" {
			req.Header.Set("User-Agent", c.UserAgent)
		}

		resp, err := c.HTTPClient.Do(req)
		if err != nil {
			lastErr = err
			if shouldRetryErr(err) && attempt < maxRetries {
				backoff(attempt)
				continue
			}
			return fmt.Errorf("hudu: %s %s: %w", method, path, err)
		}

		respBody, err := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if err != nil {
			return fmt.Errorf("hudu: read response: %w", err)
		}

		switch {
		case resp.StatusCode == 429:
			if attempt < maxRetries {
				retryAfter := parseRetryAfter(resp.Header.Get("Retry-After"))
				time.Sleep(retryAfter)
				continue
			}
		case resp.StatusCode >= 500 && resp.StatusCode < 600:
			if attempt < maxRetries {
				backoff(attempt)
				continue
			}
		}

		if resp.StatusCode == http.StatusNotFound {
			return &APIError{StatusCode: 404, Method: method, Path: path, Body: string(respBody)}
		}

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return &APIError{StatusCode: resp.StatusCode, Method: method, Path: path, Body: string(respBody)}
		}

		if out != nil && len(respBody) > 0 {
			if err := json.Unmarshal(respBody, out); err != nil {
				return fmt.Errorf("hudu: decode response: %w (body: %s)", err, truncate(respBody, 512))
			}
		}
		return nil
	}

	return lastErr
}

func mustMarshal(v any) []byte {
	b, _ := json.Marshal(v)
	return b
}

func shouldRetryErr(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "EOF") ||
		strings.Contains(msg, "connection reset") ||
		strings.Contains(msg, "timeout")
}

func backoff(attempt int) {
	// 250ms, 500ms, 1s
	d := time.Duration(250*(1<<attempt)) * time.Millisecond
	time.Sleep(d)
}

func parseRetryAfter(h string) time.Duration {
	if h == "" {
		return 2 * time.Second
	}
	if secs, err := strconv.Atoi(h); err == nil {
		return time.Duration(secs) * time.Second
	}
	if t, err := http.ParseTime(h); err == nil {
		if d := time.Until(t); d > 0 {
			return d
		}
	}
	return 2 * time.Second
}

func truncate(b []byte, n int) string {
	if len(b) <= n {
		return string(b)
	}
	return string(b[:n]) + "...(truncated)"
}
