package profile

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
)

type State struct {
	Current string `json:"current"`
}

func LoadState() (State, error) {
	pth, err := stateFilePath()
	if err != nil {
		return State{}, err
	}

	b, err := os.ReadFile(pth)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return State{Current: "default"}, nil
		}
		return State{}, err
	}

	var st State
	if err := json.Unmarshal(b, &st); err != nil {
		return State{Current: "default"}, nil
	}

	if strings.TrimSpace(st.Current) == "" {
		st.Current = "default"
	}
	return st, nil
}

func stateFilePath() (string, error) {
	if xdgConfigHome := strings.TrimSpace(os.Getenv("XDG_CONFIG_HOME")); xdgConfigHome != "" {
		return filepath.Join(xdgConfigHome, "nvimwiz", "state.json"), nil
	}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".config", "nvimwiz", "state.json"), nil
}
