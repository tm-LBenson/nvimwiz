package ui

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"nvimwiz/internal/env"
)

func installFeatureBinary(featureID string) (string, bool) {
	switch featureID {
	case "install.neovim":
		return "nvim", true
	case "install.ripgrep":
		return "rg", true
	case "install.fd":
		return "fd", true
	default:
		return "", false
	}
}

func installFeaturePresent(featureID string) bool {
	bin, ok := installFeatureBinary(featureID)
	if !ok || bin == "" {
		return false
	}

	if _, err := exec.LookPath(bin); err == nil {
		return true
	}

	lb, err := env.LocalBin()
	if err != nil {
		return false
	}

	name := bin
	if runtime.GOOS == "windows" {
		name = name + ".exe"
	}
	_, err = os.Stat(filepath.Join(lb, name))
	return err == nil
}
