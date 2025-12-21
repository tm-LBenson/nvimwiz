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

func installNeovim(ctx context.Context, logf func(string), requireChecks bool) error {
	if !isLinux() && !isDarwin() {
		return errUnsupported
	}

	home, localBin, localNvimRoot, _, err := userPaths()
	if err != nil {
		return err
	}
	_ = home

	client := githubapi.New()

	// Prefer "stable" tag to match neovim's stable channel.
	rel, err := client.GetReleaseByTag("neovim", "neovim", "stable")
	if err != nil {
		// fallback to latest release if "stable" tag not available
		logf("[WARN] Could not fetch neovim stable tag; falling back to latest")
		rel, err = client.GetLatestRelease("neovim", "neovim")
		if err != nil {
			return err
		}
	}

	var assetName string
	switch runtime.GOOS {
	case "linux":
		switch runtime.GOARCH {
		case "amd64":
			assetName = "nvim-linux-x86_64.tar.gz"
		case "arm64":
			assetName = "nvim-linux-arm64.tar.gz"
		default:
			return fmt.Errorf("unsupported arch for neovim: %s", runtime.GOARCH)
		}
	case "darwin":
		switch runtime.GOARCH {
		case "amd64":
			assetName = "nvim-macos-x86_64.tar.gz"
		case "arm64":
			assetName = "nvim-macos-arm64.tar.gz"
		default:
			return fmt.Errorf("unsupported arch for neovim: %s", runtime.GOARCH)
		}
	default:
		return errUnsupported
	}

	asset, ok := githubapi.FindAsset(rel, func(name string) bool { return name == assetName })
	if !ok {
		return fmt.Errorf("neovim release asset not found: %s", assetName)
	}

	logf(fmt.Sprintf("[INFO] Downloading Neovim %s (%s)", strings.TrimSpace(rel.TagName), assetName))
	dl, err := download.ToTempFile(ctx, asset.BrowserDownloadURL, "nvimwiz-nvim-")
	if err != nil {
		return err
	}
	defer os.Remove(dl.Path)

	// Neovim checksums are not always published as an asset. We only enforce if required.
	// (A future enhancement could scrape the release body for sha256 lines.)
	if requireChecks {
		// Best-effort attempt: look for a checksums file among assets.
		sumAsset, ok := githubapi.FindAsset(rel, func(name string) bool {
			l := strings.ToLower(name)
			return strings.Contains(l, "sha256") || strings.Contains(l, "checksum")
		})
		if !ok {
			return fmt.Errorf("checksum required, but no checksum asset found in neovim release")
		}
		txt, err := download.ReadText(ctx, sumAsset.BrowserDownloadURL)
		if err != nil {
			return err
		}
		if !checksumTextMatches(txt, assetName, dl.SHA256Hex) {
			return fmt.Errorf("neovim checksum mismatch for %s", assetName)
		}
		logf("[OK] Verified Neovim checksum")
	} else {
		logf("[WARN] Neovim checksum verification skipped (not required)")
	}

	// Extract to ~/.local/nvim/<tag>/
	tag := strings.TrimSpace(rel.TagName)
	if tag == "" {
		tag = "stable"
	}
	installDir := filepath.Join(localNvimRoot, tag)
	if err := os.MkdirAll(installDir, 0o755); err != nil {
		return err
	}

	// Extract into installDir
	if err := extract.TarGzToDir(dl.Path, installDir); err != nil {
		return err
	}

	// Find the nvim binary under installDir/**/bin/nvim
	bin, err := findFile(installDir, filepath.Join("bin", "nvim"))
	if err != nil {
		return err
	}

	// Symlink ~/.local/bin/nvim -> found bin
	targetLink := filepath.Join(localBin, "nvim")
	_ = os.Remove(targetLink)
	if err := os.Symlink(bin, targetLink); err != nil {
		// fallback: copy if symlink fails
		logf("[WARN] Symlink failed; copying nvim binary instead")
		return copyFile(bin, targetLink, 0o755)
	}

	logf(fmt.Sprintf("[OK] nvim -> %s", targetLink))
	return nil
}

func checksumTextMatches(text, filename, expectedHex string) bool {
	// Look for lines like:
	// <sha>  <file>
	// or "<sha> <file>"
	lines := strings.Split(text, "\n")
	for _, ln := range lines {
		ln = strings.TrimSpace(ln)
		if ln == "" {
			continue
		}
		parts := strings.Fields(ln)
		if len(parts) < 2 {
			continue
		}
		sha := strings.ToLower(parts[0])
		file := parts[len(parts)-1]
		if file == filename && sha == strings.ToLower(expectedHex) {
			return true
		}
	}
	return false
}

// findFile searches for a suffix path under root and returns the first match.
func findFile(root, suffix string) (string, error) {
	var found string
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		// normalize to forward slashes for suffix match
		p := filepath.ToSlash(path)
		s := filepath.ToSlash(filepath.Join(root, suffix))
		if strings.HasSuffix(p, filepath.ToSlash(suffix)) && strings.HasPrefix(p, filepath.ToSlash(root)) {
			found = path
			return fs.SkipAll
		}
		_ = s
		return nil
	})
	if err != nil {
		return "", err
	}
	if found == "" {
		return "", fmt.Errorf("could not locate %s under %s", suffix, root)
	}
	return found, nil
}
