package ui

import "github.com/rivo/tview"

func (w *Wizard) confirm(title, msg string, onOK func()) {
	modal := tview.NewModal()
	modal.SetTitle(title)
	modal.SetText(msg)
	modal.AddButtons([]string{"Cancel", "OK"})
	modal.SetDoneFunc(func(_ int, label string) {
		w.pages.RemovePage("modal")
		if w.app != nil {
			w.app.SetFocus(w.pages)
		}
		if label == "OK" && onOK != nil {
			onOK()
		}
	})
	w.pages.AddPage("modal", modal, true, true)
	if w.app != nil {
		w.app.SetFocus(modal)
	}
}

func (w *Wizard) message(title, msg string) {
	modal := tview.NewModal()
	modal.SetTitle(title)
	modal.SetText(msg)
	modal.AddButtons([]string{"OK"})
	modal.SetDoneFunc(func(_ int, _ string) {
		w.pages.RemovePage("modal")
		if w.app != nil {
			w.app.SetFocus(w.pages)
		}
	})
	w.pages.AddPage("modal", modal, true, true)
	if w.app != nil {
		w.app.SetFocus(modal)
	}
}
