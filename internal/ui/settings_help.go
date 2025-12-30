package ui

import (
	"sort"
	"strings"

	"nvimwiz/internal/profile"
)

func (w *Wizard) showSettingsFieldHelp(fieldKey string) {
	if w.settingsInfo == nil {
		return
	}

	profileName := strings.TrimSpace(w.p.Name)
	if profileName == "" {
		profileName = "default"
	}

	target := strings.ToLower(strings.TrimSpace(w.p.Target))

	lines := []string{}
	switch fieldKey {
	case "profile":
		lines = append(lines,
			"Info: Profile",
			"",
			"A profile is a saved set of settings, feature toggles, and UI choices.",
			"Use profiles when you want multiple Neovim setups (work vs personal, minimal vs full IDE, etc).",
			"",
			"What this affects:",
			"- settings on this page",
			"- features/choices on the next page",
			"- where your config is written (safe build vs system config)",
			"",
			"Current: "+profileName,
			"",
			"Tip: use the Profiles button to create/clone/rename/delete profiles.",
		)

	case "target":
		lines = append(lines,
			"Info: Target",
			"",
			"Target controls where nvimwiz writes your Neovim config.",
			"",
			"safe build (recommended):",
			"- writes to ~/.config/<build name>",
			"- never touches your existing ~/.config/nvim",
			"- launch with: NVIM_APPNAME=<build name> nvim",
			"",
			"system config:",
			"- writes to your normal ~/.config/nvim",
			"- launch with: nvim",
			"",
			"Current: "+target,
		)

	case "build_name":
		lines = append(lines,
			"Info: Build name",
			"",
			"Build name is used for safe builds only.",
			"It becomes the folder name under ~/.config and the NVIM_APPNAME value.",
			"",
			"Build name: "+strings.TrimSpace(w.p.AppName),
			"Effective app name: "+w.p.EffectiveAppName(),
		)
		if target == "safe" {
			lines = append(lines,
				"",
				"Launch:",
				"  NVIM_APPNAME="+w.p.EffectiveAppName()+" nvim",
				"",
				"Why this exists:",
				"- you can try nvimwiz without breaking your current setup",
				"- you can keep multiple builds side-by-side",
			)
		}

	case "preset":
		lines = append(lines, w.presetHelp()...)

	case "config_mode":
		lines = append(lines,
			"Info: Config mode",
			"",
			"managed:",
			"- nvimwiz owns init.lua and generated modules",
			"- easiest for a fresh setup",
			"",
			"integrate:",
			"- you keep your init.lua",
			"- you add a small require to load nvimwiz modules",
			"- best when you already have a config you want to keep",
			"",
			"Current: "+strings.TrimSpace(w.p.ConfigMode),
		)
		if target == "safe" {
			lines = append(lines, "", "Note: safe builds always use managed mode.")
		}

	case "projects_dir":
		lines = append(lines,
			"Info: Projects dir",
			"",
			"This directory is used by the dashboard and pickers to show your projects.",
			"Set it to where you keep your git repos (for example: ~/projects).",
			"",
			"Current: "+strings.TrimSpace(w.p.ProjectsDir),
		)

	case "leader":
		lines = append(lines,
			"Info: Leader",
			"",
			"Leader is a special key used as a prefix for shortcuts (keymaps).",
			"Many Neovim configs use <leader>something for commands.",
			"",
			"Common choice: Space",
			"",
			"Current: "+encodeKeyForUI(w.p.Leader),
		)

	case "local_leader":
		lines = append(lines,
			"Info: Local leader",
			"",
			"Local leader is another prefix key used by some plugins for filetype-specific shortcuts.",
			"If you are new, keeping it as Space is totally fine.",
			"",
			"Current: "+encodeKeyForUI(w.p.LocalLeader),
		)

	case "verify":
		lines = append(lines,
			"Info: Verify downloads",
			"",
			"nvimwiz can verify downloaded archives when a checksum is available.",
			"",
			"auto:",
			"- verify when possible; otherwise continue",
			"",
			"require:",
			"- fail if verification is not possible",
			"",
			"off:",
			"- skip verification",
			"",
			"Current: "+strings.TrimSpace(w.p.Verify),
		)

	default:
		w.updateSettingsInfo()
		return
	}

	lines = append(lines, "", "Press Save to return to summary.")
	w.settingsInfo.SetText(strings.Join(lines, "\n"))
}

