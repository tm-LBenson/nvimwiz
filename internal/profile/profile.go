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

// CurrentVersion is the on-disk schema version for Profile JSON files.
const CurrentVersion = 2

// State tracks the user's current profile selection.
// Stored at ~/.config/nvimwiz/state.json (or $XDG_CONFIG_HOME/nvimwiz/state.json).
type State struct {
	Current string `json:"current"`
}

// Profile is a saved configuration profile.
// Each profile is stored as JSON under ~/.config/nvimwiz/profiles/<name>.json.
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

// Load loads the current profile (based on State.Current).
// The bool return indicates whether the profile file existed on disk.
func Load(cat catalog.Catalog) (Profile, bool, error) {
	name, p, ok, err := LoadCurrent(cat)
	if err != nil {
		return Profile{}, false, err
	}
	p.Name = name
	return p, ok, nil
}

// Save saves the current profile (based on State.Current).
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

// LoadCurrent loads the active profile.
// Returns (name, profile, existed).
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
	pth, err := profilePath(name)
	if err != nil {
		return Profile{}, false, err
	}

	b, err := os.ReadFile(pth)
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
		// Corrupt JSON -> reset to defaults but signal that it existed.
		p = Default(cat)
		p.Name = name
		return p, true, nil
	}

	p.Name = name
	p.Normalize(cat)
	return p, true, nil
}

// SaveAs saves the profile under the provided name.
func SaveAs(name string, p Profile) error {
	name = sanitizeProfileName(name)
	if name == "" {
		name = "default"
	}
	p.Name = name

	pth, err := profilePath(name)
	if err != nil {
		return err
	}

	dir := filepath.Dir(pth)
	if err := os.MkdirAll(dir, 0o755); err != nil {
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
	return os.WriteFile(pth, b, 0o644)
}

// ListProfiles returns all saved profile names.
func ListProfiles() ([]string, error) {
	if err := ensureDirs(); err != nil {
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

// Exists reports whether the named profile file exists.
func Exists(name string) (bool, error) {
	name = sanitizeProfileName(name)
	if name == "" {
		return false, nil
	}
	pth, err := profilePath(name)
	if err != nil {
		return false, err
	}
	_, err = os.Stat(pth)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return false, err
}

// Delete removes a profile. The "default" profile cannot be deleted.
func Delete(name string) error {
	name = sanitizeProfileName(name)
	if name == "" {
		return fmt.Errorf("invalid profile name")
	}
	if name == "default" {
		return fmt.Errorf("cannot delete default profile")
	}

	pth, err := profilePath(name)
	if err != nil {
		return err
	}
	if err := os.Remove(pth); err != nil {
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

// Clone creates dst as a copy of src.
func Clone(src, dst string, cat catalog.Catalog) error {
	src = sanitizeProfileName(src)
	dst = sanitizeProfileName(dst)
	if src == "" || dst == "" {
		return fmt.Errorf("invalid profile name")
	}
	if src == dst {
		return fmt.Errorf("source and destination are the same")
	}

	if ok, err := Exists(dst); err != nil {
		return err
	} else if ok {
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
	// Avoid NVIM_APPNAME collisions for safe builds.
	if strings.ToLower(strings.TrimSpace(p.Target)) == "safe" {
		p.AppName = ""
	}
	p.Normalize(cat)

	return SaveAs(dst, p)
}

// Rename renames a profile file and updates state.json if needed.
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
	if ok, err := Exists(newName); err != nil {
		return err
	} else if ok {
		return fmt.Errorf("profile %q already exists", newName)
	}

	p, ok, err := LoadByName(oldName, cat)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("profile %q not found", oldName)
	}

	// Write a new file under the new name (so we can also update internal fields).
	p.Name = newName
	if strings.ToLower(strings.TrimSpace(p.Target)) == "safe" {
		// If appname was the default for the old profile, clear it so the new
		// profile gets its own default name.
		if sanitizeAppName(p.AppName) == sanitizeAppName(defaultAppName(oldName)) {
			p.AppName = ""
		}
	}
	p.Normalize(cat)
	if err := SaveAs(newName, p); err != nil {
		return err
	}

	// Remove the old file.
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
func LoadState() (State, error) {
	if err := ensureDirs(); err != nil {
		return State{}, err
	}
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
	if strings.TrimSpace(st.Current) == "" {
		st.Current = "default"
	}
	st.Current = sanitizeProfileName(st.Current)
	if st.Current == "" {
		st.Current = "default"
	}
	return st, nil
}

// SetCurrent updates ~/.config/nvimwiz/state.json.
func SetCurrent(name string) error {
	name = sanitizeProfileName(name)
	if name == "" {
		name = "default"
	}
	if err := ensureDirs(); err != nil {
		return err
	}
	pth, err := statePath()
	if err != nil {
		return err
	}
	dir := filepath.Dir(pth)
	if err := os.MkdirAll(dir, 0o755); err != nil {
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

	if p.Target == "default" {
		p.AppName = "nvim"
	} else {
		p.AppName = sanitizeAppName(p.AppName)
		if p.AppName == "" || p.AppName == "nvim" {
			p.AppName = defaultAppName(p.Name)
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

	if names, err := ListProfiles(); err == nil && len(names) > 0 {
		if _, err := os.Stat(must(statePath())); err != nil {
			_ = SetCurrent("default")
		}
		return ensureDefaultProfile(cat)
	}

	legacy, err := legacyPath()
	if err != nil {
		return err
	}
	b, err := os.ReadFile(legacy)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			_ = SetCurrent("default")
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
	// Keep legacy file as a backup. We don't delete it.
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

func must(pth string, err error) string {
	if err != nil {
		return ""
	}
	return pth
}
