package nvimcfg

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"

	"nvimwiz/internal/profile"
)

type BackupMeta struct {
	ID        string `json:"id"`
	CreatedAt string `json:"createdAt"`
	Source    string `json:"source"`
	Reason    string `json:"reason"`
}

func BackupsDir() (string, error) {
	if xdgConfigHome := strings.TrimSpace(os.Getenv("XDG_CONFIG_HOME")); xdgConfigHome != "" {
		return filepath.Join(xdgConfigHome, "nvimwiz", "backups"), nil
	}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".config", "nvimwiz", "backups"), nil
}

func BackupDefaultConfigIfNeeded(p profile.Profile, log func(string)) (string, bool, error) {
	target := strings.ToLower(strings.TrimSpace(p.Target))
	if target != "default" {
		return "", false, nil
	}

	root, err := ConfigDirForAppName("nvim")
	if err != nil {
		return "", false, err
	}

	if _, err := os.Stat(root); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", false, nil
		}
		return "", false, err
	}

	marker, ok, err := ReadMarker(root)
	if err != nil {
		return "", false, err
	}
	if ok && strings.ToLower(strings.TrimSpace(marker.ManagedBy)) == "nvimwiz" {
		return "", false, nil
	}

	backupRoot, err := BackupsDir()
	if err != nil {
		return "", false, err
	}
	if err := os.MkdirAll(backupRoot, 0o755); err != nil {
		return "", false, err
	}

	timestamp := time.Now().Format("20060102-150405")
	backupID := "nvim-" + timestamp
	backupPath := filepath.Join(backupRoot, backupID)
	if err := os.MkdirAll(backupPath, 0o755); err != nil {
		return "", false, err
	}

	dst := filepath.Join(backupPath, "config")
	if err := moveDir(root, dst); err != nil {
		return "", false, err
	}

	meta := BackupMeta{
		ID:        backupID,
		CreatedAt: time.Now().Format(time.RFC3339),
		Source:    root,
		Reason:    "replace-default",
	}
	_ = writeBackupMeta(backupPath, meta)

	if log != nil {
		log("Backed up existing Neovim config to " + backupPath)
	}

	return backupPath, true, nil
}

func writeBackupMeta(dir string, meta BackupMeta) error {
	pth := filepath.Join(dir, "meta.json")
	b, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return err
	}
	b = append(b, '\n')
	return os.WriteFile(pth, b, 0o644)
}

func moveDir(src, dst string) error {
	if err := os.Rename(src, dst); err == nil {
		return nil
	}
	if err := os.MkdirAll(dst, 0o755); err != nil {
		return err
	}
	if err := copyDir(src, dst, nil); err != nil {
		return err
	}
	return os.RemoveAll(src)
}
