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

	form.AddDropDown("Preset", presetLabels, presetIndex, func(option string, index int) {
		if index < 0 || index >= len(presetIDs) {
			return
		}
		w.applyPreset(presetIDs[index])
		w.updateSettingsInfo()
	})

	modeLabels := []string{"managed", "integrate"}
	modeIndex := 0
	if w.p.ConfigMode == "integrate" {
		modeIndex = 1
	}
	form.AddDropDown("Config mode", modeLabels, modeIndex, func(option string, index int) {
		if index == 1 {
			w.p.ConfigMode = "integrate"
		} else {
			w.p.ConfigMode = "managed"
		}
		w.updateSettingsInfo()
	})

	form.AddInputField("Projects dir", w.p.ProjectsDir, 0, nil, func(text string) {
		w.p.ProjectsDir = strings.TrimSpace(text)
		w.updateSettingsInfo()
	})

	form.AddInputField("Leader", w.p.Leader, 5, nil, func(text string) {
		w.p.Leader = text
		w.updateSettingsInfo()
	})
	form.AddInputField("Local leader", w.p.LocalLeader, 5, nil, func(text string) {
		w.p.LocalLeader = text
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
	form.AddDropDown("Verify", verifyLabels, verifyIndex, func(option string, index int) {
		if index >= 0 && index < len(verifyLabels) {
			w.p.Verify = verifyLabels[index]
		}
		w.updateSettingsInfo()
	})

	form.AddButton("Back", func() { w.gotoPage("welcome") })
	form.AddButton("Save", func() { _ = profile.Save(w.p) })
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

func (w *Wizard) systemInfoView() tview.Primitive {
	tv := tview.NewTextView()
	tv.SetDynamicColors(true)
	tv.SetBorder(true)
	tv.SetTitle("System")
	lines := []string{}
	if w.envNote != "" {
		lines = append(lines, w.envNote)
	}
	if w.sys.PrettyName != "" {
		lines = append(lines, "OS: "+w.sys.PrettyName)
	} else {
		lines = append(lines, "OS: "+w.sys.GOOS)
	}
	lines = append(lines, "Arch: "+w.sys.GOARCH)
	if w.sys.WSL {
		lines = append(lines, "WSL: yes")
	} else {
		lines = append(lines, "WSL: no")
	}
	if len(w.sys.PackageManagers) > 0 {
		lines = append(lines, "Package managers: "+strings.Join(w.sys.PackageManagers, ", "))
	}
	tnames := []string{"nvim", "rg", "fd", "git", "curl", "tar", "sha256sum"}
	for _, k := range tnames {
		ti := w.sys.Tools[k]
		if ti.Present {
			v := ti.Path
			if ti.Version != "" {
				v = ti.Version
			}
			lines = append(lines, k+": "+v)
		} else {
			lines = append(lines, k+": missing")
		}
	}
	tv.SetText(strings.Join(lines, "\n"))
	return tv
}

func (w *Wizard) updateSettingsInfo() {
	if w.settingsInfo == nil {
		return
	}
	pr, ok := w.cat.Presets[w.p.Preset]
	presetLine := w.p.Preset
	short := ""
	if ok {
		presetLine = pr.Title
		short = pr.Short
	}
	lines := []string{
		"Preset: " + presetLine,
	}
	if short != "" {
		lines = append(lines, short)
	}
	lines = append(lines, "")
	lines = append(lines, "Config mode: "+w.p.ConfigMode)
	if w.p.ConfigMode == "integrate" {
		lines = append(lines, "Integrate: require nvimwiz.loader from your init.lua")
	} else {
		lines = append(lines, "Managed: nvimwiz writes ~/.config/nvim/init.lua")
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

func (w *Wizard) applyPreset(id string) {
	pr, ok := w.cat.Presets[id]
	if !ok {
		return
	}
	w.p.Preset = id
	for fid, v := range pr.Features {
		w.p.Features[fid] = v
	}
	for ck, cv := range pr.Choices {
		w.p.Choices[ck] = cv
	}
	w.p.Normalize(w.cat)
	_ = profile.Save(w.p)
}
