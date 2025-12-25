package ui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/rivo/tview"

	"nvimwiz/internal/nvimcfg"
	"nvimwiz/internal/profile"
)


func (w *Wizard) openProfilesManager() {
	const pageName = "profiles_manager"

	if w.pages.HasPage(pageName) {
		w.pages.ShowPage(pageName)
		return
	}

	view := w.buildProfilesManagerView()
	w.pages.AddPage(pageName, view, true, true)
	w.app.SetFocus(view)
}

func (w *Wizard) closeProfilesManager() {
	w.pages.RemovePage("profiles_manager")
	w.pages.RemovePage("profiles_manager_msg")
	w.pages.RemovePage("profiles_manager_confirm")
	w.pages.RemovePage("profiles_manager_prompt")
}

func (w *Wizard) buildProfilesManagerView() tview.Primitive {
	list := tview.NewList()
	list.SetBorder(true)
	list.SetTitle("Profiles")
	list.ShowSecondaryText(false)

	detail := tview.NewTextView()
	detail.SetDynamicColors(true)
	detail.SetBorder(true)
	detail.SetTitle("Details")

	showMessage := func(title, text string) {
		modal := tview.NewModal().SetText(text).AddButtons([]string{"OK"})
		modal.SetTitle(title)
		modal.SetDoneFunc(func(_ int, _ string) {
			w.pages.RemovePage("profiles_manager_msg")
			w.app.SetFocus(list)
		})
		w.pages.AddPage("profiles_manager_msg", centered(modal, 60, 12), true, true)
		w.app.SetFocus(modal)
	}

	showConfirm := func(title, text string, onYes func()) {
		modal := tview.NewModal().SetText(text).AddButtons([]string{"Cancel", "OK"})
		modal.SetTitle(title)
		modal.SetDoneFunc(func(_ int, button string) {
			w.pages.RemovePage("profiles_manager_confirm")
			w.app.SetFocus(list)
			if button == "OK" {
				onYes()
			}
		})
		w.pages.AddPage("profiles_manager_confirm", centered(modal, 70, 12), true, true)
		w.app.SetFocus(modal)
	}

	showPrompt := func(title, label, initial string, onOK func(string)) {
		form := tview.NewForm()
		form.SetBorder(true)
		form.SetTitle(title)
		form.AddInputField(label, initial, 28, nil, nil)
		form.AddButton("Cancel", func() {
			w.pages.RemovePage("profiles_manager_prompt")
			w.app.SetFocus(list)
		})
		form.AddButton("OK", func() {
			field := form.GetFormItem(0).(*tview.InputField)
			text := field.GetText()
			w.pages.RemovePage("profiles_manager_prompt")
			w.app.SetFocus(list)
			onOK(text)
		})
		form.SetButtonsAlign(tview.AlignCenter)
		w.pages.AddPage("profiles_manager_prompt", centered(form, 70, 14), true, true)
		w.app.SetFocus(form)
	}

	rebuildAfterSwitch := func() {
		// Settings and features pages are built from w.p, so rebuild them.
		w.pages.RemovePage("settings")
		w.pages.AddPage("settings", w.pageSettings(), true, false)

		w.pages.RemovePage("features")
		w.pages.AddPage("features", w.pageFeatures(), true, false)
	}

	renderDetail := func(profileName string) {
		p, _, err := profile.LoadByName(profileName, w.cat)
		if err != nil {
			detail.SetText(err.Error())
			return
		}
		lines := []string{}
		lines = append(lines, "Name: "+profileName)
		lines = append(lines, "Preset: "+p.Preset)
		lines = append(lines, "Target: "+p.Target)
		lines = append(lines, "Build name: "+strings.TrimSpace(p.AppName))
		lines = append(lines, "Effective app name: "+p.EffectiveAppName())

		cfgDir, err := nvimcfg.ConfigDirForProfile(p)
		if err == nil && cfgDir != "" {
			lines = append(lines, "Config dir: "+cfgDir)
		}

		lines = append(lines, "")
		if strings.ToLower(strings.TrimSpace(p.Target)) == "default" {
			lines = append(lines, "Launch:")
			lines = append(lines, "  nvim")
		} else {
			lines = append(lines, "Launch:")
			lines = append(lines, "  NVIM_APPNAME="+p.EffectiveAppName()+" nvim")
		}

		detail.SetText(strings.Join(lines, "\n"))
	}

	items := []string{}
	reload := func(selectName string) {
		names, err := profile.ListProfiles()
		if err != nil {
			showMessage("Profiles", err.Error())
			return
		}

		st, _ := profile.LoadState()
		current := strings.TrimSpace(st.Current)
		if current == "" {
			current = "default"
		}

		items = names
		list.Clear()

		sort.SliceStable(items, func(i, j int) bool {
			if items[i] == current {
				return true
			}
			if items[j] == current {
				return false
			}
			return items[i] < items[j]
		})

		selectedIndex := 0
		for idx, name := range items {
			label := name
			if name == current {
				label += "  (current)"
			}
			list.AddItem(label, "", 0, nil)
			if selectName != "" && name == selectName {
				selectedIndex = idx
			}
		}

		if len(items) == 0 {
			detail.SetText("No profiles")
			return
		}
		list.SetCurrentItem(selectedIndex)
		renderDetail(items[selectedIndex])
	}

	switchTo := func(profileName string) {
		p, _, err := profile.LoadByName(profileName, w.cat)
		if err != nil {
			showMessage("Switch profile", err.Error())
			return
		}
		w.p = p
		if err := profile.SetCurrent(profileName); err != nil {
			showMessage("Switch profile", err.Error())
			return
		}

		rebuildAfterSwitch()
		w.closeProfilesManager()
		w.pages.SwitchToPage("settings")
		w.app.SetFocus(w.pages)
	}

	list.SetChangedFunc(func(index int, _ string, _ string, _ rune) {
		if index < 0 || index >= len(items) {
			return
		}
		renderDetail(items[index])
	})

	buttons := tview.NewForm()
	buttons.SetButtonsAlign(tview.AlignCenter)

	buttons.AddButton("Set active", func() {
		idx := list.GetCurrentItem()
		if idx < 0 || idx >= len(items) {
			return
		}
		switchTo(items[idx])
	})

	buttons.AddButton("New", func() {
		showPrompt("New profile", "Name", "", func(name string) {
			name = strings.TrimSpace(name)
			if name == "" {
				showMessage("New profile", "Name cannot be empty")
				return
			}

			if ok, _ := profile.Exists(name); ok {
				showMessage("New profile", "Profile already exists: "+name)
				return
			}

			p := w.p
			p.Name = name
			p.Target = "safe"
			p.AppName = ""
			p.Normalize(w.cat)

			if err := profile.SaveAs(name, p); err != nil {
				showMessage("New profile", err.Error())
				return
			}
			if err := profile.SetCurrent(name); err != nil {
				showMessage("New profile", err.Error())
				return
			}
			w.p = p
			rebuildAfterSwitch()
			reload(name)
		})
	})

	buttons.AddButton("Clone", func() {
		idx := list.GetCurrentItem()
		if idx < 0 || idx >= len(items) {
			return
		}
		src := items[idx]
		showPrompt("Clone profile", "New name", src+"-copy", func(dst string) {
			dst = strings.TrimSpace(dst)
			if dst == "" {
				showMessage("Clone profile", "Name cannot be empty")
				return
			}
			if err := profile.Clone(src, dst, w.cat); err != nil {
				showMessage("Clone profile", err.Error())
				return
			}
			reload(dst)
		})
	})

	buttons.AddButton("Rename", func() {
		idx := list.GetCurrentItem()
		if idx < 0 || idx >= len(items) {
			return
		}
		oldName := items[idx]
		if oldName == "default" {
			showMessage("Rename profile", "The default profile cannot be renamed")
			return
		}
		showPrompt("Rename profile", "New name", oldName, func(newName string) {
			newName = strings.TrimSpace(newName)
			if newName == "" {
				showMessage("Rename profile", "Name cannot be empty")
				return
			}
			if err := profile.Rename(oldName, newName, w.cat); err != nil {
				showMessage("Rename profile", err.Error())
				return
			}

			// If we renamed the active profile, we should reload it into the wizard.
			st, _ := profile.LoadState()
			if strings.TrimSpace(st.Current) == newName {
				p, _, err := profile.LoadByName(newName, w.cat)
				if err == nil {
					w.p = p
					rebuildAfterSwitch()
				}
			}
			reload(newName)
		})
	})

	buttons.AddButton("Delete", func() {
		idx := list.GetCurrentItem()
		if idx < 0 || idx >= len(items) {
			return
		}
		name := items[idx]
		if name == "default" {
			showMessage("Delete profile", "The default profile cannot be deleted")
			return
		}
		showConfirm("Delete profile", fmt.Sprintf("Delete profile %q?", name), func() {
			if err := profile.Delete(name); err != nil {
				showMessage("Delete profile", err.Error())
				return
			}

		_, p, _, err := profile.LoadCurrent(w.cat)
		if err == nil {
			w.p = p
			rebuildAfterSwitch()
		}
			reload("")
		})
	})

	buttons.AddButton("Close", func() {
		w.closeProfilesManager()
		w.app.SetFocus(w.pages)
	})

	flex := tview.NewFlex()
	flex.AddItem(list, 0, 1, true)
	flex.AddItem(detail, 0, 2, false)

	wrap := tview.NewFlex().SetDirection(tview.FlexRow)
	wrap.SetBorder(true)
	wrap.SetTitle("Profile manager")
	wrap.AddItem(flex, 0, 1, true)
	wrap.AddItem(buttons, 3, 0, false)

	// First load.
	reload("")

	return wrap
}

// centered wraps a primitive with flexible padding so it appears centered.
func centered(p tview.Primitive, width, height int) tview.Primitive {
	row := tview.NewFlex()
	row.AddItem(nil, 0, 1, false)
	row.AddItem(p, width, 0, true)
	row.AddItem(nil, 0, 1, false)

	col := tview.NewFlex().SetDirection(tview.FlexRow)
	col.AddItem(nil, 0, 1, false)
	col.AddItem(row, height, 0, true)
	col.AddItem(nil, 0, 1, false)
	return col
}
