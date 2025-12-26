package ui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"nvimwiz/internal/profile"
)

const (
	featuresNameCol        = 0
	featuresActionCol      = 1
	featuresNameColWidth   = 20
	featuresActionColWidth = 12
	featuresRowStride      = 2
)

func (w *Wizard) pageFeatures() tview.Primitive {
	categoryNames := make([]string, 0, len(w.cat.Categories))
	categoryNames = append(categoryNames, w.cat.Categories...)
	if len(categoryNames) == 0 {
		categoryNames = []string{"Install"}
	}
	if w.currentCategory == "" {
		w.currentCategory = categoryNames[0]
	}

	categoryTabs := tview.NewTable()
	categoryTabs.SetBorder(true)
	categoryTabs.SetTitle("Category")
	categoryTabs.SetSelectable(true, true)
	categoryTabs.SetFixed(0, 0)

	for i, category := range categoryNames {
		text := " " + category + " "
		cell := tview.NewTableCell(text)
		cell.SetAlign(tview.AlignCenter)
		cell.SetSelectable(true)
		categoryTabs.SetCell(0, i, cell)
		if category == w.currentCategory {
			categoryTabs.Select(0, i)
		}
	}

	categoryTabs.SetSelectionChangedFunc(func(row, col int) {
		if col < 0 || col >= len(categoryNames) {
			return
		}
		w.pages.RemovePage("picker")
		w.currentCategory = categoryNames[col]
		w.renderFeatureTable()
	})

	w.featureTable = tview.NewTable()
	w.featureTable.SetBorder(true)
	w.featureTable.SetTitle("Features")
	w.featureTable.SetSelectable(true, true)
	w.featureTable.SetFixed(1, 0)

	w.detailView = tview.NewTextView()
	w.detailView.SetDynamicColors(true)
	w.detailView.SetBorder(true)
	w.detailView.SetTitle("Details")

	mouseJustClicked := false
	w.featureTable.SetMouseCapture(func(action tview.MouseAction, event *tcell.EventMouse) (tview.MouseAction, *tcell.EventMouse) {
		if action == tview.MouseLeftClick {
			mouseJustClicked = true
		}
		return action, event
	})

	w.featureTable.SetSelectionChangedFunc(func(row, col int) {
		if row <= 0 {
			w.detailView.SetText("")
			return
		}
		if row%featuresRowStride == 0 {
			row = row - 1
			if row < 1 {
				row = 1
			}
			w.featureTable.Select(row, col)
			return
		}

		it, ok := w.itemAtRow(row)
		if !ok {
			w.detailView.SetText("")
			return
		}
		w.renderDetails(it)

		if mouseJustClicked && col == featuresActionCol {
			mouseJustClicked = false
			w.openPicker(it, row)
			return
		}
		mouseJustClicked = false
	})

	w.featureTable.SetSelectedFunc(func(row, col int) {
		if row <= 0 {
			return
		}
		if row%featuresRowStride == 0 {
			row = row - 1
			if row < 1 {
				row = 1
			}
		}
		it, ok := w.itemAtRow(row)
		if !ok {
			return
		}
		if col == featuresActionCol {
			w.openPicker(it, row)
			return
		}
		w.renderDetails(it)
	})

	w.featureTable.SetInputCapture(func(ev *tcell.EventKey) *tcell.EventKey {
		switch ev.Key() {
		case tcell.KeyUp:
			if len(w.rowItems) == 0 {
				return nil
			}
			r, c := w.featureTable.GetSelection()
			r = r - featuresRowStride
			if r < 1 {
				r = 1
			}
			w.featureTable.Select(r, c)
			return nil
		case tcell.KeyDown:
			if len(w.rowItems) == 0 {
				return nil
			}
			r, c := w.featureTable.GetSelection()
			maxRow := 1 + (len(w.rowItems)-1)*featuresRowStride
			r = r + featuresRowStride
			if r > maxRow {
				r = maxRow
			}
			w.featureTable.Select(r, c)
			return nil
		case tcell.KeyEnter:
			r, c := w.featureTable.GetSelection()
			if r <= 0 {
				return nil
			}
			if r%featuresRowStride == 0 {
				r = r - 1
			}
			it, ok := w.itemAtRow(r)
			if !ok {
				return nil
			}
			if c != featuresActionCol {
				c = featuresActionCol
				w.featureTable.Select(r, c)
			}
			w.openPicker(it, r)
			return nil
		case tcell.KeyRune:
			if ev.Rune() == ' ' {
				r, _ := w.featureTable.GetSelection()
				if r <= 0 {
					return nil
				}
				if r%featuresRowStride == 0 {
					r = r - 1
				}
				it, ok := w.itemAtRow(r)
				if !ok {
					return nil
				}
				w.featureTable.Select(r, featuresActionCol)
				w.openPicker(it, r)
				return nil
			}
		}
		return ev
	})

	w.renderFeatureTable()

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

	help := tview.NewTextView()
	help.SetBorder(true)
	help.SetTitle("Help")
	help.SetText("Tab: switch pane   Up/Down: move   Enter/Space/Click: open Action dropdown")

	body := tview.NewFlex()
	body.AddItem(w.featureTable, 0, 2, true)
	body.AddItem(w.detailView, 0, 3, false)

	root := tview.NewFlex().SetDirection(tview.FlexRow)
	root.AddItem(categoryTabs, 3, 0, true)
	root.AddItem(body, 0, 1, true)
	root.AddItem(help, 3, 0, false)
	root.AddItem(buttons, 3, 0, true)

	focusOrder := []tview.Primitive{categoryTabs, w.featureTable, w.detailView}
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
