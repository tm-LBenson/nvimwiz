package nvimcfg

import (
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"nvimwiz/internal/assets"
	"nvimwiz/internal/catalog"
	"nvimwiz/internal/profile"
)

func ConfigDir() (string, error) {
	if v := strings.TrimSpace(os.Getenv("XDG_CONFIG_HOME")); v != "" {
		return filepath.Join(v, "nvim"), nil
	}
	h, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(h, ".config", "nvim"), nil
}

func Write(p profile.Profile, cat catalog.Catalog, log func(string)) error {
	root, err := ConfigDir()
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
	if err := copyDir(src, root, map[string]bool{
		filepath.Join("lua", "nvimwiz", "user.lua"): true,
	}); err != nil {
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
	if b, err := os.ReadFile(initPath); err == nil {
		if strings.Contains(string(b), "vim.g.nvimwiz_managed = true") {
			return os.WriteFile(initPath, []byte(newInit), 0o644)
		}
		ts := time.Now().Format("20060102-150405")
		bak := initPath + ".bak-" + ts
		if err := os.WriteFile(bak, b, 0o644); err != nil {
			return err
		}
	}
	return os.WriteFile(initPath, []byte(newInit), 0o644)
}

func buildConfigLua(p profile.Profile, cat catalog.Catalog) (string, error) {
	modules := []string{}
	featIDs := make([]string, 0, len(cat.Features))
	for id := range cat.Features {
		featIDs = append(featIDs, id)
	}
	sort.Strings(featIDs)
	for _, id := range featIDs {
		if !p.Features[id] {
			continue
		}
		f, ok := cat.Features[id]
		if !ok {
			continue
		}
		for _, m := range f.Modules {
			modules = append(modules, m)
		}
	}

	choices := map[string]string{}
	choiceKeys := make([]string, 0, len(cat.Choices))
	for key := range cat.Choices {
		choiceKeys = append(choiceKeys, key)
	}
	sort.Strings(choiceKeys)
	for _, key := range choiceKeys {
		val := p.Choices[key]

		ch, ok := cat.Choices[key]
		if !ok {
			continue
		}
		optID := val
		if optID == "" {
			optID = ch.Default
		}
		opt, ok := findChoiceOption(ch, optID)
		if !ok {
			opt, _ = findChoiceOption(ch, ch.Default)
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
		for _, m := range opt.Modules {
			modules = append(modules, m)
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
	for _, k := range []string{"explorer", "theme", "statusline"} {
		v := choices[k]
		if v == "" {
			continue
		}
		b.WriteString("\t\t" + k + " = " + luaString(v) + ",\n")
	}
	b.WriteString("\t},\n")
	b.WriteString("\tlsp = {\n")
	for _, k := range []string{"typescript", "python", "web", "go", "bash", "lua", "java"} {
		b.WriteString("\t\t" + k + " = " + luaBool(lsp[k]) + ",\n")
	}
	b.WriteString("\t},\n")
	b.WriteString("\tmodules = {\n")
	for _, m := range modules {
		b.WriteString("\t\t" + luaString(m) + ",\n")
	}
	b.WriteString("\t},\n")
	b.WriteString("}\n")
	return b.String(), nil
}

func findChoiceOption(ch catalog.Choice, id string) (catalog.ChoiceOption, bool) {
	for _, opt := range ch.Options {
		if opt.ID == id {
			return opt, true
		}
	}
	return catalog.ChoiceOption{}, false
}

func uniq(in []string) []string {
	seen := map[string]bool{}
	out := []string{}
	for _, v := range in {
		if v == "" {
			continue
		}
		if seen[v] {
			continue
		}
		seen[v] = true
		out = append(out, v)
	}
	return out
}

func expandTilde(pth string) (string, error) {
	pth = strings.TrimSpace(pth)
	if pth == "" {
		return "", errors.New("empty path")
	}
	if strings.HasPrefix(pth, "~/") || pth == "~" {
		h, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		if pth == "~" {
			return h, nil
		}
		return filepath.Join(h, strings.TrimPrefix(pth, "~/")), nil
	}
	return pth, nil
}

func luaString(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	s = strings.ReplaceAll(s, "\n", "\\n")
	s = strings.ReplaceAll(s, "\r", "")
	return "\"" + s + "\""
}

func luaBool(b bool) string {
	if b {
		return "true"
	}
	return "false"
}
