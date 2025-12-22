package ui

import (
	"sort"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"nvimwiz/internal/catalog"
	"nvimwiz/internal/profile"
)

func (w *Wizard) pageFeatures() tview.Primitive {
	w.categoryList = tview.NewList()
	w.categoryList.SetBorder(true)
	w.categoryList.SetTitle("Category")

	w.featureTable = tview.NewTable()
	w.featureTable.SetBorder(true)
	w.featureTable.SetTitle("Features")
	w.featureTable.SetSelectable(true, true)
	w.featureTable.SetFixed(1, 0)

	w.detailView = tview.NewTextView()
	w.detailView.SetDynamicColors(true)
	w.detailView.SetBorder(true)
	w.detailView.SetTitle("Details")
	w.detailView.SetText("Select a row. Space or Enter toggles. Tab changes focus. Select [ ? ] for more details.")

	helpBar := tview.NewTextView()
	helpBar.SetDynamicColors(true)
	helpBar.SetBorder(true)
	helpBar.SetTitle("Help")
	helpBar.SetText("Up/Down: move   Tab: switch pane   Space/Enter: toggle or choose   Left/Right: move columns   [ ? ]: details")

	for _, c := range w.cat.Categories {
		cat := c
		w.categoryList.AddItem(cat, "", 0, nil)
	}

	w.categoryList.SetChangedFunc(func(_ int, mainText string, _ string, _ rune) {
		if mainText == "" {
			return
		}
		w.currentCategory = mainText
		w.renderFeatureTable()
	})

	w.featureTable.SetSelectionChangedFunc(func(row, col int) {
		if row <= 0 || row-1 >= len(w.rowItems) {
			w.detailView.SetText("")
			return
		}
		it := w.rowItems[row-1]
		long := col == 2
		w.updateDetail(it, long)
	})

	w.featureTable.SetSelectedFunc(func(row, col int) {
		if row <= 0 || row-1 >= len(w.rowItems) {
			return
		}
		it := w.rowItems[row-1]
		if col == 2 {
			w.updateDetail(it, true)
			return
		}
		switch it.Kind {
		case itemFeature:
			w.toggleFeature(it.ID)
			w.renderFeatureTable()
		case itemChoice:
			w.pickChoice(it.ID)
		}
	})

	w.featureTable.SetInputCapture(func(ev *tcell.EventKey) *tcell.EventKey {
		if ev.Key() == tcell.KeyRune && ev.Rune() == ' ' {
			r, c := w.featureTable.GetSelection()
			if r > 0 && r-1 < len(w.rowItems) {
				it := w.rowItems[r-1]
				if c == 2 {
					w.updateDetail(it, true)
					return nil
				}
				if it.Kind == itemFeature {
					w.toggleFeature(it.ID)
					w.renderFeatureTable()
					return nil
				}
				if it.Kind == itemChoice {
					w.pickChoice(it.ID)
					return nil
				}
			}
		}
		if ev.Key() == tcell.KeyRune && ev.Rune() == '?' {
			r, _ := w.featureTable.GetSelection()
			if r > 0 && r-1 < len(w.rowItems) {
				w.updateDetail(w.rowItems[r-1], true)
				return nil
			}
		}
		return ev
	})

	if len(w.cat.Categories) > 0 {
		w.currentCategory = w.cat.Categories[0]
		w.categoryList.SetCurrentItem(0)
	}
	w.renderFeatureTable()

	main := tview.NewFlex()
	main.AddItem(w.categoryList, 0, 1, true)
	main.AddItem(w.featureTable, 0, 2, false)
	main.AddItem(w.detailView, 0, 3, false)

	buttons := tview.NewForm()
	buttons.AddButton("Back", func() { w.gotoPage("settings") })
	buttons.AddButton("Save", func() { _ = profile.Save(w.p) })
	buttons.AddButton("Summary", func() {
		w.p.Normalize(w.cat)
		_ = profile.Save(w.p)
		w.gotoPage("summary")
	})
	buttons.AddButton("Quit", func() { w.app.Stop() })
	buttons.SetButtonsAlign(tview.AlignCenter)

	root := tview.NewFlex().SetDirection(tview.FlexRow)
	root.AddItem(main, 0, 1, true)
	root.AddItem(helpBar, 3, 0, false)
	root.AddItem(buttons, 3, 0, true)

	focusOrder := []tview.Primitive{w.categoryList, w.featureTable, w.detailView}
	root.SetInputCapture(func(ev *tcell.EventKey) *tcell.EventKey {
		if ev.Key() == tcell.KeyTab {
			cur := w.app.GetFocus()
			next := 0
			for i, p := range focusOrder {
				if p == cur {
					next = (i + 1) % len(focusOrder)
					break
				}
			}
			w.app.SetFocus(focusOrder[next])
			return nil
		}
		return ev
	})

	return root
}

