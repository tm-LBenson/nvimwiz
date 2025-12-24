package nvimcfg

import (
	"encoding/json"
	"os"
	"path/filepath"
)

func readBackupMeta(dir string) (BackupMeta, bool) {
	pth := filepath.Join(dir, "meta.json")
	b, err := os.ReadFile(pth)
	if err != nil {
		return BackupMeta{}, false
	}
	var meta BackupMeta
	if err := json.Unmarshal(b, &meta); err != nil {
		return BackupMeta{}, false
	}
	if meta.ID == "" {
		meta.ID = filepath.Base(dir)
	}
	return meta, true
}
