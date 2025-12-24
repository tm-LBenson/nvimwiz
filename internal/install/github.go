package install

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

type ghAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

type ghRelease struct {
	TagName string    `json:"tag_name"`
	Body    string    `json:"body"`
	Assets  []ghAsset `json:"assets"`
}

func fetchLatestRelease(ctx context.Context, owner, repo string) (ghRelease, error) {
	url := "https://api.github.com/repos/" + owner + "/" + repo + "/releases/latest"

	const maxAttempts = 4
	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		if attempt > 1 {
			backoff := time.Duration(1<<uint(attempt-2)) * time.Second
			select {
			case <-time.After(backoff):
			case <-ctx.Done():
				return ghRelease{}, ctx.Err()
			}
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return ghRelease{}, err
		}
		req.Header.Set("Accept", "application/vnd.github+json")
		req.Header.Set("User-Agent", "nvimwiz")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			lastErr = err
			continue
		}

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			// Retry transient errors.
			retryable := resp.StatusCode >= 500 || resp.StatusCode == 429
			_ = resp.Body.Close()
			lastErr = fmt.Errorf("github api error: status %d", resp.StatusCode)
			if retryable {
				continue
			}
			return ghRelease{}, lastErr
		}

		var rel ghRelease
		err = json.NewDecoder(resp.Body).Decode(&rel)
		_ = resp.Body.Close()
		if err != nil {
			lastErr = err
			continue
		}
		return rel, nil
	}
	if lastErr == nil {
		lastErr = errors.New("github api error")
	}
	return ghRelease{}, lastErr
}

func findAsset(rel ghRelease, fn func(a ghAsset) bool) (ghAsset, bool) {
	for _, a := range rel.Assets {
		if fn(a) {
			return a, true
		}
	}
	return ghAsset{}, false
}
