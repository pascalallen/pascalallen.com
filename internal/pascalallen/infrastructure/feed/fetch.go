package feed

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	userAgent       = "pascalallen-site-feed/1"
	requestTimeout  = 5 * time.Second
	maxResponseSize = 1 << 20
)

var ErrResponseTooLarge = errors.New("upstream response exceeds the size cap")

func NewHTTPClient() *http.Client {
	return &http.Client{Timeout: requestTimeout}
}

func fetchJSON(ctx context.Context, client *http.Client, url string, bearer string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}

	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "application/json")
	if bearer != "" {
		req.Header.Set("Authorization", "Bearer "+bearer)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("request upstream: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("upstream returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseSize+1))
	if err != nil {
		return fmt.Errorf("read upstream body: %w", err)
	}
	if len(body) > maxResponseSize {
		return ErrResponseTooLarge
	}

	if err := json.Unmarshal(body, out); err != nil {
		return fmt.Errorf("decode upstream body: %w", err)
	}

	return nil
}
