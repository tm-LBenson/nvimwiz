package ui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/rivo/tview"

	"nvimwiz/internal/tasks"
)

func (w *Wizard) pageApply() tview.Primitive {
	w.logView = tview.NewTextView()
	w.logView.SetDynamicColors(true)
	w.logView.SetScrollable(true)
	w.logView.SetChangedFunc(func() { w.app.Draw() })
	w.logView.SetBorder(true)
	w.logView.SetTitle("Run")

	w.progressView = tview.NewTextView()
	w.progressView.SetBorder(true)
	w.progressView.SetTitle("Progress")

	buttons := tview.NewForm()
	buttons.AddButton("Back", func() { w.gotoPage("summary") })
	buttons.AddButton("Quit", func() { w.app.Stop() })
	buttons.SetButtonsAlign(tview.AlignCenter)

	wrap := tview.NewFlex().SetDirection(tview.FlexRow)
	wrap.AddItem(w.logView, 0, 1, false)
	wrap.AddItem(w.progressView, 3, 0, false)
	wrap.AddItem(buttons, 3, 0, true)
	return wrap
}

func (w *Wizard) startApply() {
	w.logView.SetText("")
	w.progressView.SetText("")
	w.taskPlan = tasks.Plan(w.p, w.cat)

	ctx, cancel := context.WithCancel(context.Background())
	_ = cancel

	logFn := func(msg string) {
		w.app.QueueUpdateDraw(func() {
			fmt.Fprintln(w.logView, strings.TrimRight(msg, "\n"))
			w.logView.ScrollToEnd()
		})
	}

	progressFn := func(done, total int) {
		w.app.QueueUpdateDraw(func() {
			w.progressView.SetText(fmt.Sprintf("%d/%d", done, total))
		})
	}

	go func() {
		start := time.Now()
		err := tasks.RunAll(ctx, w.taskPlan, logFn, progressFn)
		dur := time.Since(start).Round(time.Second)
		w.app.QueueUpdateDraw(func() {
			if err != nil {
				fmt.Fprintln(w.logView, "")
				fmt.Fprintln(w.logView, "Failed: "+err.Error())
			} else {
				fmt.Fprintln(w.logView, "")
				fmt.Fprintln(w.logView, "Done in "+dur.String())
				if w.p.ConfigMode == "integrate" {
					fmt.Fprintln(w.logView, "")
					fmt.Fprintln(w.logView, "Integrate mode: add require(\"nvimwiz.loader\") to your init.lua")
				}
			}
			w.logView.ScrollToEnd()
		})
	}()
}
