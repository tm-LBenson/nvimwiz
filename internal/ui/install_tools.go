package ui

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"nvimwiz/internal/env"
)

func installCommandName(featureID string) (string, bool) {
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

func installToolPresent(featureID string) bool {
	cmd, ok := installCommandName(featureID)
	if !ok {
		return false
	}

	if _, err := exec.LookPath(cmd); err == nil {
		return true
	}

	lb, err := env.LocalBin()
	if err != nil {
		return false
	}

	name := cmd
	if runtime.GOOS == "windows" {
		name = cmd + ".exe"
	}

	_, err = os.Stat(filepath.Join(lb, name))
	return err == nil
}
