package profile

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"
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

type State struct {
	Current string `json:"current"`
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
		Target:      "default",
		AppName:     "",
	}

	if preset, ok := cat.Presets[p.Preset]; ok {
		for featureID, enabled := range preset.Features {
			p.Features[featureID] = enabled
		}
		for choiceKey, choiceValue := range preset.Choices {
			p.Choices[choiceKey] = choiceValue
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

func Load(cat catalog.Catalog) (Profile, bool, error) {
	_, p, err := LoadCurrent(cat)
	if err != nil {
		return Profile{}, false, err
	}
	return p, true, nil
}

func Save(p Profile) error {
	currentName, err := CurrentName()
	if err != nil {
		return err
	}
	return SaveAs(currentName, p)
}

func LoadState() (State, error) {
	state, _, err := loadState()
	return state, err
}

func CurrentName() (string, error) {
	if err := migrateLegacy(); err != nil {
		return "", err
	}
	state, ok, err := loadState()
	if err != nil {
		return "", err
	}
	name := "default"
	if ok {
		name = sanitizeProfileName(state.Current)
		if name == "" {
			name = "default"
		}
	}
	return name, nil
}

func LoadCurrent(cat catalog.Catalog) (string, Profile, error) {
	if err := migrateLegacy(); err != nil {
		return "", Profile{}, err
	}

	state, ok, err := loadState()
	if err != nil {
		return "", Profile{}, err
	}

	name := "default"
	if ok {
		name = sanitizeProfileName(state.Current)
		if name == "" {
			name = "default"
		}
	}

	p, _, err := LoadByName(name, cat)
	if err != nil {
		return "", Profile{}, err
	}
	p.Normalize(cat)
	_ = SaveAs(name, p)
	_ = setCurrent(name)
	return name, p, nil
}

func SetCurrent(name string) error {
	name = sanitizeProfileName(name)
	if name == "" {
		name = "default"
	}
	if err := migrateLegacy(); err != nil {
		return err
	}
	return setCurrent(name)
}

func ListProfiles() ([]string, error) {
	if err := migrateLegacy(); err != nil {
		return nil, err
	}
	dir, err := profilesDir()
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return []string{"default"}, nil
		}
		return nil, err
	}
	names := []string{}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		fileName := entry.Name()
		if !strings.HasSuffix(fileName, ".json") {
			continue
		}
		base := strings.TrimSuffix(fileName, ".json")
		base = sanitizeProfileName(base)
		if base == "" {
			continue
		}
		names = append(names, base)
	}
	if len(names) == 0 {
		names = append(names, "default")
	}
	sort.Strings(names)
	return names, nil
}

func LoadByName(name string, cat catalog.Catalog) (Profile, bool, error) {
	name = sanitizeProfileName(name)
	if name == "" {
		name = "default"
	}
	pth, err := profilePath(name)
	if err != nil {
		return Profile{}, false, err
	}
	bytes, err := os.ReadFile(pth)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			p := Default(cat)
			p.Normalize(cat)
			return p, false, nil
		}
		return Profile{}, false, err
	}
	var p Profile
	if err := json.Unmarshal(bytes, &p); err != nil {
		p = Default(cat)
		p.Normalize(cat)
		return p, true, nil
	}
	p.Normalize(cat)
	return p, true, nil
}

func SaveAs(name string, p Profile) error {
	name = sanitizeProfileName(name)
	if name == "" {
		name = "default"
	}
	dir, err := profilesDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	pth := filepath.Join(dir, name+".json")
	encoded, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err
	}
	encoded = append(encoded, '\n')
	return os.WriteFile(pth, encoded, 0o644)
}

func baseDir() (string, error) {
	if xdgConfigHome := strings.TrimSpace(os.Getenv("XDG_CONFIG_HOME")); xdgConfigHome != "" {
		return filepath.Join(xdgConfigHome, "nvimwiz"), nil
	}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".config", "nvimwiz"), nil
}

func profilesDir() (string, error) {
	base, err := baseDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "profiles"), nil
}

func profilePath(name string) (string, error) {
	dir, err := profilesDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, name+".json"), nil
}

func legacyPath() (string, error) {
	base, err := baseDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "profile.json"), nil
}

func statePath() (string, error) {
	base, err := baseDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "state.json"), nil
}

func loadState() (State, bool, error) {
	pth, err := statePath()
	if err != nil {
		return State{}, false, err
	}
	bytes, err := os.ReadFile(pth)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return State{Current: "default"}, false, nil
		}
		return State{}, false, err
	}
	var state State
	if err := json.Unmarshal(bytes, &state); err != nil {
		return State{Current: "default"}, true, nil
	}
	return state, true, nil
}

func setCurrent(name string) error {
	pth, err := statePath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(pth), 0o755); err != nil {
		return err
	}
	state := State{Current: name}
	encoded, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	encoded = append(encoded, '\n')
	return os.WriteFile(pth, encoded, 0o644)
}

func migrateLegacy() error {
	dir, err := profilesDir()
	if err != nil {
		return err
	}
	defaultProfilePath := filepath.Join(dir, "default.json")
	if _, err := os.Stat(defaultProfilePath); err == nil {
		return nil
	}

	legacyProfilePath, err := legacyPath()
	if err != nil {
		return err
	}

	var p Profile
	legacyBytes, err := os.ReadFile(legacyProfilePath)
	if err == nil {
		if err := json.Unmarshal(legacyBytes, &p); err != nil {
			p = Profile{}
		}
	} else {
		p = Profile{}
	}

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	encoded, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err
	}
	encoded = append(encoded, '\n')
	if err := os.WriteFile(defaultProfilePath, encoded, 0o644); err != nil {
		return err
	}

	if err := setCurrent("default"); err != nil {
		return err
	}

	if err == nil {
		_ = os.Rename(legacyProfilePath, legacyProfilePath+".bak")
	}

	return nil
}

func sanitizeProfileName(input string) string {
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
