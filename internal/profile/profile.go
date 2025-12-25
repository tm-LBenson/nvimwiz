package profile

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"nvimwiz/internal/catalog"
)

const CurrentVersion = 2

type State struct {
	Current string `json:"current"`
}

type Profile struct {
	Version     int               `json:"version"`
	Name        string            `json:"name"`
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

func Load(cat catalog.Catalog) (Profile, bool, error) {
	name, p, ok, err := LoadCurrent(cat)
	if err != nil {
		return Profile{}, false, err
	}
	p.Name = name
	return p, ok, nil
}

func Save(p Profile) error {
	st, err := LoadState()
	if err != nil {
		return err
	}
	name := strings.TrimSpace(st.Current)
	if name == "" {
		name = "default"
	}
	return SaveAs(name, p)
}

func LoadCurrent(cat catalog.Catalog) (string, Profile, bool, error) {
	if err := migrateLegacyProfileIfNeeded(cat); err != nil {
		return "", Profile{}, false, err
	}

	st, err := LoadState()
	if err != nil {
		return "", Profile{}, false, err
	}

	name := sanitizeProfileName(st.Current)
	if name == "" {
		name = "default"
	}

	p, ok, err := LoadByName(name, cat)
	if err != nil {
		return "", Profile{}, false, err
	}
	p.Name = name
	return name, p, ok, nil
}

func LoadByName(name string, cat catalog.Catalog) (Profile, bool, error) {
	name = sanitizeProfileName(name)
	if name == "" {
		name = "default"
	}

	path, err := profilePath(name)
	if err != nil {
		return Profile{}, false, err
	}

	b, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			p := Default(cat)
			p.Name = name
			return p, false, nil
		}
		return Profile{}, false, err
	}

	var p Profile
	if err := json.Unmarshal(b, &p); err != nil {
		p = Default(cat)
		p.Name = name
		return p, true, nil
	}

	p.Name = name
	p.Normalize(cat)
	return p, true, nil
}

func SaveAs(name string, p Profile) error {
	name = sanitizeProfileName(name)
	if name == "" {
		name = "default"
	}
	p.Name = name

	path, err := profilePath(name)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	if p.Version == 0 {
		p.Version = CurrentVersion
	}

	b, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err
	}
	b = append(b, '\n')
	return os.WriteFile(path, b, 0o644)
}

func ListProfiles() ([]string, error) {
	if err := ensureDirs(); err != nil {
		return nil, err
	}

	path, err := profilesDir()
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return []string{"default"}, nil
		}
		return nil, err
	}

	seen := map[string]bool{}
	out := []string{}
	for _, ent := range entries {
		if ent.IsDir() {
			continue
		}
		name := ent.Name()
		if !strings.HasSuffix(name, ".json") {
			continue
		}
		base := strings.TrimSuffix(name, ".json")
		base = sanitizeProfileName(base)
		if base == "" {
			continue
		}
		if !seen[base] {
			seen[base] = true
			out = append(out, base)
		}
	}

	if !seen["default"] {
		out = append(out, "default")
	}

	sort.Strings(out)
	return out, nil
}

func Exists(name string) (bool, error) {
	name = sanitizeProfileName(name)
	if name == "" {
		return false, nil
	}
	path, err := profilePath(name)
	if err != nil {
		return false, err
	}
	_, err = os.Stat(path)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return false, err
}

func Delete(name string) error {
	name = sanitizeProfileName(name)
	if name == "" {
		return fmt.Errorf("invalid profile name")
	}
	if name == "default" {
		return fmt.Errorf("cannot delete default profile")
	}

	path, err := profilePath(name)
	if err != nil {
		return err
	}
	if err := os.Remove(path); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}

	st, err := LoadState()
	if err == nil {
		if sanitizeProfileName(st.Current) == name {
			_ = SetCurrent("default")
		}
	}
	return nil
}

func Clone(src, dst string, cat catalog.Catalog) error {
	src = sanitizeProfileName(src)
	dst = sanitizeProfileName(dst)
	if src == "" || dst == "" {
		return fmt.Errorf("invalid profile name")
	}
	if src == dst {
		return fmt.Errorf("source and destination are the same")
	}

	ok, err := Exists(dst)
	if err != nil {
		return err
	}
	if ok {
		return fmt.Errorf("profile %q already exists", dst)
	}

	p, ok, err := LoadByName(src, cat)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("profile %q not found", src)
	}

	p.Name = dst
	if strings.ToLower(strings.TrimSpace(p.Target)) == "safe" {
		p.AppName = ""
	}
	p.Normalize(cat)
	return SaveAs(dst, p)
}

