package install

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"nvimwiz/internal/env"
)

func InstallRipgrep(ctx context.Context, verify string, log func(string)) (string, error) {
	rel, err := fetchLatestRelease(ctx, "BurntSushi", "ripgrep")
	if err != nil {
		return "", err
	}

	latest := normalizeVersion(rel.TagName)
	if cur, path, ok := installedCommandVersion(ctx, "rg", "--version"); ok && cur == latest {
		if log != nil {
			log("ripgrep already up-to-date (" + rel.TagName + "), skipping")
		}
		return path, nil
	}
	asset, ok := findAsset(rel, func(a ghAsset) bool {
		name := a.Name
		if !strings.HasPrefix(name, "ripgrep-") || !strings.HasSuffix(name, ".tar.gz") {
			return false
		}
		return matchRipgrepAsset(name)
	})
	if !ok {
		return "", errors.New("ripgrep asset not found")
	}

	tmpDir, err := os.MkdirTemp("", "nvimwiz-rg-*")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(tmpDir)

	tarPath := filepath.Join(tmpDir, asset.Name)
	if log != nil {
		log("Downloading " + asset.Name)
	}
	if err := downloadFile(ctx, asset.BrowserDownloadURL, tarPath); err != nil {
		return "", err
	}

	if err := verifyAssetIfPossible(ctx, verify, rel, asset, tarPath, log); err != nil {
		return "", err
	}

	extractDir := filepath.Join(tmpDir, "x")
	if err := os.MkdirAll(extractDir, 0o755); err != nil {
		return "", err
	}
	top, err := extractTarGz(tarPath, extractDir)
	if err != nil {
		return "", err
	}

	binName := "rg"
	if runtime.GOOS == "windows" {
		binName = "rg.exe"
	}
	binPath, err := findFile(filepath.Join(extractDir, top), binName)
	if err != nil {
		return "", err
	}

	lb, err := env.LocalBin()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(lb, 0o755); err != nil {
		return "", err
	}
	dst := filepath.Join(lb, binName)
	if err := copyFile(binPath, dst); err != nil {
		return "", err
	}
	if log != nil {
		log("Installed rg to " + dst)
	}
	return dst, nil
}

func matchRipgrepAsset(name string) bool {
	goos := runtime.GOOS
	arch := runtime.GOARCH
	cands := []string{}
	if goos == "linux" && arch == "amd64" {
		cands = []string{"x86_64-unknown-linux-musl"}
	}
	if goos == "linux" && arch == "arm64" {
		cands = []string{"aarch64-unknown-linux-musl", "arm64-unknown-linux-musl"}
	}
	if goos == "darwin" && arch == "amd64" {
		cands = []string{"x86_64-apple-darwin"}
	}
	if goos == "darwin" && arch == "arm64" {
		cands = []string{"aarch64-apple-darwin", "arm64-apple-darwin"}
	}
	if goos == "windows" && arch == "amd64" {
		cands = []string{"x86_64-pc-windows-msvc"}
	}
	for _, c := range cands {
		if strings.Contains(name, c) {
			return true
		}
	}
	return false
}
