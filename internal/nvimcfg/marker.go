package nvimcfg

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Marker struct {
	Version   int    `json:"version"`
	ManagedBy string `json:"managedBy"`
	Target    string `json:"target"`
	AppName   string `json:"appName"`
	Mode      string `json:"mode"`
	UpdatedAt string `json:"updatedAt"`
}

func markerPath(root string) string {
	return filepath.Join(root, ".nvimwiz.json")
}

func ReadMarker(root string) (Marker, bool, error) {
	pth := markerPath(root)
	b, err := os.ReadFile(pth)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Marker{}, false, nil
		}
		return Marker{}, false, err
	}
	var m Marker
	if err := json.Unmarshal(b, &m); err != nil {
		return Marker{}, false, nil
	}
	return m, true, nil
}

func WriteMarker(root string, m Marker) error {
	if m.Version == 0 {
		m.Version = 1
	}
	if strings.TrimSpace(m.UpdatedAt) == "" {
		m.UpdatedAt = time.Now().Format(time.RFC3339)
	}
	pth := markerPath(root)
	b, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	b = append(b, '\n')
	return os.WriteFile(pth, b, 0o644)
}
