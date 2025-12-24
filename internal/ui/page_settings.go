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

	presetIDs := make([]string, 0, len(w.cat.Presets))
	for id := range w.cat.Presets {
		presetIDs = append(presetIDs, id)
	}
	sort.Strings(presetIDs)

	presetLabels := make([]string, 0, len(presetIDs))
	presetIndex := 0
	for i, id := range presetIDs {
		pr := w.cat.Presets[id]
		presetLabels = append(presetLabels, pr.Title)
		if id == w.p.Preset {
			presetIndex = i
		}
	}

	targetLabels := []string{"safe build (recommended)", "system config (~/.config/nvim)"}
	targetIndex := 0
	if strings.ToLower(strings.TrimSpace(w.p.Target)) == "default" {
		targetIndex = 1
	}
	form.AddDropDown("Target", targetLabels, targetIndex, func(_ string, index int) {
		desired := "safe"
		if index == 1 {
			desired = "default"
		}
		cur := strings.ToLower(strings.TrimSpace(w.p.Target))
		if desired != cur {
			w.p.Target = desired
			w.p.Normalize(w.cat)
			_ = profile.Save(w.p)

			w.pages.RemovePage("settings")
			w.pages.AddPage("settings", w.pageSettings(), true, true)
			w.gotoPage("settings")
			return
		}
		w.updateSettingsInfo()
	})

	fieldWidth := 24
	if strings.ToLower(strings.TrimSpace(w.p.Target)) == "safe" {
		form.AddInputField("Build name", w.p.AppName, fieldWidth, nil, func(text string) {
			w.p.AppName = strings.TrimSpace(text)
			w.p.Normalize(w.cat)
			_ = profile.Save(w.p)
			w.updateSettingsInfo()
		})
	}

	form.AddDropDown("Preset", presetLabels, presetIndex, func(_ string, index int) {
		if index < 0 || index >= len(presetIDs) {
			return
		}
		w.applyPreset(presetIDs[index])
		w.updateSettingsInfo()
	})

	if strings.ToLower(strings.TrimSpace(w.p.Target)) == "default" {
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
	}

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
			w.p.Normalize(w.cat)
			_ = profile.Save(w.p)
		}
		w.updateSettingsInfo()
	})

	form.AddButton("Back", func() { w.gotoPage("welcome") })
	form.AddButton("Save", func() {
		w.p.Normalize(w.cat)
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

	leftRight := tview.NewFlex()
	leftRight.AddItem(form, 0, 2, true)
	leftRight.AddItem(w.settingsInfo, 0, 3, false)

	wrap := tview.NewFlex().SetDirection(tview.FlexRow)
	wrap.AddItem(leftRight, 0, 1, true)
	wrap.AddItem(w.systemInfoView(), 12, 0, false)
	return wrap
}
