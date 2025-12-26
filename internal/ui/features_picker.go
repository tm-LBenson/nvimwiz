package ui

import (
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"nvimwiz/internal/catalog"
	"nvimwiz/internal/profile"
)

func (w *Wizard) openPicker(it itemRef, tableRow int) {
	options, current := w.pickerOptions(it)
	if len(options) == 0 {
		return
	}

	list := tview.NewList()
	list.ShowSecondaryText(false)
	list.SetBorder(true)

	list.SetMainTextColor(tcell.ColorWhite)
	list.SetSelectedTextColor(tcell.ColorWhite)
	list.SetSelectedBackgroundColor(tcell.ColorBlue)
	list.SetBackgroundColor(tcell.ColorGreen)

	for idx, opt := range options {
		opt := opt
		list.AddItem(opt, "", 0, func() {
			w.applyPickerSelection(it, opt)
			w.pages.RemovePage("picker")
			w.renderFeatureTable()
			w.featureTable.Select(tableRow, featuresActionCol)
			w.app.SetFocus(w.featureTable)
		})
		if idx == current {
			list.SetCurrentItem(idx)
		}
	}

	list.SetInputCapture(func(ev *tcell.EventKey) *tcell.EventKey {
		if ev.Key() == tcell.KeyEsc {
			w.pages.RemovePage("picker")
			w.app.SetFocus(w.featureTable)
			return nil
		}
		return ev
	})

	x, y := w.pickerAnchor(tableRow)
	wid := maxInt(featuresActionColWidth+2, maxStringLen(options)+2)
	hei := len(options) + 2

	_, _, sw, sh := w.pages.GetRect()
	if sw == 0 || sh == 0 {
		sw, sh = 120, 40
	}

	if x+wid > sw {
		x = sw - wid
	}
	if x < 0 {
		x = 0
	}

	yOpen := y + 1
	if yOpen+hei > sh {
		yOpen = y - hei
	}
	if yOpen < 0 {
		yOpen = 0
	}

	grid := tview.NewGrid()
	grid.SetRows(yOpen, hei, 0)
	grid.SetColumns(x, wid, 0)
	grid.AddItem(list, 1, 1, 1, 1, 0, 0, true)

	w.pages.RemovePage("picker")
	w.pages.AddPage("picker", grid, true, true)
	w.app.SetFocus(list)
}

func (w *Wizard) pickerAnchor(tableRow int) (int, int) {
	tx, ty, _, _ := w.featureTable.GetRect()
	x := tx + 1 + featuresNameColWidth + 1
	y := ty + 1 + tableRow
	return x, y
}

func (w *Wizard) pickerOptions(it itemRef) ([]string, int) {
	if it.Kind == itemChoice {
		ch, ok := w.cat.Choices[it.ID]
		if !ok {
			return nil, 0
		}

		opts := make([]string, 0, len(ch.Options))
		val := w.p.Choices[it.ID]
		if val == "" {
			val = ch.Default
		}

		cur := 0
		for i, o := range ch.Options {
			opts = append(opts, o.Title)
			if o.ID == val {
				cur = i
			}
		}
		return opts, cur
	}

	on := w.p.Features[it.ID]
	if strings.EqualFold(w.currentCategory, "Install") {
		if on {
			return []string{"Install/Update", "Skip"}, 0
		}
		return []string{"Install/Update", "Skip"}, 1
	}

	if on {
		return []string{"Enable", "Disable"}, 0
	}
	return []string{"Enable", "Disable"}, 1
}

func (w *Wizard) applyPickerSelection(it itemRef, picked string) {
	if it.Kind == itemChoice {
		ch, ok := w.cat.Choices[it.ID]
		if !ok {
			return
		}

		for _, o := range ch.Options {
			if o.Title == picked {
				w.p.Choices[it.ID] = o.ID
				w.p.Normalize(w.cat)
				_ = profile.Save(w.p)
				return
			}
		}
		return
	}

	if strings.EqualFold(w.currentCategory, "Install") {
		w.p.Features[it.ID] = picked == "Install/Update"
	} else {
		w.p.Features[it.ID] = picked == "Enable"
	}

	w.p.Normalize(w.cat)
	_ = profile.Save(w.p)
}

func findChoiceOption(ch catalog.Choice, id string) (catalog.ChoiceOption, bool) {
	for _, opt := range ch.Options {
		if opt.ID == id {
			return opt, true
		}
	}
	return catalog.ChoiceOption{}, false
}
