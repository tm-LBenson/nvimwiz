package ui

import (
	"strings"

	"nvimwiz/internal/profile"
)

func (w *Wizard) updateSettingsInfo() {
	if w.settingsInfo == nil {
		return
	}

	w.settingsInfo.SetTitle("Info")

	profileName := "default"
	if st, err := profile.LoadState(); err == nil {
		if strings.TrimSpace(st.Current) != "" {
			profileName = strings.TrimSpace(st.Current)
		}
	}

	presetTitle := w.p.Preset
	presetShort := ""
	if pr, ok := w.cat.Presets[w.p.Preset]; ok {
		presetTitle = pr.Title
		presetShort = pr.Short
	}

	lines := []string{}
	lines = append(lines, "Profile: "+profileName)
	lines = append(lines, "Preset: "+presetTitle)
	if strings.TrimSpace(presetShort) != "" {
		lines = append(lines, presetShort)
	}

	lines = append(lines, "")
	lines = append(lines, "Tip: select a setting to see help")

	lines = append(lines, "")
	lines = append(lines, "Config mode: "+w.p.ConfigMode)
	if strings.ToLower(strings.TrimSpace(w.p.ConfigMode)) == "integrate" {
		lines = append(lines, "Integrate means you must require nvimwiz.loader from your own init.lua")
	} else {
		lines = append(lines, "Managed means nvimwiz writes ~/.config/nvim/init.lua")
	}

	lines = append(lines, "")
	lines = append(lines, "Verify downloads: "+w.p.Verify)

	lines = append(lines, "")
	lines = append(lines, "Projects dir: "+w.p.ProjectsDir)
	lines = append(lines, "Leader: "+encodeKeyForUI(w.p.Leader))
	lines = append(lines, "Local leader: "+encodeKeyForUI(w.p.LocalLeader))

	target := strings.ToLower(strings.TrimSpace(w.p.Target))
	lines = append(lines, "")
	lines = append(lines, "Target: "+target)

	buildName := strings.TrimSpace(w.p.AppName)
	if buildName == "" {
		buildName = "(empty)"
	}
	lines = append(lines, "Build name: "+buildName)
	lines = append(lines, "Effective app name: "+w.p.EffectiveAppName())

	lines = append(lines, "")
	lines = append(lines, "Launch:")
	if target == "safe" {
		lines = append(lines, "  NVIM_APPNAME="+w.p.EffectiveAppName()+" nvim")
	} else {
		lines = append(lines, "  nvim")
	}

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
