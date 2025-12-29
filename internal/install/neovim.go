package install

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"nvimwiz/internal/env"
)

func InstallNeovim(ctx context.Context, verify string, log func(string)) (string, error) {
	rel, err := fetchLatestRelease(ctx, "neovim", "neovim")
	if err != nil {
		return "", err
	}

	latest := normalizeVersion(rel.TagName)
	if cur, path, ok := installedCommandVersion(ctx, "nvim", "--version"); ok && cur == latest {
		if log != nil {
			log("Neovim already up to date (" + rel.TagName + "), skipping download")
		}
		return path, nil
	}
	asset, ok := findAsset(rel, func(a ghAsset) bool {
		name := a.Name
		if !strings.HasSuffix(name, ".tar.gz") {
			return false
		}
		return matchNeovimAsset(name)
	})
	if !ok {
		return "", errors.New("neovim asset not found")
	}

	workRoot, err := localNvimRoot()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(workRoot, 0o755); err != nil {
		return "", err
	}

	tmpDir, err := os.MkdirTemp("", "nvimwiz-nvim-*")
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

	extractTmp := filepath.Join(workRoot, ".tmp-"+time.Now().Format("20060102-150405"))
	if err := os.MkdirAll(extractTmp, 0o755); err != nil {
		return "", err
	}
	top, err := extractTarGz(tarPath, extractTmp)
	if err != nil {
		_ = os.RemoveAll(extractTmp)
		return "", err
	}
	srcDir := filepath.Join(extractTmp, top)
	if _, err := os.Stat(srcDir); err != nil {
		_ = os.RemoveAll(extractTmp)
		return "", errors.New("extraction failed")
	}

	targetDir := filepath.Join(workRoot, rel.TagName)
	_ = os.RemoveAll(targetDir)
	if err := os.Rename(srcDir, targetDir); err != nil {
		_ = os.RemoveAll(extractTmp)
		return "", err
	}
	_ = os.RemoveAll(extractTmp)

	lb, err := env.LocalBin()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(lb, 0o755); err != nil {
		return "", err
	}
	bin := filepath.Join(targetDir, "bin", "nvim")
	link := filepath.Join(lb, "nvim")
	if runtime.GOOS == "windows" {
		bin = filepath.Join(targetDir, "bin", "nvim.exe")
		link = filepath.Join(lb, "nvim.exe")
	}
	if err := replaceSymlink(link, bin); err != nil {
		return "", err
	}
	if log != nil {
		log("Installed nvim to " + bin)
	}
	return link, nil
}

func localNvimRoot() (string, error) {
	h, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(h, ".local", "nvim"), nil
}

func matchNeovimAsset(name string) bool {
	goos := runtime.GOOS
	arch := runtime.GOARCH
	if goos == "linux" && arch == "amd64" {
		return name == "nvim-linux-x86_64.tar.gz"
	}
	if goos == "linux" && arch == "arm64" {
		return name == "nvim-linux-arm64.tar.gz"
	}
	if goos == "darwin" && arch == "amd64" {
		return name == "nvim-macos-x86_64.tar.gz"
	}
	if goos == "darwin" && arch == "arm64" {
		return name == "nvim-macos-arm64.tar.gz"
	}
	return false
}
