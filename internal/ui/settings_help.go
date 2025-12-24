package ui

import "strings"

func (w *Wizard) showSettingsFieldHelp(fieldKey string) {
	if w.settingsInfo == nil {
		return
	}

	target := strings.ToLower(strings.TrimSpace(w.p.Target))

	lines := []string{}
	switch fieldKey {
	case "target":
		lines = append(lines, "Info: Target", "")
		lines = append(lines, "default config writes to your normal ~/.config/nvim")
		lines = append(lines, "safe build writes to ~/.config/<build name> and does not touch your current config")
		lines = append(lines, "", "Current: "+target)

	case "build_name":
		lines = append(lines, "Info: Build name", "")
		lines = append(lines, "Used for safe builds only.")
		lines = append(lines, "This becomes the folder name under ~/.config.")
		lines = append(lines, "", "Build name: "+w.p.AppName)
		lines = append(lines, "Effective app name: "+w.p.EffectiveAppName())
		lines = append(lines, "")
		if target == "safe" {
			lines = append(lines, "Launch: NVIM_APPNAME="+w.p.EffectiveAppName()+" nvim")
		} else {
			lines = append(lines, "Launch: nvim")
		}

	case "preset":
		lines = append(lines, "Info: Preset", "")
		lines = append(lines, "A preset enables a curated set of features and defaults.")
		lines = append(lines, "Use it to start from a known working baseline.")
		lines = append(lines, "", "Current: "+w.p.Preset)

	case "config_mode":
		lines = append(lines, "Info: Config mode", "")
		lines = append(lines, "managed means nvimwiz owns init.lua and generated modules")
		lines = append(lines, "integrate means you keep your init.lua and require nvimwiz.loader")
		lines = append(lines, "", "Current: "+w.p.ConfigMode)

	case "projects_dir":
		lines = append(lines, "Info: Projects dir", "")
		lines = append(lines, "Used by the Neovim start screen to list your projects.")
		lines = append(lines, "Point it at the folder where you keep your repos.")
		lines = append(lines, "", "Current: "+w.p.ProjectsDir)

	case "leader":
		lines = append(lines, "Info: Leader", "")
		lines = append(lines, "Leader prefixes shortcuts in many Neovim setups.")
		lines = append(lines, "Space is a common choice.")
		lines = append(lines, "", "Current: "+encodeKeyForUI(w.p.Leader))

	case "local_leader":
		lines = append(lines, "Info: Local leader", "")
		lines = append(lines, "Local leader is used by some plugins for filetype specific shortcuts.")
		lines = append(lines, "", "Current: "+encodeKeyForUI(w.p.LocalLeader))

	case "verify":
		lines = append(lines, "Info: Verify downloads", "")
		lines = append(lines, "auto verifies when checksums are available")
		lines = append(lines, "require fails if verification is not possible")
		lines = append(lines, "off skips verification")
		lines = append(lines, "", "Current: "+w.p.Verify)

	default:
		w.updateSettingsInfo()
		return
	}

	lines = append(lines, "", "Press Save to return to summary.")
	w.settingsInfo.SetText(strings.Join(lines, "\n"))
}
