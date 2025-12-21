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

func installFd(ctx context.Context, logf func(string), requireChecks bool) error {
	if !isLinux() {
		return errUnsupported
	}

	_, localBin, _, _, err := userPaths()
	if err != nil {
		return err
	}

	client := githubapi.New()
	rel, err := client.GetLatestRelease("sharkdp", "fd")
	if err != nil {
		return err
	}

	// fd uses tags like v10.3.0 and asset names like:
	// fd-v10.3.0-x86_64-unknown-linux-gnu.tar.gz
	var archToken string
	switch runtime.GOARCH {
	case "amd64":
		archToken = "x86_64"
	case "arm64":
		archToken = "aarch64"
	default:
		return fmt.Errorf("unsupported arch for fd: %s", runtime.GOARCH)
	}

	tag := rel.TagName // includes v prefix
	assetName := fmt.Sprintf("fd-%s-%s-unknown-linux-gnu.tar.gz", tag, archToken)

	asset, ok := githubapi.FindAsset(rel, func(name string) bool { return name == assetName })
	if !ok {
		// fallback: match by suffix
		asset, ok = githubapi.FindAsset(rel, func(name string) bool {
			return strings.HasPrefix(name, "fd-") && strings.HasSuffix(name, "unknown-linux-gnu.tar.gz") && strings.Contains(name, archToken)
		})
		if !ok {
			return fmt.Errorf("fd release asset not found: %s", assetName)
		}
		assetName = asset.Name
	}

	logf(fmt.Sprintf("[INFO] Downloading fd %s (%s)", rel.TagName, assetName))
	dl, err := download.ToTempFile(ctx, asset.BrowserDownloadURL, "nvimwiz-fd-")
	if err != nil {
		return err
	}
	defer os.Remove(dl.Path)

	if requireChecks {
		// Try to find a checksum asset first.
		sumAsset, ok := githubapi.FindAsset(rel, func(name string) bool {
			l := strings.ToLower(name)
			return strings.Contains(l, "sha256") || strings.Contains(l, "checksum")
		})
		if !ok {
			// As a fallback, parse the release body looking for "sha256:" lines.
			// If we can't find any, fail.
			if !bodyHasSha256For(rel.Body, assetName, dl.SHA256Hex) {
				return fmt.Errorf("checksum required, but no checksums found for fd release")
			}
			logf("[OK] Verified fd checksum (from release notes)")
		} else {
			txt, err := download.ReadText(ctx, sumAsset.BrowserDownloadURL)
			if err != nil {
				return err
			}
			if !checksumTextMatches(txt, assetName, dl.SHA256Hex) && !bodyHasSha256For(txt, assetName, dl.SHA256Hex) {
				return fmt.Errorf("fd checksum mismatch for %s", assetName)
			}
			logf("[OK] Verified fd checksum")
		}
	} else {
		logf("[WARN] fd checksum verification skipped (not required)")
	}

	tmpDir, err := os.MkdirTemp("", "nvimwiz-fd-extract-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)

	if err := extract.TarGzToDir(dl.Path, tmpDir); err != nil {
		return err
	}

	// Find fd binary (often under fd-<tag>-<arch>/fd)
	bin, err := findAnyBinary(tmpDir, "fd")
	if err != nil {
		return err
	}

	dst := filepath.Join(localBin, "fd")
	if err := copyFile(bin, dst, 0o755); err != nil {
		return err
	}
	logf(fmt.Sprintf("[OK] Installed fd -> %s", dst))
	return nil
}

func findAnyBinary(root, name string) (string, error) {
	var found string
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if filepath.Base(path) == name {
			found = path
			return fs.SkipAll
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	if found == "" {
		return "", fmt.Errorf("could not locate %s in extracted archive", name)
	}
	return found, nil
}

func bodyHasSha256For(body, assetName, expectedHex string) bool {
	// Very loose parsing: look for "<assetName>" and a nearby "sha256:" hash.
	// This matches common GitHub release note formats.
	bodyLower := strings.ToLower(body)
	assetLower := strings.ToLower(assetName)
	idx := strings.Index(bodyLower, assetLower)
	if idx < 0 {
		// maybe filename isn't present; just try to match hash only.
		return strings.Contains(bodyLower, "sha256:"+strings.ToLower(expectedHex))
	}
	windowStart := idx
	if windowStart > 0 {
		windowStart -= min(windowStart, 500)
	}
	windowEnd := min(len(bodyLower), idx+len(assetLower)+500)
	window := bodyLower[windowStart:windowEnd]
	return strings.Contains(window, "sha256:"+strings.ToLower(expectedHex))
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
