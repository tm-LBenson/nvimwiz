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
	form.AddDropDown("Profile", profileNames, profileIndex, func(_ string, index int) {
		if profileInit {
			return
		}
		if index < 0 || index >= len(profileNames) {
			return
		}
		selected := profileNames[index]
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
	profileInit = false

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
		w.updateSettingsInfo()
	})

	form.AddInputField("Build name", w.p.AppName, fieldWidth, nil, func(text string) {
		w.p.AppName = strings.TrimSpace(text)
		w.p.Normalize(w.cat)
		_ = profile.Save(w.p)
		w.updateSettingsInfo()
	})

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
	form.AddDropDown("Preset", presetLabels, presetIndex, func(_ string, index int) {
		if index < 0 || index >= len(presetIDs) {
			return
		}
		id := presetIDs[index]
		w.applyPreset(id)
		w.updateSettingsInfo()
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
		w.updateSettingsInfo()
	})

	form.AddInputField("Projects dir", w.p.ProjectsDir, fieldWidth, nil, func(text string) {
		w.p.ProjectsDir = strings.TrimSpace(text)
		w.p.Normalize(w.cat)
		_ = profile.Save(w.p)
		w.updateSettingsInfo()
	})

	form.AddInputField("Leader", w.p.Leader, fieldWidth, nil, func(text string) {
		w.p.Leader = text
		w.p.Normalize(w.cat)
		_ = profile.Save(w.p)
		w.updateSettingsInfo()
	})
	form.AddInputField("Local leader", w.p.LocalLeader, fieldWidth, nil, func(text string) {
		w.p.LocalLeader = text
		w.p.Normalize(w.cat)
		_ = profile.Save(w.p)
		w.updateSettingsInfo()
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
		w.updateSettingsInfo()
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

	state, _ := profile.LoadState()
	profileName := strings.TrimSpace(state.Current)
	if profileName == "" {
		profileName = "default"
	}

	lines := []string{}
	lines = append(lines, "Profile: "+profileName)
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
