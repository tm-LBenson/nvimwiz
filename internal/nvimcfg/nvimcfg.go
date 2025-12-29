package nvimcfg

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"nvimwiz/internal/assets"
	"nvimwiz/internal/catalog"
	"nvimwiz/internal/profile"
)

func ConfigDirForAppName(appName string) (string, error) {
	name := strings.TrimSpace(appName)
	if name == "" {
		name = "nvim"
	}
	if xdgConfigHome := strings.TrimSpace(os.Getenv("XDG_CONFIG_HOME")); xdgConfigHome != "" {
		return filepath.Join(xdgConfigHome, name), nil
	}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".config", name), nil
}

func ConfigDirForProfile(p profile.Profile) (string, error) {
	return ConfigDirForAppName(p.EffectiveAppName())
}

func ConfigDir() (string, error) {
	return ConfigDirForAppName("nvim")
}

func Write(p profile.Profile, cat catalog.Catalog, log func(string)) error {
	root, err := ConfigDirForProfile(p)
	if err != nil {
		return err
	}
	src, err := assets.FindNvimAssets()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(root, 0o755); err != nil {
		return err
	}

	ignore := map[string]bool{
		filepath.Join("lua", "nvimwiz", "user.lua"): true,
	}
	if err := copyDir(src, root, ignore); err != nil {
		return err
	}

	genDir := filepath.Join(root, "lua", "nvimwiz", "generated")
	if err := os.MkdirAll(genDir, 0o755); err != nil {
		return err
	}

	cfgLua, err := buildConfigLua(p, cat)
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(genDir, "config.lua"), []byte(cfgLua), 0o644); err != nil {
		return err
	}

	if p.ConfigMode == "managed" {
		if err := writeInitLua(root); err != nil {
			return err
		}
	}

	headless := "lua require(\"nvimwiz.loader\")\n"
	if err := os.WriteFile(filepath.Join(root, "nvimwiz_headless_init.vim"), []byte(headless), 0o644); err != nil {
		return err
	}

	if log != nil {
		log("Wrote Neovim config to " + root)
	}
	return nil
}

func writeInitLua(root string) error {
	initPath := filepath.Join(root, "init.lua")
	newInit := "vim.g.nvimwiz_managed = true\nrequire(\"nvimwiz.loader\")\n"
	if existingBytes, err := os.ReadFile(initPath); err == nil {
		if strings.Contains(string(existingBytes), "vim.g.nvimwiz_managed = true") {
			return os.WriteFile(initPath, []byte(newInit), 0o644)
		}
		timestamp := time.Now().Format("20060102-150405")
		backupPath := initPath + ".bak-" + timestamp
		if err := os.WriteFile(backupPath, existingBytes, 0o644); err != nil {
			return err
		}
	}
	return os.WriteFile(initPath, []byte(newInit), 0o644)
}

func buildConfigLua(p profile.Profile, cat catalog.Catalog) (string, error) {
	modules := []string{}
	featureIDs := make([]string, 0, len(cat.Features))
	for id := range cat.Features {
		featureIDs = append(featureIDs, id)
	}
	sort.Strings(featureIDs)

	for _, id := range featureIDs {
		if !p.Features[id] {
			continue
		}
		feature, ok := cat.Features[id]
		if !ok {
			continue
		}
		for _, mod := range feature.Modules {
			modules = append(modules, mod)
		}
	}

	choices := map[string]string{}
	choiceKeys := make([]string, 0, len(cat.Choices))
	for key := range cat.Choices {
		choiceKeys = append(choiceKeys, key)
	}
	sort.Strings(choiceKeys)

	for _, key := range choiceKeys {
		choice, ok := cat.Choices[key]
		if !ok {
			continue
		}
		optID := strings.TrimSpace(p.Choices[key])
		if optID == "" {
			optID = choice.Default
		}

		opt, ok := findChoiceOption(choice, optID)
		if !ok {
			opt, _ = findChoiceOption(choice, choice.Default)
			optID = opt.ID
		}

		if key == "ui.explorer" {
			choices["explorer"] = optID
		}
		if key == "ui.theme" {
			choices["theme"] = optID
		}
		if key == "ui.statusline" {
			choices["statusline"] = optID
		}
		if key == "core.linenumbers" {
			choices["line_numbers"] = optID
		}

		for _, mod := range opt.Modules {
			modules = append(modules, mod)
		}
	}

	modules = uniq(modules)

	projectsDir, err := expandTilde(p.ProjectsDir)
	if err != nil {
		return "", err
	}

	lsp := map[string]bool{
		"typescript": p.Features["lsp.typescript"],
		"python":     p.Features["lsp.python"],
		"web":        p.Features["lsp.web"],
		"emmet":      p.Features["lsp.emmet"],
		"go":         p.Features["lsp.go"],
		"bash":       p.Features["lsp.bash"],
		"lua":        p.Features["lsp.lua"],
		"java":       p.Features["lsp.java"],
	}

	b := &strings.Builder{}
	b.WriteString("return {\n")
	b.WriteString("\tleader = " + luaString(p.Leader) + ",\n")
	b.WriteString("\tlocalleader = " + luaString(p.LocalLeader) + ",\n")
	b.WriteString("\tprojects_dir = " + luaString(projectsDir) + ",\n")
	b.WriteString("\tchoices = {\n")
	for _, key := range []string{"explorer", "theme", "statusline", "line_numbers"} {
		val := choices[key]
		if val == "" {
			continue
		}
		b.WriteString("\t\t" + key + " = " + luaString(val) + ",\n")
	}
	b.WriteString("\t},\n")
	b.WriteString("\tlsp = {\n")
	for _, key := range []string{"typescript", "python", "web", "emmet", "go", "bash", "lua", "java"} {
		b.WriteString("\t\t" + key + " = " + luaBool(lsp[key]) + ",\n")
	}
	b.WriteString("\t},\n")
	b.WriteString("\tmodules = {\n")
	for _, mod := range modules {
		b.WriteString("\t\t" + luaString(mod) + ",\n")
	}
	b.WriteString("\t},\n")
	b.WriteString("}\n")
	return b.String(), nil
}

func findChoiceOption(choice catalog.Choice, id string) (catalog.ChoiceOption, bool) {
	for _, opt := range choice.Options {
		if opt.ID == id {
			return opt, true
		}
	}
	return catalog.ChoiceOption{}, false
}
