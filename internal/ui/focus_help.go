package ui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type mouseCapturer interface {
	SetMouseCapture(func(action tview.MouseAction, event *tcell.EventMouse) (tview.MouseAction, *tcell.EventMouse))
}

type inputCapturer interface {
	SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey)
}

func setFocusHelp(item tview.FormItem, onFocus func()) {
	if onFocus == nil {
		return
	}

	if p, ok := item.(mouseCapturer); ok {
		p.SetMouseCapture(func(action tview.MouseAction, event *tcell.EventMouse) (tview.MouseAction, *tcell.EventMouse) {
			onFocus()
			return action, event
		})
	}

	if p, ok := item.(inputCapturer); ok {
		p.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			onFocus()
			return event
		})
	}
}
