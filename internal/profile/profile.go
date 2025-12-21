package profile

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"time"
)

const CurrentVersion = 1

// Profile is the persisted state of the wizard.
//
// Design goals:
//   - Backwards compatible (Versioned)
//   - Human readable (pretty-printed JSON)
//   - Safe defaults
//
// NOTE: We keep it fairly flat for now. If we grow beyond this, we can
// move to a "features" map without breaking older profiles.
type Profile struct {
	Version int `json:"version"`

	ProjectsDir  string `json:"projectsDir"`
	Leader       string `json:"leader"`
	LocalLeader  string `json:"localLeader"`
	ConfigMode   string `json:"configMode"` // "managed" or "integrate"
	LastModified string `json:"lastModified"`

	InstallNeovim   bool `json:"installNeovim"`
	InstallRipgrep  bool `json:"installRipgrep"`
	InstallFd       bool `json:"installFd"`
	WriteNvimConfig bool `json:"writeNvimConfig"`

	EnableTree      bool `json:"enableTree"`
	EnableTelescope bool `json:"enableTelescope"`
	EnableLSP       bool `json:"enableLSP"`
	EnableJava      bool `json:"enableJava"`

	RunLazySync      bool `json:"runLazySync"`
	RequireChecksums bool `json:"requireChecksums"`
}

func Default() Profile {
	return Profile{
		Version:         CurrentVersion,
		ProjectsDir:     "~/projects",
		Leader:          " ",
		LocalLeader:     " ",
		ConfigMode:      "managed",
		InstallNeovim:   true,
		InstallRipgrep:  true,
		InstallFd:       true,
		WriteNvimConfig: true,
		EnableTree:      true,
		EnableTelescope: true,
		EnableLSP:       true,
		EnableJava:      false,
		RunLazySync:     true,
	}
}

func (p *Profile) Normalize() {
	if p.Version == 0 {
		p.Version = CurrentVersion
	}
	if p.ProjectsDir == "" {
		p.ProjectsDir = "~/projects"
	}
	if p.Leader == "" {
		p.Leader = " "
	}
	if p.LocalLeader == "" {
		p.LocalLeader = " "
	}
	if p.ConfigMode == "" {
		p.ConfigMode = "managed"
	}
	if p.ConfigMode != "managed" && p.ConfigMode != "integrate" {
		p.ConfigMode = "managed"
	}
}

func ConfigDir() (string, error) {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "nvimwiz"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "nvimwiz"), nil
}

func ProfilePath() (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "profile.json"), nil
}

// Load returns the profile, whether it existed, and an error if something went wrong.
func Load() (Profile, bool, error) {
	path, err := ProfilePath()
	if err != nil {
		return Profile{}, false, err
	}
	b, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			p := Default()
			p.Normalize()
			return p, false, nil
		}
		return Profile{}, false, err
	}

	var p Profile
	if err := json.Unmarshal(b, &p); err != nil {
		return Profile{}, true, err
	}
	p.Normalize()
	return p, true, nil
}

func Save(p Profile) error {
	p.Normalize()
	p.Version = CurrentVersion
	p.LastModified = time.Now().Format(time.RFC3339)

	path, err := ProfilePath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	b, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err
	}
	b = append(b, '\n')
	return os.WriteFile(path, b, 0o644)
}
