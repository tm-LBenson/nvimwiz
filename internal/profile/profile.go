package profile

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"

	"nvimwiz/internal/catalog"
)

const CurrentVersion = 1

type Profile struct {
	Version     int               `json:"version"`
	Preset      string            `json:"preset"`
	ProjectsDir string            `json:"projectsDir"`
	Leader      string            `json:"leader"`
	LocalLeader string            `json:"localLeader"`
	ConfigMode  string            `json:"configMode"`
	Verify      string            `json:"verify"`
	Features    map[string]bool   `json:"features"`
	Choices     map[string]string `json:"choices"`
}

func Path() (string, error) {
	if v := strings.TrimSpace(os.Getenv("XDG_CONFIG_HOME")); v != "" {
		return filepath.Join(v, "nvimwiz", "profile.json"), nil
	}
	h, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(h, ".config", "nvimwiz", "profile.json"), nil
}

func Load(cat catalog.Catalog) (Profile, bool, error) {
	pth, err := Path()
	if err != nil {
		return Profile{}, false, err
	}
	b, err := os.ReadFile(pth)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			p := Default(cat)
			return p, false, nil
		}
		return Profile{}, false, err
	}
	var p Profile
	if err := json.Unmarshal(b, &p); err != nil {
		p = Default(cat)
		return p, true, nil
	}
	p.Normalize(cat)
	return p, true, nil
}

func Save(p Profile) error {
	pth, err := Path()
	if err != nil {
		return err
	}
	dir := filepath.Dir(pth)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err
	}
	b = append(b, '\n')
	return os.WriteFile(pth, b, 0o644)
}

func Default(cat catalog.Catalog) Profile {
	p := Profile{
		Version:     CurrentVersion,
		Preset:      "lazyvim",
		ProjectsDir: "~/projects",
		Leader:      " ",
		LocalLeader: " ",
		ConfigMode:  "managed",
		Verify:      "auto",
		Features:    map[string]bool{},
		Choices:     map[string]string{},
	}
	if pr, ok := cat.Presets[p.Preset]; ok {
		for fid, v := range pr.Features {
			p.Features[fid] = v
		}
		for ck, cv := range pr.Choices {
			p.Choices[ck] = cv
		}
	}
	p.Normalize(cat)
	return p
}

func (p *Profile) Normalize(cat catalog.Catalog) {
	if p.Version == 0 {
		p.Version = CurrentVersion
	}
	if strings.TrimSpace(p.Preset) == "" {
		p.Preset = "lazyvim"
	}
	if strings.TrimSpace(p.ProjectsDir) == "" {
		p.ProjectsDir = "~/projects"
	}
	if p.Leader == "" {
		p.Leader = " "
	}
	if p.LocalLeader == "" {
		p.LocalLeader = " "
	}
	m := strings.ToLower(strings.TrimSpace(p.ConfigMode))
	if m != "integrate" {
		m = "managed"
	}
	p.ConfigMode = m

	v := strings.ToLower(strings.TrimSpace(p.Verify))
	if v != "require" && v != "off" {
		v = "auto"
	}
	p.Verify = v

	if p.Features == nil {
		p.Features = map[string]bool{}
	}
	if p.Choices == nil {
		p.Choices = map[string]string{}
	}

	for id, f := range cat.Features {
		if _, ok := p.Features[id]; !ok {
			p.Features[id] = f.Default
		}
	}

	for key, ch := range cat.Choices {
		if _, ok := p.Choices[key]; !ok {
			p.Choices[key] = ch.Default
		} else {
			cur := p.Choices[key]
			valid := false
			for _, opt := range ch.Options {
				if opt.ID == cur {
					valid = true
					break
				}
			}
			if !valid {
				p.Choices[key] = ch.Default
			}
		}
	}

	for id, enabled := range p.Features {
		if !enabled {
			continue
		}
		f, ok := cat.Features[id]
		if !ok {
			continue
		}
		for _, req := range f.Requires {
			p.Features[req] = true
		}
	}
}
