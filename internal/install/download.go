package install

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

func downloadFile(ctx context.Context, url, dst string) error {
	const maxAttempts = 4
	var lastErr error

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		if attempt > 1 {
			// Exponential backoff: 1s, 2s, 4s...
			backoff := time.Second * time.Duration(1<<(attempt-2))
			select {
			case <-time.After(backoff):
			case <-ctx.Done():
				return ctx.Err()
			}
		}

		tmp := dst + ".part"
		_ = os.Remove(tmp)

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return err
		}
		req.Header.Set("User-Agent", "nvimwiz")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			lastErr = err
			continue
		}
		func() {
			defer resp.Body.Close()
			if resp.StatusCode < 200 || resp.StatusCode >= 300 {
				b, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
				lastErr = &httpError{StatusCode: resp.StatusCode, Body: string(b)}
				return
			}

			f, err := os.Create(tmp)
			if err != nil {
				lastErr = err
				return
			}
			defer f.Close()

			if _, err := io.Copy(f, resp.Body); err != nil {
				lastErr = err
				return
			}
			if err := f.Sync(); err != nil {
				lastErr = err
				return
			}
			if err := f.Close(); err != nil {
				lastErr = err
				return
			}

			_ = os.Remove(dst)
			if err := os.Rename(tmp, dst); err != nil {
				lastErr = err
				return
			}
			lastErr = nil
		}()

		if lastErr == nil {
			return nil
		}

		// Only retry on transient HTTP errors (5xx) and rate limiting (429).
		if he, ok := lastErr.(*httpError); ok {
			if he.StatusCode < 500 && he.StatusCode != 429 {
				return lastErr
			}
		}
	}

	if lastErr == nil {
		lastErr = fmt.Errorf("download failed")
	}
	return lastErr
}

type httpError struct {
	StatusCode int
	Body       string
}

func (e *httpError) Error() string {
	msg := strings.TrimSpace(e.Body)
	if msg == "" {
		return fmt.Sprintf("http error (%d)", e.StatusCode)
	}
	return fmt.Sprintf("http error (%d): %s", e.StatusCode, msg)
}

func sha256File(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
