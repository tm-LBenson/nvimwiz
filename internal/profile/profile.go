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
	if xdgConfigHome := strings.TrimSpace(os.Getenv("XDG_CONFIG_HOME")); xdgConfigHome != "" {
		return filepath.Join(xdgConfigHome, "nvimwiz", "profile.json"), nil
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

	profileBytes, err := os.ReadFile(profilePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			defaultProfile := Default(cat)
			return defaultProfile, false, nil
		}
		return Profile{}, false, err
	}

	var loaded Profile
	if err := json.Unmarshal(profileBytes, &loaded); err != nil {
		defaultProfile := Default(cat)
		return defaultProfile, true, nil
	}

	loaded.Normalize(cat)
	return loaded, true, nil
}

func Save(p Profile) error {
	profilePath, err := Path()
	if err != nil {
		return err
	}

	parentDir := filepath.Dir(profilePath)
	if err := os.MkdirAll(parentDir, 0o755); err != nil {
		return err
	}

	encoded, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err
	}
	encoded = append(encoded, '\n')
	return os.WriteFile(profilePath, encoded, 0o644)
}

func Default(cat catalog.Catalog) Profile {
	defaultProfile := Profile{
		Version:     CurrentVersion,
		Preset:      "lazyvim",
		ProjectsDir: "~/projects",
		Leader:      " ",
		LocalLeader: " ",
		ConfigMode:  "managed",
		Verify:      "auto",
		Features:    map[string]bool{},
		Choices:     map[string]string{},
		Target:      "default",
		AppName:     "",
	}

	if preset, ok := cat.Presets[defaultProfile.Preset]; ok {
		for featureID, enabled := range preset.Features {
			defaultProfile.Features[featureID] = enabled
		}
		for choiceKey, value := range preset.Choices {
			defaultProfile.Choices[choiceKey] = value
		}
	}

	defaultProfile.Normalize(cat)
	return defaultProfile
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

	configMode := strings.ToLower(strings.TrimSpace(p.ConfigMode))
	if configMode != "integrate" {
		configMode = "managed"
	}
	p.ConfigMode = configMode

	verifyMode := strings.ToLower(strings.TrimSpace(p.Verify))
	if verifyMode != "require" && verifyMode != "off" {
		verifyMode = "auto"
	}
	p.Verify = verifyMode

	target := strings.ToLower(strings.TrimSpace(p.Target))
	if target != "safe" && target != "default" {
		target = "default"
	}
	p.Target = target

	safeBuildName := sanitizeAppName(p.AppName)
	if safeBuildName == "" || safeBuildName == "nvim" {
		safeBuildName = "nvimwiz"
	}
	p.AppName = safeBuildName

	if p.Features == nil {
		p.Features = map[string]bool{}
	}
	if p.Choices == nil {
		p.Choices = map[string]string{}
	}

	for featureID, feature := range cat.Features {
		if _, exists := p.Features[featureID]; !exists {
			p.Features[featureID] = feature.Default
		}
	}

	for choiceKey, choice := range cat.Choices {
		currentValue, exists := p.Choices[choiceKey]
		if !exists || strings.TrimSpace(currentValue) == "" {
			p.Choices[choiceKey] = choice.Default
			continue
		}

		valid := false
		for _, option := range choice.Options {
			if option.ID == currentValue {
				valid = true
				break
			}
		}
		if !valid {
			p.Choices[choiceKey] = choice.Default
		}
	}

	for {
		changed := false
		for featureID, enabled := range p.Features {
			if !enabled {
				continue
			}
			feature, ok := cat.Features[featureID]
			if !ok {
				continue
			}
			for _, requiredID := range feature.Requires {
				if !p.Features[requiredID] {
					p.Features[requiredID] = true
					changed = true
				}
			}
		}
		if !changed {
			break
		}
	}
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

func sanitizeAppName(input string) string {
	normalized := strings.TrimSpace(strings.ToLower(input))
	normalized = strings.ReplaceAll(normalized, " ", "-")

	runes := make([]rune, 0, len(normalized))
	for _, r := range normalized {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			runes = append(runes, r)
		}
	}

	out := strings.Trim(strings.TrimSpace(string(runes)), "-_")
	if len(out) > 40 {
		out = out[:40]
	}
	return out
}
