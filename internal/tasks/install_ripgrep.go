package tasks

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"nvimwiz/internal/download"
	"nvimwiz/internal/extract"
	"nvimwiz/internal/githubapi"
)

func installRipgrep(ctx context.Context, logf func(string), requireChecks bool) error {
	if !isLinux() {
		return errUnsupported
	}

	_, localBin, _, _, err := userPaths()
	if err != nil {
		return err
	}

	client := githubapi.New()
	rel, err := client.GetLatestRelease("BurntSushi", "ripgrep")
	if err != nil {
		return err
	}

	var assetName string
	switch runtime.GOARCH {
	case "amd64":
		assetName = fmt.Sprintf("ripgrep-%s-x86_64-unknown-linux-musl.tar.gz", strings.TrimPrefix(rel.TagName, "v"))
	case "arm64":
		assetName = fmt.Sprintf("ripgrep-%s-aarch64-unknown-linux-musl.tar.gz", strings.TrimPrefix(rel.TagName, "v"))
	default:
		return fmt.Errorf("unsupported arch for ripgrep: %s", runtime.GOARCH)
	}

	asset, ok := githubapi.FindAsset(rel, func(name string) bool { return name == assetName })
	if !ok {
		// fallback: pick by suffix
		asset, ok = githubapi.FindAsset(rel, func(name string) bool {
			return strings.HasSuffix(name, "unknown-linux-musl.tar.gz") && strings.Contains(name, runtimeArchToken(runtime.GOARCH))
		})
		if !ok {
			return fmt.Errorf("ripgrep release asset not found: %s", assetName)
		}
		assetName = asset.Name
	}

	logf(fmt.Sprintf("[INFO] Downloading ripgrep %s (%s)", rel.TagName, assetName))
	dl, err := download.ToTempFile(ctx, asset.BrowserDownloadURL, "nvimwiz-rg-")
	if err != nil {
		return err
	}
	defer os.Remove(dl.Path)

	if requireChecks {
		shaAsset, ok := githubapi.FindAsset(rel, func(name string) bool {
			return name == assetName+".sha256" || name == assetName+".sha256sum" || strings.HasSuffix(name, ".sha256")
		})
		if !ok {
			return fmt.Errorf("checksum required, but no checksum asset found for ripgrep")
		}
		txt, err := download.ReadText(ctx, shaAsset.BrowserDownloadURL)
		if err != nil {
			return err
		}
		expected := parseFirstHex64(txt)
		if expected == "" {
			return fmt.Errorf("could not parse checksum file for ripgrep")
		}
		if strings.ToLower(expected) != strings.ToLower(dl.SHA256Hex) {
			return fmt.Errorf("ripgrep checksum mismatch")
		}
		logf("[OK] Verified ripgrep checksum")
	} else {
		logf("[WARN] ripgrep checksum verification skipped (not required)")
	}

	tmpDir, err := os.MkdirTemp("", "nvimwiz-rg-extract-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)

	if err := extract.TarGzToDir(dl.Path, tmpDir); err != nil {
		return err
	}

	// Find rg binary
	bin, err := findFile(tmpDir, filepath.Join("rg"))
	if err != nil {
		// try typical path: ripgrep-*/rg
		bin, err = findAnyRg(tmpDir)
		if err != nil {
			return err
		}
	}

	dst := filepath.Join(localBin, "rg")
	if err := copyFile(bin, dst, 0o755); err != nil {
		return err
	}
	logf(fmt.Sprintf("[OK] Installed rg -> %s", dst))
	return nil
}

func runtimeArchToken(goarch string) string {
	switch goarch {
	case "amd64":
		return "x86_64"
	case "arm64":
		return "aarch64"
	default:
		return goarch
	}
}

func parseFirstHex64(s string) string {
	// Find first 64-hex chunk at the start of a line.
	for _, ln := range strings.Split(s, "\n") {
		ln = strings.TrimSpace(ln)
		if len(ln) < 64 {
			continue
		}
		hex := ln[:64]
		if isHex64(hex) {
			return hex
		}
	}
	return ""
}

func isHex64(s string) bool {
	if len(s) != 64 {
		return false
	}
	for _, r := range s {
		if (r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F') {
			continue
		}
		return false
	}
	return true
}

func findAnyRg(root string) (string, error) {
	var found string
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if filepath.Base(path) == "rg" {
			found = path
			return fs.SkipAll
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	if found == "" {
		return "", fmt.Errorf("could not locate rg in extracted archive")
	}
	return found, nil
}
