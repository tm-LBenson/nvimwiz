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

func setFocusHelp(p any, fn func()) {
	type focusSetter interface {
		SetFocusFunc(func()) *tview.Box
	}
	if f, ok := p.(focusSetter); ok {
		f.SetFocusFunc(fn)
	}
}

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
	fields := tview.NewForm()
	fields.SetBorder(true)
	fields.SetTitle("Settings")

	w.settingsInfo = tview.NewTextView()
	w.settingsInfo.SetDynamicColors(true)
	w.settingsInfo.SetBorder(true)
	w.settingsInfo.SetTitle("Info")

	setFormLabelWidth(fields, 30)
	fieldWidth := 24

	profileNames, err := profile.ListProfiles()
	if err != nil || len(profileNames) == 0 {
		profileNames = []string{"default"}
	}

	state, _ := profile.LoadState()
	currentName := strings.TrimSpace(state.Current)
	if currentName == "" {
		currentName = "default"
	}

	profileIndex := 0
	for i, name := range profileNames {
		if name == currentName {
			profileIndex = i
			break
		}
	}

	profileInit := true
	fields.AddDropDown("Profile", profileNames, profileIndex, func(_ string, index int) {
		if profileInit {
			return
		}
		if index < 0 || index >= len(profileNames) {
			return
		}
		selected := profileNames[index]
		if strings.TrimSpace(selected) == "" {
			return
		}
		if selected == currentName {
			return
		}

		p, _, loadErr := profile.LoadByName(selected, w.cat)
		if loadErr != nil {
			return
		}
		w.p = p
		_ = profile.SetCurrent(selected)

		w.app.QueueUpdateDraw(func() {
			w.pages.RemovePage("settings")
			w.pages.AddPage("settings", w.pageSettings(), true, false)
			w.pages.SwitchToPage("settings")
			w.app.SetFocus(w.pages)
		})
	})
	attachSettingsHelp(w, fields, "profile")
	profileInit = false

	targetLabels := []string{"default config", "safe build"}
	targetIndex := 0
	if strings.ToLower(strings.TrimSpace(w.p.Target)) == "safe" {
		targetIndex = 1
	}
	fields.AddDropDown("Target", targetLabels, targetIndex, func(_ string, index int) {
		if index == 1 {
			w.p.Target = "safe"
		} else {
			w.p.Target = "default"
		}
		w.p.Normalize(w.cat)
		_ = profile.Save(w.p)
		w.showSettingsFieldHelp("target")
	})
	attachSettingsHelp(w, fields, "target")

	fields.AddInputField("Build name", w.p.AppName, fieldWidth, nil, func(text string) {
		w.p.AppName = strings.TrimSpace(text)
		w.p.Normalize(w.cat)
		_ = profile.Save(w.p)
		w.showSettingsFieldHelp("build_name")
	})
	attachSettingsHelp(w, fields, "build_name")

	presetIDs := make([]string, 0, len(w.cat.Presets))
	for id := range w.cat.Presets {
		presetIDs = append(presetIDs, id)
	}
	sort.Strings(presetIDs)

	presetLabels := make([]string, 0, len(presetIDs))
	presetIndex := 0
	for i, id := range presetIDs {
		preset := w.cat.Presets[id]
		presetLabels = append(presetLabels, preset.Title)
		if id == w.p.Preset {
			presetIndex = i
		}
	}
	fields.AddDropDown("Preset", presetLabels, presetIndex, func(_ string, index int) {
		if index < 0 || index >= len(presetIDs) {
			return
		}
		id := presetIDs[index]
		w.applyPreset(id)
		w.showSettingsFieldHelp("preset")
	})
	attachSettingsHelp(w, fields, "preset")

	modeLabels := []string{"managed", "integrate"}
	modeIndex := 0
	if w.p.ConfigMode == "integrate" {
		modeIndex = 1
	}
	fields.AddDropDown("Config mode", modeLabels, modeIndex, func(_ string, index int) {
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
		_ = profile.Save(w.p)
		w.updateSettingsInfo()
	})
	buttons.AddButton("Next", func() {
		w.p.Normalize(w.cat)
		_ = profile.Save(w.p)
		w.gotoPage("features")
	})
	buttons.SetButtonsAlign(tview.AlignCenter)

	w.updateSettingsInfo()

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
