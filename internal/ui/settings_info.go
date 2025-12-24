package ui

import "strings"

func (w *Wizard) updateSettingsInfo() {
	if w.settingsInfo == nil {
		return
	}

	pr, ok := w.cat.Presets[w.p.Preset]
	presetTitle := w.p.Preset
	presetShort := ""
	if ok {
		presetTitle = pr.Title
		presetShort = pr.Short
	}

	target := strings.ToLower(strings.TrimSpace(w.p.Target))
	if target != "safe" && target != "default" {
		target = "safe"
	}

	lines := []string{}
	lines = append(lines, "Preset: "+presetTitle)
	if strings.TrimSpace(presetShort) != "" {
		lines = append(lines, presetShort)
	}

	lines = append(lines, "")
	if target == "safe" {
		lines = append(lines, "Target: safe build")
		lines = append(lines, "Build name: "+w.p.EffectiveAppName())
		lines = append(lines, "")
		lines = append(lines, "Launch:")
		lines = append(lines, "  NVIM_APPNAME="+w.p.EffectiveAppName()+" nvim")
		lines = append(lines, "")
		lines = append(lines, "Config mode: managed")
		lines = append(lines, "Safe builds always use a separate config directory.")
	} else {
		lines = append(lines, "Target: system config (~/.config/nvim)")
		lines = append(lines, "")
		lines = append(lines, "Config mode: "+w.p.ConfigMode)
		if w.p.ConfigMode == "integrate" {
			lines = append(lines, "Integrate means you must require nvimwiz.loader from your own init.lua")
		} else {
			lines = append(lines, "Managed means nvimwiz writes ~/.config/nvim/init.lua")
		}
	}

	lines = append(lines, "")
	lines = append(lines, "Verify downloads: "+w.p.Verify)
	lines = append(lines, "")
	lines = append(lines, "Projects dir: "+w.p.ProjectsDir)
	lines = append(lines, "Leader: "+encodeKeyForUI(w.p.Leader))
	lines = append(lines, "Local leader: "+encodeKeyForUI(w.p.LocalLeader))

	w.settingsInfo.SetText(strings.Join(lines, "\n"))
}

func encodeKeyForUI(key string) string {
	if key == "" {
		return "(empty)"
	}
	if key == " " {
		return "(space)"
	}
	if key == "\t" {
		return "(tab)"
	}
	if key == "\n" {
		return "(enter)"
	}
	return key
}
