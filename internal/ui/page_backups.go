package ui

import (
	"sort"
	"strings"

	"github.com/rivo/tview"

	"nvimwiz/internal/nvimcfg"
)

func (w *Wizard) pageBackups() tview.Primitive {
	list := tview.NewList()
	list.SetBorder(true)
	list.SetTitle("Backups")

	detail := tview.NewTextView()
	detail.SetDynamicColors(true)
	detail.SetBorder(true)
	detail.SetTitle("Details")

	reload := func() []nvimcfg.Backup {
		items, err := nvimcfg.ListBackups()
		if err != nil {
			detail.SetText(err.Error())
			return []nvimcfg.Backup{}
		}
		sort.Slice(items, func(i, j int) bool { return items[i].ID > items[j].ID })
		list.Clear()
		for _, b := range items {
			label := b.ID
			if strings.TrimSpace(b.Reason) != "" {
				label += "  (" + b.Reason + ")"
			}
			list.AddItem(label, "", 0, nil)
		}
		if len(items) == 0 {
			detail.SetText("No backups yet")
		} else {
			list.SetCurrentItem(0)
			renderBackupDetail(detail, items[0])
		}
		return items
	}

	items := reload()

	list.SetChangedFunc(func(index int, mainText, _ string, _ rune) {
		if index < 0 {
			return
		}
		if len(items) == 0 {
			return
		}
		if index >= len(items) {
			return
		}
		renderBackupDetail(detail, items[index])
	})

	buttons := tview.NewForm()
	buttons.AddButton("Restore", func() {
		idx := list.GetCurrentItem()
		if idx < 0 || idx >= len(items) {
			return
		}
		b := items[idx]
	w.confirm("Restore", "Restore ~/.config/nvim from backup \""+b.ID+"\"?", func() {
			if err := nvimcfg.RestoreBackupToDefault(b.ID, func(string) {}); err != nil {
				w.message("Restore", err.Error())
				return
			}
			w.message("Restore", "Restored ~/.config/nvim from backup: "+b.ID)
			items = reload()
		})
	})
	buttons.AddButton("Refresh", func() {
		items = reload()
	})
	buttons.AddButton("Back", func() { w.gotoPage("settings") })
	buttons.SetButtonsAlign(tview.AlignCenter)

	flex := tview.NewFlex()
	flex.AddItem(list, 0, 1, true)
	flex.AddItem(detail, 0, 2, false)

	wrap := tview.NewFlex().SetDirection(tview.FlexRow)
	wrap.AddItem(flex, 0, 1, true)
	wrap.AddItem(buttons, 3, 0, true)
	return wrap
}

func renderBackupDetail(tv *tview.TextView, b nvimcfg.Backup) {
	lines := []string{}
	lines = append(lines, "ID: "+b.ID)
	if strings.TrimSpace(b.CreatedAt) != "" {
		lines = append(lines, "Created: "+b.CreatedAt)
	}
	if strings.TrimSpace(b.Reason) != "" {
		lines = append(lines, "Reason: "+b.Reason)
	}
	if strings.TrimSpace(b.Source) != "" {
		lines = append(lines, "Source: "+b.Source)
	}
	if strings.TrimSpace(b.Path) != "" {
		lines = append(lines, "Path: "+b.Path)
	}
	tv.SetText(strings.Join(lines, "\n"))
}
