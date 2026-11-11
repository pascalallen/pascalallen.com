package slack

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
	"unicode"

	"github.com/pascalallen/pascalallen.com/internal/pascalallen/domain/visit"
)

const (
	DefaultTimeout = 3 * time.Second
	userAgent      = "pascalallen-site-visit/1"
)

type Notifier struct {
	webhookURL string
	enabled    bool
	client     *http.Client
}

func NewNotifier(webhookURL string, appEnv string, client *http.Client) *Notifier {
	if client == nil {
		client = &http.Client{Timeout: DefaultTimeout}
	}

	return &Notifier{
		webhookURL: webhookURL,
		enabled:    webhookURL != "" && appEnv == "production",
		client:     client,
	}
}

func NewNotifierWithTimeout(webhookURL string, appEnv string, timeout time.Duration) *Notifier {
	return NewNotifier(webhookURL, appEnv, &http.Client{Timeout: timeout})
}

func plain(s string) string {
	cleaned := strings.Map(func(r rune) rune {
		switch {
		case unicode.IsControl(r), unicode.Is(unicode.Cf, r), unicode.Is(unicode.Zl, r), unicode.Is(unicode.Zp, r):
			return -1
		case r == '`':
			return '\''
		}

		return r
	}, s)

	return strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;").Replace(cleaned)
}

func (n *Notifier) Notify(ctx context.Context, v visit.Visit) error {
	if !n.enabled {
		return nil
	}

	payload, err := json.Marshal(map[string]string{
		"text": fmt.Sprintf("pascalallen.com visit: lang=`%s` ua=`%s`", plain(v.Language), plain(v.UserAgent)),
	})
	if err != nil {
		return fmt.Errorf("encode slack payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, n.webhookURL, bytes.NewReader(payload))
	if err != nil {
		return errors.New("build slack request")
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", userAgent)

	resp, err := n.client.Do(req)
	if err != nil {
		var urlErr *url.Error
		if errors.As(err, &urlErr) {
			return fmt.Errorf("post to slack: %w", urlErr.Err)
		}

		return errors.New("post to slack")
	}
	defer func() { _ = resp.Body.Close() }()
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 4096))

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("slack returned status %d", resp.StatusCode)
	}

	return nil
}
