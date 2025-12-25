package ui

import (
	"sort"
	"strings"

	"github.com/rivo/tview"

	"nvimwiz/internal/profile"
)

func attachSettingsHelp(w *Wizard, form *tview.Form, key string) {
	if w == nil || form == nil {
		return
	}
	idx := form.GetFormItemCount() - 1
	if idx < 0 {
		return
	}
	item := form.GetFormItem(idx)
	setFocusHelp(item, func() {
		w.showSettingsFieldHelp(key)
	})
}

func (w *Wizard) pageSettings() tview.Primitive {
	fieldWidth := 28

	if w.settingsInfo == nil {
		w.settingsInfo = tview.NewTextView().SetDynamicColors(true)
		w.settingsInfo.SetBorder(true)
		w.settingsInfo.SetTitle("Info")
	}

	initializing := true

	fields := tview.NewForm()
	fields.SetBorder(true)
	fields.SetTitle("Settings")
	fields.SetButtonsAlign(tview.AlignCenter)

	selectedProfile := strings.TrimSpace(w.p.Name)
	if selectedProfile == "" {
		selectedProfile = "default"
	}

	profileNames, err := profile.ListProfiles()
	if err != nil {
		w.settingsInfo.SetText(err.Error())
		return fields
	}

	profileIndex := 0
	for i, name := range profileNames {
		if name == selectedProfile {
			profileIndex = i
			break
		}
	}

	fields.AddDropDown("Profile", profileNames, profileIndex, func(_ string, index int) {
		if initializing {
			return
		}
		if index < 0 || index >= len(profileNames) {
			return
		}
		name := profileNames[index]

		p, _, loadErr := profile.LoadByName(name, w.cat)
		if loadErr != nil {
			w.settingsInfo.SetText(loadErr.Error())
			return
		}
		p.Name = name
		w.p = p
		_ = profile.SetCurrent(name)

		w.pages.RemovePage("features")
		w.pages.AddPage("features", w.pageFeatures(), true, false)
		w.pages.RemovePage("settings")
		w.pages.AddPage("settings", w.pageSettings(), true, false)
		w.pages.SwitchToPage("settings")
		w.updateSettingsInfo()
	})
	attachSettingsHelp(w, fields, "profile")

	targetLabels := []string{"safe build (recommended)", "system config"}
	targetIndex := 0
	if strings.ToLower(strings.TrimSpace(w.p.Target)) == "default" {
		targetIndex = 1
	}
	fields.AddDropDown("Target", targetLabels, targetIndex, func(_ string, index int) {
		if initializing {
			return
		}
		if index == 1 {
			w.p.Target = "default"
		} else {
			w.p.Target = "safe"
		}
		w.p.Normalize(w.cat)
		_ = profile.Save(w.p)
		w.pages.RemovePage("settings")
		w.pages.AddPage("settings", w.pageSettings(), true, false)
		w.pages.SwitchToPage("settings")
		w.updateSettingsInfo()
	})
	attachSettingsHelp(w, fields, "target")

	if strings.ToLower(strings.TrimSpace(w.p.Target)) == "safe" {
		fields.AddInputField("Build name", w.p.AppName, fieldWidth, nil, func(text string) {
			w.p.AppName = strings.TrimSpace(text)
			w.p.Normalize(w.cat)
			_ = profile.Save(w.p)
			w.showSettingsFieldHelp("build_name")
		})
		attachSettingsHelp(w, fields, "build_name")
	}

	presetLabels := []string{}
	presetIDForLabel := map[string]string{}

	for id, preset := range w.cat.Presets {
		label := strings.TrimSpace(preset.Title)
		if label == "" {
			label = id
		}
		presetLabels = append(presetLabels, label)
		presetIDForLabel[label] = id
	}
	sort.Strings(presetLabels)

	presetIndex := 0
	currentPreset := strings.TrimSpace(w.p.Preset)
	if currentPreset == "" {
		currentPreset = "lazyvim"
	}
	for i, label := range presetLabels {
		if presetIDForLabel[label] == currentPreset {
			presetIndex = i
			break
		}
	}

	fields.AddDropDown("Preset", presetLabels, presetIndex, func(option string, _ int) {
		if initializing {
			return
		}
		presetID := presetIDForLabel[option]
		if strings.TrimSpace(presetID) == "" {
			presetID = "lazyvim"
		}
		w.p.Preset = presetID
		w.p.Normalize(w.cat)
		w.applyPreset(presetID)
		_ = profile.Save(w.p)

		w.pages.RemovePage("features")
		w.pages.AddPage("features", w.pageFeatures(), true, false)

		w.showSettingsFieldHelp("preset")
	})
	attachSettingsHelp(w, fields, "preset")

	if strings.ToLower(strings.TrimSpace(w.p.Target)) == "default" {
		configModeLabels := []string{"managed", "integrate"}
		configModeIndex := 0
		if strings.ToLower(strings.TrimSpace(w.p.ConfigMode)) == "integrate" {
			configModeIndex = 1
		}
		fields.AddDropDown("Config mode", configModeLabels, configModeIndex, func(_ string, index int) {
			if initializing {
				return
			}
			if index == 1 {
				w.p.ConfigMode = "integrate"
			} else {
				w.p.ConfigMode = "managed"
			}
			w.p.Normalize(w.cat)
			_ = profile.Save(w.p)
			w.showSettingsFieldHelp("config_mode")
		})
		attachSettingsHelp(w, fields, "config_mode")
	}

	fields.AddInputField("Projects dir", w.p.ProjectsDir, fieldWidth, nil, func(text string) {
		w.p.ProjectsDir = strings.TrimSpace(text)
		w.p.Normalize(w.cat)
		_ = profile.Save(w.p)
		w.showSettingsFieldHelp("projects_dir")
	})
	attachSettingsHelp(w, fields, "projects_dir")

	fields.AddInputField("Leader", w.p.Leader, fieldWidth, nil, func(text string) {
		w.p.Leader = text
		w.p.Normalize(w.cat)
		_ = profile.Save(w.p)
		w.showSettingsFieldHelp("leader")
	})
	attachSettingsHelp(w, fields, "leader")

	fields.AddInputField("Local leader", w.p.LocalLeader, fieldWidth, nil, func(text string) {
		w.p.LocalLeader = text
		w.p.Normalize(w.cat)
		_ = profile.Save(w.p)
		w.showSettingsFieldHelp("local_leader")
	})
	attachSettingsHelp(w, fields, "local_leader")

	verifyLabels := []string{"auto", "require", "off"}
	verifyIndex := 0
	for i, v := range verifyLabels {
		if v == w.p.Verify {
			verifyIndex = i
			break
		}
	}
	fields.AddDropDown("Verify", verifyLabels, verifyIndex, func(_ string, index int) {
		if initializing {
			return
		}
		if index >= 0 && index < len(verifyLabels) {
			w.p.Verify = verifyLabels[index]
		}
		w.p.Normalize(w.cat)
		_ = profile.Save(w.p)
		w.showSettingsFieldHelp("verify")
	})
	attachSettingsHelp(w, fields, "verify")

	buttons := tview.NewForm()
	buttons.AddButton("Back", func() { w.gotoPage("welcome") })
	buttons.AddButton("Save", func() {
		w.p.Normalize(w.cat)
		_ = profile.Save(w.p)
		w.updateSettingsInfo()
	})
	buttons.AddButton("Profiles", func() {
		w.openProfilesManager()
	})
	buttons.AddButton("Next", func() {
		w.p.Normalize(w.cat)
		_ = profile.Save(w.p)
		w.gotoPage("features")
	})
	buttons.SetButtonsAlign(tview.AlignCenter)

	w.updateSettingsInfo()
	initializing = false

	left := tview.NewFlex().SetDirection(tview.FlexRow)
	left.AddItem(fields, 0, 1, true)
	left.AddItem(buttons, 3, 0, false)

	right := tview.NewFlex().SetDirection(tview.FlexRow)
	right.AddItem(w.settingsInfo, 0, 2, false)
	right.AddItem(w.systemInfoView(), 0, 1, false)

	main := tview.NewFlex()
	main.AddItem(left, 0, 2, true)
	main.AddItem(right, 0, 3, false)
	return main
}
