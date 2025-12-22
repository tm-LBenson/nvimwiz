package ui

import (
	"sort"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"nvimwiz/internal/catalog"
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
	categories := make([]string, 0, len(w.cat.Categories))
	categories = append(categories, w.cat.Categories...)
	if len(categories) == 0 {
		categories = []string{"Install"}
	}
	if w.currentCategory == "" {
		w.currentCategory = categories[0]
	}

	tabs := tview.NewTable()
	tabs.SetBorder(true)
	tabs.SetTitle("Category")
	tabs.SetSelectable(true, true)
	tabs.SetFixed(0, 0)

	for i, cat := range categories {
		txt := " " + cat + " "
		cell := tview.NewTableCell(txt)
		cell.SetAlign(tview.AlignCenter)
		cell.SetSelectable(true)
		tabs.SetCell(0, i, cell)
		if cat == w.currentCategory {
			tabs.Select(0, i)
		}
	}

	tabs.SetSelectionChangedFunc(func(row, col int) {
		if col < 0 || col >= len(categories) {
			return
		}
		w.pages.RemovePage("picker")
		w.currentCategory = categories[col]
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
	root.AddItem(tabs, 3, 0, true)
	root.AddItem(body, 0, 1, true)
	root.AddItem(help, 3, 0, false)
	root.AddItem(buttons, 3, 0, true)

	focusOrder := []tview.Primitive{tabs, w.featureTable, w.detailView}
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

	h0 := tview.NewTableCell(padRight("Name", featuresNameColWidth))
	h1 := tview.NewTableCell(padRight("Action", featuresActionColWidth))
	h0.SetSelectable(false)
	h1.SetSelectable(false)
	w.featureTable.SetCell(0, featuresNameCol, h0)
	w.featureTable.SetCell(0, featuresActionCol, h1)

	items := make([]itemRef, 0)
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
		row := 1 + i*featuresRowStride
		spacer := row + 1

		name := fixedWidth(w.itemTitle(it), featuresNameColWidth)
		action := fixedWidth(" "+w.itemActionLabel(it)+" ", featuresActionColWidth)

		nameCell := tview.NewTableCell(name)
		nameCell.SetSelectable(true)
		actionCell := tview.NewTableCell(action)
		actionCell.SetSelectable(true)
		actionCell.SetAlign(tview.AlignCenter)
		actionCell.SetTextColor(tcell.ColorWhite)
		actionCell.SetBackgroundColor(tcell.ColorBlue)
		actionCell.SetAttributes(tcell.AttrBold)

		w.featureTable.SetCell(row, featuresNameCol, nameCell)
		w.featureTable.SetCell(row, featuresActionCol, actionCell)

		sp0 := tview.NewTableCell("")
		sp1 := tview.NewTableCell("")
		sp0.SetSelectable(false)
		sp1.SetSelectable(false)
		w.featureTable.SetCell(spacer, featuresNameCol, sp0)
		w.featureTable.SetCell(spacer, featuresActionCol, sp1)
	}

	if len(items) > 0 {
		w.featureTable.Select(1, featuresNameCol)
		it := w.rowItems[0]
		w.renderDetails(it)
	} else {
		w.detailView.SetText("")
	}
}

func (w *Wizard) itemAtRow(row int) (itemRef, bool) {
	if row <= 0 {
		return itemRef{}, false
	}
	if row%featuresRowStride == 0 {
		return itemRef{}, false
	}
	idx := (row - 1) / featuresRowStride
	if idx < 0 || idx >= len(w.rowItems) {
		return itemRef{}, false
	}
	return w.rowItems[idx], true
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

func (w *Wizard) itemActionLabel(it itemRef) string {
	if it.Kind == itemChoice {
		ch, ok := w.cat.Choices[it.ID]
		if !ok {
			return "Select"
		}
		val := w.p.Choices[it.ID]
		if val == "" {
			val = ch.Default
		}
		if opt, ok := findChoiceOption(ch, val); ok {
			return opt.Title
		}
		return val
	}

	on := w.p.Features[it.ID]
	if strings.EqualFold(w.currentCategory, "Install") {
		if on {
			return "Install"
		}
		return "Skip"
	}
	if on {
		return "Enable"
	}
	return "Disable"
}

func (w *Wizard) renderDetails(it itemRef) {
	title := w.itemTitle(it)

	lines := []string{title, ""}

	if it.Kind == itemFeature {
		if f, ok := w.cat.Features[it.ID]; ok {
			note := beginnerNote(it.ID, f.Title)
			if note != "" {
				lines = append(lines, "Info:", note, "")
			}

			desc := strings.TrimSpace(f.Long)
			if desc == "" {
				desc = strings.TrimSpace(f.Short)
			}
			if desc != "" {
				lines = append(lines, desc, "")
			}

			if len(f.Requires) > 0 {
				lines = append(lines, "Requires: "+strings.Join(f.Requires, ", "), "")
			}
		}
		lines = append(lines, "Current: "+w.itemActionLabel(it))
		w.detailView.SetText(strings.Join(trimTrailingEmpty(lines), "\n"))
		return
	}

	if it.Kind == itemChoice {
		if c, ok := w.cat.Choices[it.ID]; ok {
			note := beginnerNote(it.ID, c.Title)
			if note != "" {
				lines = append(lines, "Info:", note, "")
			}

			desc := strings.TrimSpace(c.Long)
			if desc == "" {
				desc = strings.TrimSpace(c.Short)
			}
			if desc != "" {
				lines = append(lines, desc, "")
			}

			lines = append(lines, "Options:")
			for _, opt := range c.Options {
				lines = append(lines, " - "+opt.Title)
			}
			lines = append(lines, "")
		}
		lines = append(lines, "Current: "+w.itemActionLabel(it))
		w.detailView.SetText(strings.Join(trimTrailingEmpty(lines), "\n"))
		return
	}
}

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
			return []string{"Install", "Skip"}, 0
		}
		return []string{"Install", "Skip"}, 1
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
		w.p.Features[it.ID] = picked == "Install"
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

func beginnerNote(id, title string) string {
	low := strings.ToLower(title + " " + id)

	switch {
	case strings.Contains(low, "telescope"):
		return "Telescope is a fast finder. Use it to open files, search text (grep), and jump around your project without a mouse."
	case strings.Contains(low, "treesitter"):
		return "Treesitter improves syntax highlighting and code navigation using real parsing (not regex). It usually makes editing feel more \"IDE-like\"."
	case strings.Contains(low, "lsp"):
		return "LSP adds IDE features: autocomplete, go-to-definition, rename, diagnostics, hover docs. If you code, you almost always want this."
	case strings.Contains(low, "mason"):
		return "Mason installs language servers and dev tools (like TypeScript/Go/Python servers) so Neovim can provide IDE features."
	case strings.Contains(low, "nvim-tree") || strings.Contains(low, "file tree") || strings.Contains(low, "tree"):
		return "A file explorer sidebar (like VS Code). Useful if you prefer browsing folders visually."
	case strings.Contains(low, "ripgrep") || strings.Contains(low, " rg"):
		return "ripgrep (rg) is a very fast search tool. Many Neovim search features use it under the hood."
	case strings.Contains(low, "fd"):
		return "fd is a faster, friendlier 'find' command. Plugins use it to list files quickly."
	case strings.Contains(low, "neovim") || strings.Contains(low, "nvim"):
		return "This is the editor itself. Installing a stable version avoids random breakage from old distro packages."
	case strings.Contains(low, "statusline"):
		return "A statusline shows mode, file info, git branch, diagnostics, and more at the bottom of the screen."
	case strings.Contains(low, "theme") || strings.Contains(low, "colorscheme"):
		return "A theme changes the colors of the UI and syntax highlighting. Pick one that is comfortable for your eyes."
	}

	return ""
}

func fixedWidth(s string, width int) string {
	s = strings.ReplaceAll(s, "\t", " ")
	if width <= 0 {
		return s
	}
	r := []rune(s)
	if len(r) > width {
		r = r[:width]
	}
	out := string(r)
	if len([]rune(out)) < width {
		out = padRight(out, width)
	}
	return out
}

func padRight(s string, width int) string {
	r := []rune(s)
	if len(r) >= width {
		return s
	}
	return s + strings.Repeat(" ", width-len(r))
}

func trimTrailingEmpty(lines []string) []string {
	i := len(lines) - 1
	for i >= 0 && strings.TrimSpace(lines[i]) == "" {
		i--
	}
	return lines[:i+1]
}

func maxStringLen(ss []string) int {
	m := 0
	for _, s := range ss {
		if len([]rune(s)) > m {
			m = len([]rune(s))
		}
	}
	return m
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
