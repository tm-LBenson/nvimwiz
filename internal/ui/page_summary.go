package ui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/rivo/tview"

	"nvimwiz/internal/profile"
	"nvimwiz/internal/tasks"
)

func (w *Wizard) pageSummary() tview.Primitive {
	tv := tview.NewTextView()
	tv.SetDynamicColors(true)
	tv.SetBorder(true)
	tv.SetTitle("Summary")

	w.summaryView = tv

	buttons := tview.NewForm()
	buttons.AddButton("Back", func() { w.gotoPage("features") })
	buttons.AddButton("Apply", func() {
		w.gotoPage("apply")
		w.startApply()
	})
	buttons.AddButton("Save", func() { _ = profile.Save(w.p) })
	buttons.AddButton("Quit", func() { w.app.Stop() })
	buttons.SetButtonsAlign(tview.AlignCenter)

	flex := tview.NewFlex().SetDirection(tview.FlexRow)
	flex.AddItem(tv, 0, 1, false)
	flex.AddItem(buttons, 3, 0, true)

	w.renderSummary()
	return flex
}

func (w *Wizard) renderSummary() {
	if w.summaryView == nil {
		return
	}

	w.taskPlan = tasks.Plan(w.p, w.cat)

	lines := []string{}
	lines = append(lines, "Preset: "+w.p.Preset)
	lines = append(lines, "Config mode: "+w.p.ConfigMode)
	lines = append(lines, "Verify: "+w.p.Verify)
	lines = append(lines, "Projects dir: "+w.p.ProjectsDir)
	lines = append(lines, "")

	lines = append(lines, "Choices:")
	for _, k := range []string{"ui.explorer", "ui.theme", "ui.statusline"} {
		v := w.p.Choices[k]
		if v == "" {
			continue
		}
		lines = append(lines, fmt.Sprintf(" - %s = %s", k, v))
	}

	lines = append(lines, "")
	lines = append(lines, "Enabled features:")
	fids := []string{}
	for id, on := range w.p.Features {
		if on {
			fids = append(fids, id)
		}
	}
	sort.Strings(fids)
	for _, id := range fids {
		lines = append(lines, " - "+id)
	}

	lines = append(lines, "")
	lines = append(lines, "Plan:")
	for _, t := range w.taskPlan {
		lines = append(lines, " - "+t.Name)
	}

	w.summaryView.SetText(strings.Join(lines, "\n"))
}
