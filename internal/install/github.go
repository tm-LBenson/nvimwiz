package install

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
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
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return ghRelease{}, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "nvimwiz")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return ghRelease{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return ghRelease{}, errors.New("github api error")
	}
	var rel ghRelease
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return ghRelease{}, err
	}
	return rel, nil
}

func findAsset(rel ghRelease, fn func(a ghAsset) bool) (ghAsset, bool) {
	for _, a := range rel.Assets {
		if fn(a) {
			return a, true
		}
	}
	return ghAsset{}, false
}
