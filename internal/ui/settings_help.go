package ui

import "strings"

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
		presetID := strings.TrimSpace(w.p.Preset)
		pr, ok := w.cat.Presets[presetID]
		name := presetID
		desc := ""
		if ok {
			if strings.TrimSpace(pr.Title) != "" {
				name = pr.Title
			}
			desc = strings.TrimSpace(pr.Short)
		}
		lines = append(lines,
			"Info: Preset",
			"",
			"A preset is a curated starting point. It sets a recommended baseline of features and UI choices.",
			"You can still customize everything after selecting a preset.",
			"",
			"Current: "+name,
		)
		if desc != "" {
			lines = append(lines, "", desc)
		}
		lines = append(lines,
			"",
			"How to use presets:",
			"- pick the closest vibe (minimal vs IDE-like)",
			"- then go to Features to toggle what you want",
		)

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
