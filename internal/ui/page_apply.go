package ui

import (
	"context"
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	"github.com/rivo/tview"

	"nvimwiz/internal/tasks"
)

var applyRunning int32

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

	buttonsNormal := tview.NewForm()
	buttonsNormal.AddButton("Back", func() { w.gotoPage("summary") })
	buttonsNormal.AddButton("Run again", func() { w.startApply() })
	buttonsNormal.AddButton("Quit", func() { w.app.Stop() })
	buttonsNormal.SetButtonsAlign(tview.AlignCenter)

	buttonsFailed := tview.NewForm()
	buttonsFailed.AddButton("Back", func() { w.gotoPage("summary") })
	buttonsFailed.AddButton("Retry failed", func() { w.retryFailedApply() })
	buttonsFailed.AddButton("Run again", func() { w.startApply() })
	buttonsFailed.AddButton("Quit", func() { w.app.Stop() })
	buttonsFailed.SetButtonsAlign(tview.AlignCenter)

	w.applyButtons = tview.NewPages()
	w.applyButtons.AddPage("normal", buttonsNormal, true, true)
	w.applyButtons.AddPage("failed", buttonsFailed, true, false)

	wrap := tview.NewFlex().SetDirection(tview.FlexRow)
	wrap.AddItem(w.logView, 0, 1, false)
	wrap.AddItem(w.progressView, 3, 0, false)
	wrap.AddItem(w.applyButtons, 3, 0, true)
	return wrap
}

func (w *Wizard) showApplyButtonsFailed(failed bool) {
	if w.applyButtons == nil {
		return
	}
	if failed {
		w.applyButtons.SwitchToPage("failed")
		return
	}
	w.applyButtons.SwitchToPage("normal")
}

func (w *Wizard) startApply() {
	w.startApplyFrom(0, true)
}

func (w *Wizard) retryFailedApply() {
	if w.applyFailedIndex < 0 {
		w.message("Retry failed", "No failed step to retry.")
		return
	}
	w.startApplyFrom(w.applyFailedIndex, false)
}

func (w *Wizard) startApplyFrom(startIndex int, reset bool) {
	if !atomic.CompareAndSwapInt32(&applyRunning, 0, 1) {
		return
	}

	if reset {
		w.showApplyButtonsFailed(false)
	}

	if reset {
		w.logView.SetText("")
		w.progressView.SetText("")
		w.taskPlan = tasks.Plan(w.p, w.cat)
		w.taskState = &tasks.State{}
		w.applyFailedIndex = -1
	} else {
		if w.taskPlan == nil || len(w.taskPlan) == 0 {
			w.taskPlan = tasks.Plan(w.p, w.cat)
		}
		if w.taskState == nil {
			w.taskState = &tasks.State{}
		}
		if startIndex < 0 {
			startIndex = 0
		}
		if startIndex >= len(w.taskPlan) {
			atomic.StoreInt32(&applyRunning, 0)
			w.message("Retry failed", "Nothing to retry.")
			return
		}
		fmt.Fprintln(w.logView, "")
		fmt.Fprintln(w.logView, "Retrying: "+w.taskPlan[startIndex].Name)
		w.logView.ScrollToEnd()
	}

	total := len(w.taskPlan)
	if startIndex < 0 {
		startIndex = 0
	}
	if startIndex > total {
		startIndex = total
	}
	w.progressView.SetText(fmt.Sprintf("%d/%d", startIndex, total))

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

	go func(startAt int, total int) {
		defer atomic.StoreInt32(&applyRunning, 0)
		start := time.Now()
		st, failedAt, err := tasks.RunFrom(ctx, w.taskPlan, w.taskState, startAt, logFn, progressFn)
		dur := time.Since(start).Round(time.Second)
		w.app.QueueUpdateDraw(func() {
			w.taskState = st
			if err != nil {
				w.applyFailedIndex = failedAt
				w.showApplyButtonsFailed(true)
				fmt.Fprintln(w.logView, "")
				if failedAt >= 0 && failedAt < len(w.taskPlan) {
					fmt.Fprintln(w.logView, fmt.Sprintf("Failed at %d/%d (%s): %s", failedAt+1, total, w.taskPlan[failedAt].Name, err.Error()))
				} else {
					fmt.Fprintln(w.logView, "Failed: "+err.Error())
				}
			} else {
				w.applyFailedIndex = -1
				w.showApplyButtonsFailed(false)
				fmt.Fprintln(w.logView, "")
				fmt.Fprintln(w.logView, "Done in "+dur.String())
				if w.p.ConfigMode == "integrate" {
					fmt.Fprintln(w.logView, "")
					fmt.Fprintln(w.logView, "Integrate mode: add require(\"nvimwiz.loader\") to your init.lua")
				}
			}
			w.logView.ScrollToEnd()
		})
	}(startIndex, total)
}
