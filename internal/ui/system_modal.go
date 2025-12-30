package ui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func (w *Wizard) showSystemModal() {
	if w == nil || w.app == nil || w.pages == nil {
		return
	}

	const pageName = "system_modal"
	w.pages.RemovePage(pageName)

	content := tview.NewTextView()
	content.SetDynamicColors(true)
	content.SetText(w.systemInfoText())
	content.SetScrollable(true)
	content.SetInputCapture(func(ev *tcell.EventKey) *tcell.EventKey {
		if ev.Key() == tcell.KeyEsc {
			w.pages.RemovePage(pageName)
			w.app.SetFocus(w.pages)
			return nil
		}
		return ev
	})

	buttons := tview.NewForm()
	buttons.SetButtonsAlign(tview.AlignCenter)
	buttons.AddButton("Close", func() {
		w.pages.RemovePage(pageName)
		w.app.SetFocus(w.pages)
	})

	box := tview.NewFlex().SetDirection(tview.FlexRow)
	box.SetBorder(true)
	box.SetTitle("System")
	box.AddItem(content, 0, 1, true)
	box.AddItem(buttons, 3, 0, false)

	view := overlayCentered(box, 90, 26)
	w.pages.AddPage(pageName, view, true, true)
	w.app.SetFocus(content)
}
