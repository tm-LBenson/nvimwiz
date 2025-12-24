package ui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"nvimwiz/internal/catalog"
	"nvimwiz/internal/env"
	"nvimwiz/internal/profile"
	"nvimwiz/internal/sysinfo"
	"nvimwiz/internal/tasks"
)

type itemKind int

const (
	itemFeature itemKind = iota
	itemChoice
)

type itemRef struct {
	Kind itemKind
	ID   string
}

type Wizard struct {
	app   *tview.Application
	pages *tview.Pages

	cat catalog.Catalog
	p   profile.Profile
	sys sysinfo.Info

	envNote string

	settingsInfo *tview.TextView

	categoryList *tview.List
	featureTable *tview.Table
	detailView   *tview.TextView

	actionDropDown *tview.DropDown
	actionLabel    *tview.TextView

	currentCategory string
	rowItems        []itemRef

	logView      *tview.TextView
	progressView *tview.TextView
	summaryView  *tview.TextView

	taskPlan []tasks.Task
}

func New(app *tview.Application) (*Wizard, error) {
	cat := catalog.Get()
	p, _, err := profile.Load(cat)
	if err != nil {
		return nil, err
	}
	changed, lb, err := env.EnsureLocalBinInPath()
	note := ""
	if err == nil && changed {
		note = "Added " + lb + " to PATH for this run"
	}
	w := &Wizard{
		app:     app,
		pages:   tview.NewPages(),
		cat:     cat,
		p:       p,
		sys:     sysinfo.Collect(),
		envNote: note,
	}
	return w, nil
}

func (w *Wizard) Run() error {
	w.app.EnableMouse(true)
	w.pages.AddPage("welcome", w.pageWelcome(), true, true)
	w.pages.AddPage("settings", w.pageSettings(), true, false)
	w.pages.AddPage("features", w.pageFeatures(), true, false)
	w.pages.AddPage("summary", w.pageSummary(), true, false)
	w.pages.AddPage("apply", w.pageApply(), true, false)

	w.app.SetRoot(w.pages, true)
	w.app.SetInputCapture(func(ev *tcell.EventKey) *tcell.EventKey {
		if ev.Key() == tcell.KeyCtrlC {
			w.app.Stop()
			return nil
		}
		return ev
	})
	return w.app.Run()
}

func (w *Wizard) gotoPage(name string) {
	if name == "summary" {
		w.renderSummary()
	}
	w.pages.SwitchToPage(name)
}
func (w *Wizard) applyPreset(presetID string) {
	preset, ok := w.cat.Presets[presetID]
	if !ok {
		return
	}
	w.p.Preset = presetID
	if w.p.Features == nil {
		w.p.Features = map[string]bool{}
	}
	if w.p.Choices == nil {
		w.p.Choices = map[string]string{}
	}
	for featureID, enabled := range preset.Features {
		w.p.Features[featureID] = enabled
	}
	for choiceKey, choiceValue := range preset.Choices {
		w.p.Choices[choiceKey] = choiceValue
	}
	w.p.Normalize(w.cat)
	_ = profile.Save(w.p)
}
