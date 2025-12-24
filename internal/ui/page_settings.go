package ui

import (
	"sort"
	"strings"

	"github.com/rivo/tview"

	"nvimwiz/internal/profile"
)

func setFormLabelWidth(form *tview.Form, width int) {
	type labelWidther interface {
		SetLabelWidth(int) *tview.Form
	}
	if f, ok := interface{}(form).(labelWidther); ok {
		f.SetLabelWidth(width)
	}
}

func (w *Wizard) pageSettings() tview.Primitive {
	form := tview.NewForm()
	form.SetBorder(true)
	form.SetTitle("Settings")

	w.settingsInfo = tview.NewTextView()
	w.settingsInfo.SetDynamicColors(true)
	w.settingsInfo.SetBorder(true)
	w.settingsInfo.SetTitle("Info")

	setFormLabelWidth(form, 30)

	fieldWidth := 24

	targetLabels := []string{"default config", "safe build"}
	targetIndex := 0
	if strings.ToLower(strings.TrimSpace(w.p.Target)) == "safe" {
		targetIndex = 1
	}
	form.AddDropDown("Target", targetLabels, targetIndex, func(_ string, index int) {
		if index == 1 {
			w.p.Target = "safe"
		} else {
			w.p.Target = "default"
		}
		w.p.Normalize(w.cat)
		_ = profile.Save(w.p)
		w.showSettingsFieldHelp("target")
	})

	form.AddInputField("Build name", w.p.AppName, fieldWidth, nil, func(text string) {
		w.p.AppName = strings.TrimSpace(text)
		w.p.Normalize(w.cat)
		_ = profile.Save(w.p)
		w.showSettingsFieldHelp("build_name")
	})

	presetIDs := make([]string, 0, len(w.cat.Presets))
	for presetID := range w.cat.Presets {
		presetIDs = append(presetIDs, presetID)
	}
	sort.Strings(presetIDs)

	presetLabels := make([]string, 0, len(presetIDs))
	presetIndex := 0
	for i, presetID := range presetIDs {
		preset := w.cat.Presets[presetID]
		presetLabels = append(presetLabels, preset.Title)
		if presetID == w.p.Preset {
			presetIndex = i
		}
	}

	form.AddDropDown("Preset", presetLabels, presetIndex, func(_ string, index int) {
		if index < 0 || index >= len(presetIDs) {
			return
		}
		w.applyPreset(presetIDs[index])
		w.showSettingsFieldHelp("preset")
	})

	modeLabels := []string{"managed", "integrate"}
	modeIndex := 0
	if w.p.ConfigMode == "integrate" {
		modeIndex = 1
	}
	form.AddDropDown("Config mode", modeLabels, modeIndex, func(_ string, index int) {
		if index == 1 {
			w.p.ConfigMode = "integrate"
		} else {
			w.p.ConfigMode = "managed"
		}
		w.p.Normalize(w.cat)
		_ = profile.Save(w.p)
		w.showSettingsFieldHelp("config_mode")
	})

	form.AddInputField("Projects dir", w.p.ProjectsDir, fieldWidth, nil, func(text string) {
		w.p.ProjectsDir = strings.TrimSpace(text)
		w.p.Normalize(w.cat)
		_ = profile.Save(w.p)
		w.showSettingsFieldHelp("projects_dir")
	})

	form.AddInputField("Leader", w.p.Leader, fieldWidth, nil, func(text string) {
		w.p.Leader = text
		w.p.Normalize(w.cat)
		_ = profile.Save(w.p)
		w.showSettingsFieldHelp("leader")
	})

	form.AddInputField("Local leader", w.p.LocalLeader, fieldWidth, nil, func(text string) {
		w.p.LocalLeader = text
		w.p.Normalize(w.cat)
		_ = profile.Save(w.p)
		w.showSettingsFieldHelp("local_leader")
	})

	verifyLabels := []string{"auto", "require", "off"}
	verifyIndex := 0
	for i, v := range verifyLabels {
		if v == w.p.Verify {
			verifyIndex = i
			break
		}
	}
	form.AddDropDown("Verify", verifyLabels, verifyIndex, func(_ string, index int) {
		if index >= 0 && index < len(verifyLabels) {
			w.p.Verify = verifyLabels[index]
		}
		w.p.Normalize(w.cat)
		_ = profile.Save(w.p)
		w.showSettingsFieldHelp("verify")
	})

	form.AddButton("Back", func() { w.gotoPage("welcome") })
	form.AddButton("Save", func() {
		_ = profile.Save(w.p)
		w.updateSettingsInfo()
	})
	form.AddButton("Next", func() {
		w.p.Normalize(w.cat)
		_ = profile.Save(w.p)
		w.gotoPage("features")
	})
	form.SetButtonsAlign(tview.AlignCenter)

	w.updateSettingsInfo()

	flex := tview.NewFlex()
	flex.AddItem(form, 0, 2, true)
	flex.AddItem(w.settingsInfo, 0, 3, false)

	wrap := tview.NewFlex().SetDirection(tview.FlexRow)
	wrap.AddItem(flex, 0, 1, true)
	wrap.AddItem(w.systemInfoView(), 12, 0, false)
	return wrap
}

