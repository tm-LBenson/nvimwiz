package assets

import (
	"errors"
	"os"
	"path/filepath"
)

func FindNvimAssets() (string, error) {
	if v := os.Getenv("NVIMWIZ_ASSETS"); v != "" {
		p := filepath.Join(v, "nvim")
		if isDir(p) {
			return p, nil
		}
	}
	if isDir(filepath.Join("assets", "nvim")) {
		return filepath.Join("assets", "nvim"), nil
	}
	exe, err := os.Executable()
	if err == nil {
		p := filepath.Join(filepath.Dir(exe), "assets", "nvim")
		if isDir(p) {
			return p, nil
		}
	}
	return "", errors.New("assets not found")
}

func isDir(p string) bool {
	s, err := os.Stat(p)
	if err != nil {
		return false
	}
	return s.IsDir()
}