func Rename(oldName, newName string, cat catalog.Catalog) error {
	oldName = sanitizeProfileName(oldName)
	newName = sanitizeProfileName(newName)
	if oldName == "" || newName == "" {
		return fmt.Errorf("invalid profile name")
	}
	if oldName == "default" {
		return fmt.Errorf("cannot rename default profile")
	}
	if oldName == newName {
		return nil
	}

	ok, err := Exists(newName)
	if err != nil {
		return err
	}
	if ok {
		return fmt.Errorf("profile %q already exists", newName)
	}

	p, ok, err := LoadByName(oldName, cat)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("profile %q not found", oldName)
	}

	p.Name = newName
	if strings.ToLower(strings.TrimSpace(p.Target)) == "safe" {
		if sanitizeAppName(p.AppName) == sanitizeAppName(defaultAppName(oldName)) {
			p.AppName = ""
		}
	}
	p.Normalize(cat)
	if err := SaveAs(newName, p); err != nil {
		return err
	}

	oldPath, err := profilePath(oldName)
	if err == nil {
		_ = os.Remove(oldPath)
	}

	st, err := LoadState()
	if err == nil {
		if sanitizeProfileName(st.Current) == oldName {
			_ = SetCurrent(newName)
		}
	}
	return nil
}

func LoadState() (State, error) {
	if err := ensureDirs(); err != nil {
		return State{}, err
	}
	path, err := statePath()
	if err != nil {
		return State{}, err
	}
	b, err := os.ReadFile(path)
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

	st.Current = sanitizeProfileName(st.Current)
	if st.Current == "" {
		st.Current = "default"
	}
	return st, nil
}

func SetCurrent(name string) error {
	name = sanitizeProfileName(name)
	if name == "" {
		name = "default"
	}

	if err := ensureDirs(); err != nil {
		return err
	}
	path, err := statePath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	st := State{Current: name}
	b, err := json.MarshalIndent(st, "", "  ")
	if err != nil {
		return err
	}
	b = append(b, '\n')
	return os.WriteFile(path, b, 0o644)
}

func BackupsDir() (string, error) {
	root, err := baseDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, "backups"), nil
}

func BaseDir() (string, error) {
	return baseDir()
}

func ProfilesDir() (string, error) {
	return profilesDir()
}

func Path() (string, error) {
	st, err := LoadState()
	if err != nil {
		return "", err
	}
	name := sanitizeProfileName(st.Current)
	if name == "" {
		name = "default"
	}
	return profilePath(name)
}

func PathFor(profileName string) (string, error) {
	name := sanitizeProfileName(profileName)
	if name == "" {
		return "", errors.New("profile name is empty")
	}
	return profilePath(name)
}

