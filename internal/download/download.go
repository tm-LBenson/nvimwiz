package download

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type Result struct {
	Path      string
	SHA256Hex string
	Bytes     int64
}

func ToTempFile(ctx context.Context, url string, prefix string) (Result, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return Result{}, err
	}
	req.Header.Set("User-Agent", "nvimwiz")

	client := &http.Client{Timeout: 10 * time.Minute}
	resp, err := client.Do(req)
	if err != nil {
		return Result{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return Result{}, fmt.Errorf("download %s: %s", url, resp.Status)
	}

	f, err := os.CreateTemp("", prefix)
	if err != nil {
		return Result{}, err
	}
	defer func() { _ = f.Close() }()

	h := sha256.New()
	n, err := io.Copy(io.MultiWriter(f, h), resp.Body)
	if err != nil {
		return Result{}, err
	}

	sum := hex.EncodeToString(h.Sum(nil))
	return Result{Path: f.Name(), SHA256Hex: sum, Bytes: n}, nil
}

func ReadText(ctx context.Context, url string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "nvimwiz")
	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("download %s: %s", url, resp.Status)
	}
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
