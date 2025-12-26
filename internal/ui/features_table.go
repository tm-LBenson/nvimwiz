package ui

import (
	"sort"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func (w *Wizard) renderFeatureTable() {
	w.featureTable.Clear()

	h0 := tview.NewTableCell(padRight("Name", featuresNameColWidth))
	h1 := tview.NewTableCell(padRight("Action", featuresActionColWidth))
	h0.SetSelectable(false)
	h1.SetSelectable(false)
	w.featureTable.SetCell(0, featuresNameCol, h0)
	w.featureTable.SetCell(0, featuresActionCol, h1)

	rowItems := make([]itemRef, 0)

	for _, choice := range w.cat.Choices {
		if choice.Category == w.currentCategory {
			rowItems = append(rowItems, itemRef{Kind: itemChoice, ID: choice.Key})
		}
	}
	for _, feat := range w.cat.Features {
		if feat.Category == w.currentCategory {
			rowItems = append(rowItems, itemRef{Kind: itemFeature, ID: feat.ID})
		}
	}

	sort.Slice(rowItems, func(i, j int) bool {
		at := w.itemTitle(rowItems[i])
		bt := w.itemTitle(rowItems[j])
		return strings.ToLower(at) < strings.ToLower(bt)
	})

	w.rowItems = rowItems

	for i, ref := range rowItems {
		row := 1 + i*featuresRowStride
		spacer := row + 1

		name := fixedWidth(w.itemTitle(ref), featuresNameColWidth)
		action := fixedWidth(" "+w.itemActionLabel(ref)+" ", featuresActionColWidth)

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

	if len(rowItems) > 0 {
		w.featureTable.Select(1, featuresNameCol)
		ref := w.rowItems[0]
		w.renderDetails(ref)
		return
	}

	w.detailView.SetText("")
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
			return "Install/Update"
		}
		return "Skip"
	}

	if on {
		return "Enable"
	}
	return "Disable"
}