func Default(cat catalog.Catalog) Profile {
	p := Profile{
		Version:     CurrentVersion,
		Name:        "default",
		Preset:      "lazyvim",
		ProjectsDir: "~/projects",
		Leader:      " ",
		LocalLeader: " ",
		ConfigMode:  "managed",
		Verify:      "auto",
		Features:    map[string]bool{},
		Choices:     map[string]string{},
		Target:      "safe",
		AppName:     "",
	}

	if pr, ok := cat.Presets[p.Preset]; ok {
		for featureID, enabled := range pr.Features {
			p.Features[featureID] = enabled
		}
		for choiceKey, choiceValue := range pr.Choices {
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

	p.Name = sanitizeProfileName(p.Name)
	if p.Name == "" {
		p.Name = "default"
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
	if target != "default" && target != "safe" {
		target = "safe"
	}
	p.Target = target

	mode := strings.ToLower(strings.TrimSpace(p.ConfigMode))
	if mode != "integrate" {
		mode = "managed"
	}
	if p.Target == "safe" {
		mode = "managed"
	}
	p.ConfigMode = mode

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

	for choiceKey, choice := range cat.Choices {
		cur, ok := p.Choices[choiceKey]
		if !ok {
			p.Choices[choiceKey] = choice.Default
			continue
		}
		valid := false
		for _, opt := range choice.Options {
			if opt.ID == cur {
				valid = true
				break
			}
		}
		if !valid {
			p.Choices[choiceKey] = choice.Default
		}
	}

	if p.Target == "safe" {
		p.AppName = sanitizeAppName(p.AppName)
		if p.AppName == "" || p.AppName == "nvim" {
			p.AppName = defaultAppName(p.Name)
		}
	} else {
		p.AppName = "nvim"
	}

	for featureID, enabled := range p.Features {
		if !enabled {
			continue
		}
		feature, ok := cat.Features[featureID]
		if !ok {
			continue
		}
		for _, requiredID := range feature.Requires {
			p.Features[requiredID] = true
		}
	}
}

func (p Profile) EffectiveAppName() string {
	if strings.ToLower(strings.TrimSpace(p.Target)) == "default" {
		return "nvim"
	}
	app := sanitizeAppName(p.AppName)
	if app == "" || app == "nvim" {
		return defaultAppName(p.Name)
	}
	return app
}

func defaultAppName(profileName string) string {
	profileName = sanitizeProfileName(profileName)
	if profileName == "" {
		profileName = "default"
	}
	return "nvimwiz-" + profileName
}

func sanitizeAppName(value string) string {
	value = strings.TrimSpace(strings.ToLower(value))
	value = strings.ReplaceAll(value, " ", "-")
	out := make([]rune, 0, len(value))
	for _, r := range value {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			out = append(out, r)
		}
	}
	res := strings.Trim(string(out), "-_")
	if len(res) > 40 {
		res = res[:40]
	}
	return res
}

func sanitizeProfileName(value string) string {
	value = strings.TrimSpace(strings.ToLower(value))
	value = strings.ReplaceAll(value, " ", "-")
	out := make([]rune, 0, len(value))
	for _, r := range value {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			out = append(out, r)
		}
	}
	res := strings.Trim(string(out), "-_")
	if res == "" {
		return ""
	}
	if res == "nvim" {
		return "nvim-profile"
	}
	if len(res) > 40 {
		res = res[:40]
	}
	return res
}

func baseDir() (string, error) {
	if v := strings.TrimSpace(os.Getenv("XDG_CONFIG_HOME")); v != "" {
		return filepath.Join(v, "nvimwiz"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "nvimwiz"), nil
}

func profilesDir() (string, error) {
	root, err := baseDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, "profiles"), nil
}

func profilePath(name string) (string, error) {
	dir, err := profilesDir()
	if err != nil {
		return "", err
	}
	name = sanitizeProfileName(name)
	if name == "" {
		name = "default"
	}
	return filepath.Join(dir, name+".json"), nil
}

func statePath() (string, error) {
	root, err := baseDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, "state.json"), nil
}

func legacyPath() (string, error) {
	root, err := baseDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, "profile.json"), nil
}

func ensureDirs() error {
	root, err := baseDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(root, 0o755); err != nil {
		return err
	}
	dir, err := profilesDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	return nil
}

func migrateLegacyProfileIfNeeded(cat catalog.Catalog) error {
	if err := ensureDirs(); err != nil {
		return err
	}

	dir, err := profilesDir()
	if err != nil {
		return err
	}

	hasAnyProfileFile := false
	if entries, err := os.ReadDir(dir); err == nil {
		for _, ent := range entries {
			if ent.IsDir() {
				continue
			}
			if strings.HasSuffix(ent.Name(), ".json") {
				hasAnyProfileFile = true
				break
			}
		}
	}

	stateFile, err := statePath()
	if err != nil {
		return err
	}
	if _, err := os.Stat(stateFile); err != nil {
		_ = SetCurrent("default")
	}

	if hasAnyProfileFile {
		return ensureDefaultProfile(cat)
	}

	legacy, err := legacyPath()
	if err != nil {
		return err
	}

	b, err := os.ReadFile(legacy)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return ensureDefaultProfile(cat)
		}
		return err
	}

	var p Profile
	if err := json.Unmarshal(b, &p); err != nil {
		p = Default(cat)
	}
	p.Name = "default"
	p.Normalize(cat)
	if err := SaveAs("default", p); err != nil {
		return err
	}
	_ = SetCurrent("default")
	return nil
}

func ensureDefaultProfile(cat catalog.Catalog) error {
	ok, err := Exists("default")
	if err != nil {
		return err
	}
	if ok {
		return nil
	}
	p := Default(cat)
	p.Name = "default"
	p.Normalize(cat)
	return SaveAs("default", p)
}
