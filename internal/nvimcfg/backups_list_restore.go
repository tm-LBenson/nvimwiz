package nvimcfg

import (
	"errors"
	"os"
	"path/filepath"
	"sort"
	"time"
)

type Backup struct {
	ID        string
	CreatedAt string
	Source    string
	Reason    string
	Path      string
}

func ListBackups() ([]Backup, error) {
	backupRoot, err := BackupsDir()
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(backupRoot)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return []Backup{}, nil
		}
		return nil, err
	}
	items := []Backup{}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		id := e.Name()
		p := filepath.Join(backupRoot, id)
		meta, ok := readBackupMeta(p)
		if ok {
			items = append(items, Backup{
				ID:        meta.ID,
				CreatedAt: meta.CreatedAt,
				Source:    meta.Source,
				Reason:    meta.Reason,
				Path:      p,
			})
			continue
		}
		items = append(items, Backup{ID: id, Path: p})
	}
	sort.Slice(items, func(i, j int) bool { return items[i].ID > items[j].ID })
	return items, nil
}

func RestoreBackupToDefault(id string, log func(string)) error {
	backupRoot, err := BackupsDir()
	if err != nil {
		return err
	}
	backupPath := filepath.Join(backupRoot, id)
	srcCfg := filepath.Join(backupPath, "config")
	if _, err := os.Stat(srcCfg); err != nil {
		srcCfg = backupPath
	}

	defaultRoot, err := ConfigDirForAppName("nvim")
	if err != nil {
		return err
	}

	if _, err := os.Stat(defaultRoot); err == nil {
		preRoot, err := BackupsDir()
		if err != nil {
			return err
		}
		if err := os.MkdirAll(preRoot, 0o755); err != nil {
			return err
		}
		ts := time.Now().Format("20060102-150405")
		id2 := "nvim-pre-restore-" + ts
		p2 := filepath.Join(preRoot, id2)
		if err := os.MkdirAll(p2, 0o755); err != nil {
			return err
		}
		if err := moveDir(defaultRoot, filepath.Join(p2, "config")); err != nil {
			return err
		}
		meta := BackupMeta{
			ID:        id2,
			CreatedAt: time.Now().Format(time.RFC3339),
			Source:    defaultRoot,
			Reason:    "pre-restore",
		}
		_ = writeBackupMeta(p2, meta)
		if log != nil {
			log("Saved current Neovim config to " + p2)
		}
	}

	if err := os.MkdirAll(defaultRoot, 0o755); err != nil {
		return err
	}

	if err := copyDir(srcCfg, defaultRoot, nil); err != nil {
		return err
	}

	if log != nil {
		log("Restored Neovim config from " + backupPath)
	}
	return nil
}
