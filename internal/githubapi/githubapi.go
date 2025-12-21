package githubapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Release struct {
	TagName string  `json:"tag_name"`
	Name    string  `json:"name"`
	Body    string  `json:"body"`
	Assets  []Asset `json:"assets"`
}

type Asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
	Size               int64  `json:"size"`
}

type Client struct {
	http *http.Client
}

func New() *Client {
	return &Client{
		http: &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *Client) GetLatestRelease(owner, repo string) (Release, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", owner, repo)
	return c.getRelease(url)
}

func (c *Client) GetReleaseByTag(owner, repo, tag string) (Release, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/tags/%s", owner, repo, tag)
	return c.getRelease(url)
}

func (c *Client) getRelease(url string) (Release, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return Release{}, err
	}
	// GitHub API is nicer with a user agent.
	req.Header.Set("User-Agent", "nvimwiz")
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := c.http.Do(req)
	if err != nil {
		return Release{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return Release{}, fmt.Errorf("GitHub API %s: %s", url, resp.Status)
	}

	var r Release
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return Release{}, err
	}
	return r, nil
}

func FindAsset(r Release, predicate func(name string) bool) (Asset, bool) {
	for _, a := range r.Assets {
		if predicate(a.Name) {
			return a, true
		}
	}
	return Asset{}, false
}
