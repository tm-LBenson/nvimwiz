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
	Target      string            `json:"target"`
	AppName     string            `json:"appName"`
}

func Path() (string, error) {
	if configHome := strings.TrimSpace(os.Getenv("XDG_CONFIG_HOME")); configHome != "" {
		return filepath.Join(configHome, "nvimwiz", "profile.json"), nil
	}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".config", "nvimwiz", "profile.json"), nil
}

func Load(cat catalog.Catalog) (Profile, bool, error) {
	profilePath, err := Path()
	if err != nil {
		return Profile{}, false, err
	}
	data, err := os.ReadFile(profilePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			p := Default(cat)
			return p, false, nil
		}
		return Profile{}, false, err
	}

	var p Profile
	if err := json.Unmarshal(data, &p); err != nil {
		p = Default(cat)
		return p, true, nil
	}
	p.Normalize(cat)
	return p, true, nil
}

func Save(p Profile) error {
	profilePath, err := Path()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(profilePath), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(profilePath, data, 0o644)
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
		Target:      "safe",
		AppName:     "nvimwiz",
	}

	if preset, ok := cat.Presets[p.Preset]; ok {
		for featureID, enabled := range preset.Features {
			p.Features[featureID] = enabled
		}
		for choiceKey, choiceVal := range preset.Choices {
			p.Choices[choiceKey] = choiceVal
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

	target := strings.ToLower(strings.TrimSpace(p.Target))
	if target != "safe" && target != "default" {
		target = "safe"
	}
	p.Target = target

	verify := strings.ToLower(strings.TrimSpace(p.Verify))
	if verify != "require" && verify != "off" {
		verify = "auto"
	}
	p.Verify = verify

	if p.Features == nil {
		p.Features = map[string]bool{}
	}
	if p.Choices == nil {
		p.Choices = map[string]string{}
	}

	for featureID, feature := range cat.Features {
		if _, ok := p.Features[featureID]; !ok {
			p.Features[featureID] = feature.Default
		}
	}

	for key, choice := range cat.Choices {
		current, ok := p.Choices[key]
		if !ok || strings.TrimSpace(current) == "" {
			p.Choices[key] = choice.Default
			continue
		}
		valid := false
		for _, opt := range choice.Options {
			if opt.ID == current {
				valid = true
				break
			}
		}
		if !valid {
			p.Choices[key] = choice.Default
		}
	}

	for featureID, enabled := range p.Features {
		if !enabled {
			continue
		}
		feature, ok := cat.Features[featureID]
		if !ok {
			continue
		}
		for _, req := range feature.Requires {
			p.Features[req] = true
		}
	}

	mode := strings.ToLower(strings.TrimSpace(p.ConfigMode))
	if mode != "integrate" {
		mode = "managed"
	}
	if p.Target == "safe" {
		mode = "managed"
	}
	p.ConfigMode = mode

	if p.Target == "default" {
		p.AppName = "nvim"
		return
	}

	appName := sanitizeAppName(p.AppName)
	if appName == "" || appName == "nvim" {
		appName = "nvimwiz"
	}
	p.AppName = appName
}

func (p Profile) EffectiveAppName() string {
	if strings.ToLower(strings.TrimSpace(p.Target)) == "default" {
		return "nvim"
	}
	appName := sanitizeAppName(p.AppName)
	if appName == "" || appName == "nvim" {
		return "nvimwiz"
	}
	return appName
}

func sanitizeAppName(name string) string {
	s := strings.TrimSpace(strings.ToLower(name))
	s = strings.ReplaceAll(s, " ", "-")
	out := make([]rune, 0, len(s))
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			out = append(out, r)
		}
	}
	res := strings.Trim(strings.TrimSpace(string(out)), "-_")
	if len(res) > 40 {
		res = res[:40]
	}
	return res
}
