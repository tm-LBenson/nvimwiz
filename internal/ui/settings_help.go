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
			"Profiles are saved sets of settings and feature choices.",
			"Use multiple profiles to keep different Neovim builds or workflows.",
			"",
			"Current: "+profileName,
		)

	case "target":
		lines = append(lines,
			"Info: Target",
			"",
			"safe build writes to ~/.config/<build name> and does not touch ~/.config/nvim.",
			"system config writes to your normal ~/.config/nvim.",
			"",
			"Current: "+target,
		)

	case "build_name":
		lines = append(lines,
			"Info: Build name",
			"",
			"Used for safe builds only.",
			"This becomes the folder name under ~/.config.",
			"",
			"Build name: "+strings.TrimSpace(w.p.AppName),
			"Effective app name: "+w.p.EffectiveAppName(),
			"",
			"Launch:",
			"  NVIM_APPNAME="+w.p.EffectiveAppName()+" nvim",
		)

	case "preset":
		presetID := strings.TrimSpace(w.p.Preset)
		presetDisplay := presetID
		presetDesc := ""
		if pr, ok := w.cat.Presets[presetID]; ok {
			if strings.TrimSpace(pr.Title) != "" {
				presetDisplay = pr.Title
			}
			presetDesc = strings.TrimSpace(pr.Short)
		}

		lines = append(lines,
			"Info: Preset",
			"",
			"A preset is a curated baseline of defaults and features.",
			"Pick one to start from a known working setup, then customize features.",
			"",
			"Current: "+presetDisplay,
		)
		if presetDesc != "" {
			lines = append(lines, presetDesc)
		}

	case "config_mode":
		lines = append(lines,
			"Info: Config mode",
			"",
			"managed means nvimwiz writes init.lua and generated modules for you.",
			"integrate means you keep your init.lua and add a small loader require.",
			"",
			"Current: "+w.p.ConfigMode,
		)
		if target == "safe" {
			lines = append(lines, "", "Note: safe builds are always managed.")
		}

	case "projects_dir":
		lines = append(lines,
			"Info: Projects dir",
			"",
			"Used by the start screen and pickers to list your projects.",
			"Point it at the folder where you keep your repos.",
			"",
			"Current: "+w.p.ProjectsDir,
		)

	case "leader":
		lines = append(lines,
			"Info: Leader",
			"",
			"Leader prefixes shortcuts in many Neovim setups.",
			"Space is a common choice.",
			"",
			"Current: "+encodeKeyForUI(w.p.Leader),
		)

	case "local_leader":
		lines = append(lines,
			"Info: Local leader",
			"",
			"Local leader is used by some plugins for filetype-specific shortcuts.",
			"",
			"Current: "+encodeKeyForUI(w.p.LocalLeader),
		)

	case "verify":
		lines = append(lines,
			"Info: Verify downloads",
			"",
			"auto verifies when checksums are available.",
			"require fails if verification is not possible.",
			"off skips verification.",
			"",
			"Current: "+w.p.Verify,
		)

	default:
		w.updateSettingsInfo()
		return
	}

	lines = append(lines, "", "Press Save to return to summary.")
	w.settingsInfo.SetText(strings.Join(lines, "\n"))
}