func (w *Wizard) showSettingsFieldHelp(fieldKey string) {
	if w.settingsInfo == nil {
		return
	}

	target := strings.ToLower(strings.TrimSpace(w.p.Target))

	lines := []string{}
	switch fieldKey {
	case "target":
		lines = append(lines, "Info: Target", "")
		lines = append(lines, "default: uses your normal ~/.config/nvim")
		lines = append(lines, "safe: uses ~/.config/<build name> and does not touch your current config")
		lines = append(lines, "", "Current: "+target)

	case "build_name":
		lines = append(lines, "Info: Build name", "")
		lines = append(lines, "Used only for Safe builds (NVIM_APPNAME).")
		lines = append(lines, "It becomes the folder under ~/.config/<name>.")
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
		lines = append(lines, "Good starting point if you are new to Neovim.")
		lines = append(lines, "", "Current: "+w.p.Preset)

	case "config_mode":
		lines = append(lines, "Info: Config mode", "")
		lines = append(lines, "managed: nvimwiz owns init.lua and generated modules")
		lines = append(lines, "integrate: you keep your init.lua and require nvimwiz.loader")
		lines = append(lines, "", "Current: "+w.p.ConfigMode)

	case "projects_dir":
		lines = append(lines, "Info: Projects dir", "")
		lines = append(lines, "Used by the Neovim start screen to list your projects.")
		lines = append(lines, "Pick the folder where you keep your repos.")
		lines = append(lines, "", "Current: "+w.p.ProjectsDir)

	case "leader":
		lines = append(lines, "Info: Leader", "")
		lines = append(lines, "Leader prefixes shortcuts in many Neovim setups.")
		lines = append(lines, "Common choice is Space.")
		lines = append(lines, "", "Current: "+encodeKeyForUI(w.p.Leader))

	case "local_leader":
		lines = append(lines, "Info: Local leader", "")
		lines = append(lines, "Local leader is used by some plugins for filetype specific shortcuts.")
		lines = append(lines, "", "Current: "+encodeKeyForUI(w.p.LocalLeader))

	case "verify":
		lines = append(lines, "Info: Verify downloads", "")
		lines = append(lines, "auto: verify if checksum is available")
		lines = append(lines, "require: fail if checksum is missing")
		lines = append(lines, "off: skip verification")
		lines = append(lines, "", "Current: "+w.p.Verify)

	default:
		w.updateSettingsInfo()
		return
	}

	lines = append(lines, "", "Press Save to return to summary.")
	w.settingsInfo.SetText(strings.Join(lines, "\n"))
}

func (w *Wizard) updateSettingsInfo() {
	if w.settingsInfo == nil {
		return
	}

	preset := w.p.Preset
	presetShort := ""
	if pr, ok := w.cat.Presets[w.p.Preset]; ok {
		preset = pr.Title
		presetShort = pr.Short
	}

	lines := []string{}
	lines = append(lines, "Preset: "+preset)
	if presetShort != "" {
		lines = append(lines, presetShort)
	}

	lines = append(lines, "")
	lines = append(lines, "Config mode: "+w.p.ConfigMode)

	lines = append(lines, "")
	lines = append(lines, "Verify downloads: "+w.p.Verify)

	lines = append(lines, "")
	lines = append(lines, "Projects dir: "+w.p.ProjectsDir)
	lines = append(lines, "Leader: "+encodeKeyForUI(w.p.Leader))
	lines = append(lines, "Local leader: "+encodeKeyForUI(w.p.LocalLeader))

	target := strings.ToLower(strings.TrimSpace(w.p.Target))
	lines = append(lines, "")
	lines = append(lines, "Target: "+target)
	lines = append(lines, "Build name: "+w.p.AppName)
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

func (w *Wizard) applyPreset(id string) {
	pr, ok := w.cat.Presets[id]
	if !ok {
		return
	}
	w.p.Preset = id
	for featureID, enabled := range pr.Features {
		w.p.Features[featureID] = enabled
	}
	for choiceKey, choiceValue := range pr.Choices {
		w.p.Choices[choiceKey] = choiceValue
	}
	w.p.Normalize(w.cat)
	_ = profile.Save(w.p)
}
