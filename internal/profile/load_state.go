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


func stateDir() (string, error) {
	if xdgConfigHome := strings.TrimSpace(os.Getenv("XDG_CONFIG_HOME")); xdgConfigHome != "" {
		return filepath.Join(xdgConfigHome, "nvimwiz"), nil
	}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".config", "nvimwiz"), nil
}

func statePath() (string, error) {
	dir, err := stateDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "state.json"), nil
}

func LoadState() (State, error) {
	pth, err := statePath()
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
	st.Current = strings.TrimSpace(st.Current)
	if st.Current == "" {
		st.Current = "default"
	}
	return st, nil
}

func SetCurrent(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		name = "default"
	}
	pth, err := statePath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(pth), 0o755); err != nil {
		return err
	}
	st := State{Current: name}
	b, err := json.MarshalIndent(st, "", "  ")
	if err != nil {
		return err
	}
	b = append(b, '\n')
	return os.WriteFile(pth, b, 0o644)
}
