package ui

import "github.com/rivo/tview"

func (w *Wizard) pageWelcome() tview.Primitive {
	tv := tview.NewTextView()
	tv.SetDynamicColors(true)
	tv.SetTextAlign(tview.AlignLeft)
	tv.SetText("nvimwiz\n\nA setup wizard for a modular Neovim config.\n\nPress Start to configure presets and features.")
	tv.SetBorder(true)
	tv.SetTitle("Welcome")

	form := tview.NewForm()
	form.AddButton("Start", func() {
		w.gotoPage("settings")
	})
	form.AddButton("Quit", func() {
		w.app.Stop()
	})
	form.SetButtonsAlign(tview.AlignCenter)

	flex := tview.NewFlex().SetDirection(tview.FlexRow)
	flex.AddItem(tv, 0, 1, false)
	flex.AddItem(form, 3, 0, true)
	return flex
}
