package ui

import "strings"

func (w *Wizard) updateSettingsInfo() {
	if w.settingsInfo == nil {
		return
	}

	profileName := strings.TrimSpace(w.p.Name)
	if profileName == "" {
		profileName = "default"
	}

	presetID := strings.TrimSpace(w.p.Preset)
	presetDisplay := presetID
	presetDesc := ""
	if pr, ok := w.cat.Presets[presetID]; ok {
		if strings.TrimSpace(pr.Title) != "" {
			presetDisplay = pr.Title
		}
		presetDesc = strings.TrimSpace(pr.Short)
	}

	lines := []string{}
	lines = append(lines, "Profile: "+profileName)
	lines = append(lines, "Preset: "+presetDisplay)
	if presetDesc != "" {
		lines = append(lines, presetDesc)
	}
	lines = append(lines, "")

	target := strings.ToLower(strings.TrimSpace(w.p.Target))
	if target == "safe" {
		lines = append(lines, "Target: safe build")
		lines = append(lines, "Build name: "+strings.TrimSpace(w.p.AppName))
		lines = append(lines, "")
		lines = append(lines, "Launch:")
		lines = append(lines, "  NVIM_APPNAME="+w.p.EffectiveAppName()+" nvim")
		lines = append(lines, "")
		lines = append(lines, "Config mode: managed")
		lines = append(lines, "Safe builds always use a separate config directory.")
	} else {
		lines = append(lines, "Target: system config")
		lines = append(lines, "")
		lines = append(lines, "Launch:")
		lines = append(lines, "  nvim")
		lines = append(lines, "")
		lines = append(lines, "Config mode: "+strings.TrimSpace(w.p.ConfigMode))
		lines = append(lines, "System config writes to ~/.config/nvim")
	}

	lines = append(lines, "")
	lines = append(lines, "Verify downloads: "+strings.TrimSpace(w.p.Verify))
	lines = append(lines, "")
	lines = append(lines, "Projects dir: "+strings.TrimSpace(w.p.ProjectsDir))
	lines = append(lines, "Leader: "+encodeKeyForUI(w.p.Leader))
	lines = append(lines, "Local leader: "+encodeKeyForUI(w.p.LocalLeader))
	lines = append(lines, "")
	lines = append(lines, "Tip: select a setting to see details.")

	w.settingsInfo.SetText(strings.Join(lines, "\n"))
}