func (w *Wizard) renderFeatureTable() {
	w.featureTable.Clear()

	h0 := tview.NewTableCell("State")
	h1 := tview.NewTableCell("Name")
	h2 := tview.NewTableCell("")
	h0.SetSelectable(false)
	h1.SetSelectable(false)
	h2.SetSelectable(false)
	w.featureTable.SetCell(0, 0, h0)
	w.featureTable.SetCell(0, 1, h1)
	w.featureTable.SetCell(0, 2, h2)

	items := []itemRef{}
	for _, ch := range w.cat.Choices {
		if ch.Category == w.currentCategory {
			items = append(items, itemRef{Kind: itemChoice, ID: ch.Key})
		}
	}
	for _, f := range w.cat.Features {
		if f.Category == w.currentCategory {
			items = append(items, itemRef{Kind: itemFeature, ID: f.ID})
		}
	}

	sort.Slice(items, func(i, j int) bool {
		at := w.itemTitle(items[i])
		bt := w.itemTitle(items[j])
		return strings.ToLower(at) < strings.ToLower(bt)
	})

	w.rowItems = items
	for i, it := range items {
		row := i + 1
		state := w.itemState(it)
		name := w.itemTitle(it)

		stateCell := tview.NewTableCell(state)
		stateCell.SetAlign(tview.AlignCenter)
		w.featureTable.SetCell(row, 0, stateCell)

		w.featureTable.SetCell(row, 1, tview.NewTableCell(name))

		infoCell := tview.NewTableCell(" ? ")
		infoCell.SetAlign(tview.AlignCenter)
		infoCell.SetAttributes(tcell.AttrReverse)
		w.featureTable.SetCell(row, 2, infoCell)
	}

	if len(items) > 0 {
		w.featureTable.Select(1, 0)
	}
}

func (w *Wizard) itemTitle(it itemRef) string {
	if it.Kind == itemFeature {
		if f, ok := w.cat.Features[it.ID]; ok {
			return f.Title
		}
		return it.ID
	}
	if it.Kind == itemChoice {
		if c, ok := w.cat.Choices[it.ID]; ok {
			return c.Title
		}
		return it.ID
	}
	return it.ID
}

func (w *Wizard) itemState(it itemRef) string {
	if it.Kind == itemFeature {
		if w.p.Features[it.ID] {
			return tview.Escape("[ x ]")
		}
		return tview.Escape("[ ]")
	}
	if it.Kind == itemChoice {
		val := w.p.Choices[it.ID]
		ch, ok := w.cat.Choices[it.ID]
		if !ok {
			return val
		}
		if val == "" {
			val = ch.Default
		}
		opt, ok := findChoiceOption(ch, val)
		if ok {
			return opt.Title
		}
		return val
	}
	return ""
}

func findChoiceOption(ch catalog.Choice, id string) (catalog.ChoiceOption, bool) {
	for _, opt := range ch.Options {
		if opt.ID == id {
			return opt, true
		}
	}
	return catalog.ChoiceOption{}, false
}

func (w *Wizard) updateDetail(it itemRef, long bool) {
	lines := []string{}
	if it.Kind == itemFeature {
		f, ok := w.cat.Features[it.ID]
		if !ok {
			w.detailView.SetText(it.ID)
			return
		}
		lines = append(lines, f.Title)
		lines = append(lines, "")
		if long {
			lines = append(lines, f.Long)
		} else {
			lines = append(lines, f.Short)
		}
		if len(f.Requires) > 0 {
			lines = append(lines, "")
			lines = append(lines, "Requires: "+strings.Join(f.Requires, ", "))
		}
		w.detailView.SetText(strings.Join(lines, "\n"))
		return
	}
	if it.Kind == itemChoice {
		c, ok := w.cat.Choices[it.ID]
		if !ok {
			w.detailView.SetText(it.ID)
			return
		}
		lines = append(lines, c.Title)
		lines = append(lines, "")
		if long {
			lines = append(lines, c.Long)
		} else {
			lines = append(lines, c.Short)
		}
		lines = append(lines, "")
		lines = append(lines, "Options:")
		for _, opt := range c.Options {
			lines = append(lines, " - "+opt.Title+" ("+opt.ID+")")
		}
		w.detailView.SetText(strings.Join(lines, "\n"))
	}
}

func (w *Wizard) toggleFeature(id string) {
	w.p.Features[id] = !w.p.Features[id]
	w.p.Normalize(w.cat)
	_ = profile.Save(w.p)
}

func (w *Wizard) pickChoice(key string) {
	ch, ok := w.cat.Choices[key]
	if !ok {
		return
	}
	buttons := []string{}
	optByButton := map[string]string{}
	for _, opt := range ch.Options {
		buttons = append(buttons, opt.Title)
		optByButton[opt.Title] = opt.ID
	}

	modal := tview.NewModal()
	modal.SetText(ch.Title)
	modal.AddButtons(buttons)
	modal.SetDoneFunc(func(_ int, buttonLabel string) {
		if id, ok := optByButton[buttonLabel]; ok {
			w.p.Choices[key] = id
			w.p.Normalize(w.cat)
			_ = profile.Save(w.p)
			w.renderFeatureTable()
		}
		w.pages.RemovePage("choice")
		w.app.SetFocus(w.featureTable)
	})
	modal.SetBorder(true)
	modal.SetTitle("Select")
	w.pages.AddPage("choice", modal, true, true)
	w.app.SetFocus(modal)
}
