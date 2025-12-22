package install

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
)

func verifyAssetIfPossible(ctx context.Context, policy string, rel ghRelease, asset ghAsset, filePath string, log func(string)) error {
	p := strings.ToLower(strings.TrimSpace(policy))
	if p == "off" {
		return nil
	}

	cands := []string{
		asset.Name + ".sha256",
		asset.Name + ".sha256sum",
		asset.Name + ".sha256.txt",
	}

	cs, ok := findAsset(rel, func(a ghAsset) bool {
		for _, c := range cands {
			if a.Name == c {
				return true
			}
		}
		return false
	})

	if !ok {
		if p == "require" {
			return errors.New("checksum not available for " + asset.Name)
		}
		if log != nil {
			log("Checksum not available for " + asset.Name + ", skipping verification")
		}
		return nil
	}

	csPath := filepath.Join(filepath.Dir(filePath), cs.Name)
	if log != nil {
		log("Downloading checksum " + cs.Name)
	}
	if err := downloadFile(ctx, cs.BrowserDownloadURL, csPath); err != nil {
		if p == "require" {
			return err
		}
		if log != nil {
			log("Checksum download failed, skipping verification")
		}
		return nil
	}

	m, err := parseChecksumFile(csPath)
	if err != nil {
		if p == "require" {
			return err
		}
		if log != nil {
			log("Checksum parse failed, skipping verification")
		}
		return nil
	}
	expected := m[asset.Name]
	if expected == "" {
		if len(m) == 1 {
			for _, v := range m {
				expected = v
			}
		}
	}
	if expected == "" {
		if p == "require" {
			return errors.New("checksum mismatch data for " + asset.Name)
		}
		if log != nil {
			log("Checksum entry missing for " + asset.Name + ", skipping verification")
		}
		return nil
	}

	got, err := sha256File(filePath)
	if err != nil {
		return err
	}
	if strings.EqualFold(strings.TrimSpace(got), strings.TrimSpace(expected)) {
		if log != nil {
			log("Checksum verified for " + asset.Name)
		}
		return nil
	}
	return errors.New("checksum verification failed for " + asset.Name)
}