func (w *Wizard) presetHelp() []string {
	presetID := strings.TrimSpace(w.p.Preset)
	pr, ok := w.cat.Presets[presetID]

	title := presetID
	if ok {
		if strings.TrimSpace(pr.Title) != "" {
			title = pr.Title
		}
	}

	lines := []string{}
	lines = append(lines,
		"Info: Preset",
		"",
		"A preset is a curated starting point. It sets a recommended baseline of features and UI choices.",
		"You can still customize everything after selecting a preset.",
		"",
		"Current: "+title,
	)

	if !ok {
		lines = append(lines,
			"",
			"This preset is not known in the current catalog.",
		)
		return lines
	}

	if s := strings.TrimSpace(pr.Short); s != "" {
		lines = append(lines, "", s)
	}

	if s := strings.TrimSpace(pr.ModeledAfter); s != "" {
		lines = append(lines, "", "Modeled after: "+s)
	}

	links := []string{}
	for _, link := range pr.Links {
		link = strings.TrimSpace(link)
		if link == "" {
			continue
		}
		links = append(links, link)
	}
	if len(links) > 0 {
		lines = append(lines, "", "Links:")
		for _, link := range links {
			lines = append(lines, "- "+link)
		}
	}

	if s := strings.TrimSpace(pr.Audience); s != "" {
		lines = append(lines, "", "Intended audience:")
		for _, l := range splitNonEmptyLines(s) {
			lines = append(lines, l)
		}
	}

	base := profile.Default(w.cat)
	base.Preset = presetID
	for id, enabled := range pr.Features {
		base.Features[id] = enabled
	}
	for key, val := range pr.Choices {
		base.Choices[key] = val
	}
	base.Normalize(w.cat)

	catTitles := map[string][]string{}
	for featureID, enabled := range base.Features {
		if !enabled {
			continue
		}
		f, ok := w.cat.Features[featureID]
		if !ok {
			continue
		}
		catTitles[f.Category] = append(catTitles[f.Category], f.Title)
	}

	lines = append(lines, "", "Defaults:")

	knownCats := map[string]bool{}
	for _, c := range w.cat.Categories {
		knownCats[c] = true
		vals := catTitles[c]
		if len(vals) == 0 {
			continue
		}
		sort.Strings(vals)
		lines = append(lines, "- "+c+": "+strings.Join(vals, ", "))
	}

	extraCats := []string{}
	for c := range catTitles {
		if knownCats[c] {
			continue
		}
		extraCats = append(extraCats, c)
	}
	sort.Strings(extraCats)
	for _, c := range extraCats {
		vals := catTitles[c]
		if len(vals) == 0 {
			continue
		}
		sort.Strings(vals)
		lines = append(lines, "- "+c+": "+strings.Join(vals, ", "))
	}

	choiceKeys := make([]string, 0, len(w.cat.Choices))
	for key := range w.cat.Choices {
		choiceKeys = append(choiceKeys, key)
	}
	sort.SliceStable(choiceKeys, func(i, j int) bool {
		return w.cat.Choices[choiceKeys[i]].Title < w.cat.Choices[choiceKeys[j]].Title
	})

	if len(choiceKeys) > 0 {
		lines = append(lines, "", "UI choices:")
	}
	for _, key := range choiceKeys {
		ch, ok := w.cat.Choices[key]
		if !ok {
			continue
		}
		optID := strings.TrimSpace(base.Choices[key])
		optTitle := optID
		for _, opt := range ch.Options {
			if opt.ID == optID {
				optTitle = opt.Title
				break
			}
		}
		lines = append(lines, "- "+ch.Title+": "+optTitle)
	}

	if s := strings.TrimSpace(pr.Tradeoffs); s != "" {
		lines = append(lines, "", "Tradeoffs:")
		for _, l := range splitNonEmptyLines(s) {
			lines = append(lines, l)
		}
	}

	lines = append(lines,
		"",
		"How to use presets:",
		"- pick the closest vibe (minimal vs IDE-like)",
		"- then go to Features to toggle what you want",
	)
	return lines
}

func splitNonEmptyLines(s string) []string {
	out := []string{}
	for _, l := range strings.Split(s, "\n") {
		l = strings.TrimRight(l, "\r")
		l = strings.TrimSpace(l)
		if l == "" {
			continue
		}
		out = append(out, l)
	}
	return out
}
