package ui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"nvimwiz/internal/env"
	"nvimwiz/internal/profile"
	"nvimwiz/internal/sysinfo"
	"nvimwiz/internal/tasks"
)

type Wizard struct {
	app   *tview.Application
	pages *tview.Pages

	systemText *tview.TextView

	projectsDir string
	leader      string
	localLeader string
	configMode  string

	// feature selections
	installNeovim    bool
	installRipgrep   bool
	installFd        bool
	writeConfig      bool
	enableTree       bool
	enableTelescope  bool
	enableLSP        bool
	enableJava       bool
	runLazySync      bool
	requireChecksums bool

	profilePath     string
	profileLoadNote string
	localBin        string
	localBinNote    string

	progressView *tview.TextView
	logView      *tview.TextView

	cancel context.CancelFunc
}

func NewWizard(app *tview.Application) *Wizard {
	w := &Wizard{
		app:             app,
		pages:           tview.NewPages(),
		projectsDir:     "~/projects",
		leader:          " ",
		localLeader:     " ",
		configMode:      "managed",
		installNeovim:   true,
		installRipgrep:  true,
		installFd:       true,
		writeConfig:     true,
		enableTree:      true,
		enableTelescope: true,
		enableLSP:       true,
		enableJava:      false,
		runLazySync:     true,
	}

	if changed, lb, err := env.EnsureLocalBinInPath(); err == nil {
		w.localBin = lb
		if changed {
			w.localBinNote = "Added ~/.local/bin to PATH for this nvimwiz session (no relaunch needed)."
		}
	} else {
		w.localBinNote = "Could not determine ~/.local/bin: " + err.Error()
	}
	if w.localBin == "" {
		if lb, err := env.LocalBin(); err == nil {
			w.localBin = lb
		}
	}

	if pth, err := profile.ProfilePath(); err == nil {
		w.profilePath = pth
	}
	p, existed, err := profile.Load()
	if err != nil {
		w.profileLoadNote = "Profile load error: " + err.Error()
	} else if existed {
		w.profileLoadNote = "Loaded profile: " + w.profilePath
		w.applyProfile(p)
	} else {
		w.profileLoadNote = "No profile found yet (will create one on Save/Install)."
		w.applyProfile(p)
	}

	return w
}

func (w *Wizard) applyProfile(p profile.Profile) {
	p.Normalize()
	w.projectsDir = p.ProjectsDir
	w.leader = p.Leader
	w.localLeader = p.LocalLeader
	w.configMode = p.ConfigMode

	w.installNeovim = p.InstallNeovim
	w.installRipgrep = p.InstallRipgrep
	w.installFd = p.InstallFd
	w.writeConfig = p.WriteNvimConfig

	w.enableTree = p.EnableTree
	w.enableTelescope = p.EnableTelescope
	w.enableLSP = p.EnableLSP
	w.enableJava = p.EnableJava

	w.runLazySync = p.RunLazySync
	w.requireChecksums = p.RequireChecksums
}

func (w *Wizard) buildProfile() profile.Profile {
	p := profile.Profile{
		Version:          profile.CurrentVersion,
		ProjectsDir:      w.projectsDir,
		Leader:           w.leader,
		LocalLeader:      w.localLeader,
		ConfigMode:       w.configMode,
		InstallNeovim:    w.installNeovim,
		InstallRipgrep:   w.installRipgrep,
		InstallFd:        w.installFd,
		WriteNvimConfig:  w.writeConfig,
		EnableTree:       w.enableTree,
		EnableTelescope:  w.enableTelescope,
		EnableLSP:        w.enableLSP,
		EnableJava:       w.enableJava,
		RunLazySync:      w.runLazySync,
		RequireChecksums: w.requireChecksums,
	}
	p.Normalize()
	return p
}

func (w *Wizard) saveProfile() error {
	p := w.buildProfile()
	return profile.Save(p)
}

func (w *Wizard) Run() error {
	w.pages.AddPage("welcome", w.welcomePage(), true, true)
	w.pages.AddPage("system", w.systemPage(), true, false)
	w.pages.AddPage("features", w.featuresPage(), true, false)
	w.pages.AddPage("summary", w.summaryPage(), true, false)
	w.pages.AddPage("install", w.installPage(), true, false)
	w.pages.AddPage("done", w.donePage(""), true, false)

	w.app.SetRoot(w.pages, true)
	w.app.EnableMouse(true)

	w.app.SetInputCapture(func(ev *tcell.EventKey) *tcell.EventKey {
		// Ctrl+C
		if ev.Key() == tcell.KeyCtrlC {
			if w.cancel != nil {
				w.cancel()
			}
			w.app.Stop()
			return nil
		}
		return ev
	})

	return w.app.Run()
}

func (w *Wizard) welcomePage() tview.Primitive {
	tv := tview.NewTextView()
	tv.SetDynamicColors(true)
	tv.SetTextAlign(tview.AlignLeft)
	tv.SetBorder(true)
	tv.SetTitle("nvimwiz")

	fmt.Fprintln(tv, "[::b]Neovim Setup & Config Wizard (nvimwiz)[::-]")
	fmt.Fprintln(tv, "")
	fmt.Fprintln(tv, "This wizard is designed to be reliable even when distro repos are broken.")
	fmt.Fprintln(tv, "")
	fmt.Fprintln(tv, "It can:")
	fmt.Fprintln(tv, "  • Install user-local Neovim / ripgrep / fd from GitHub releases")
	fmt.Fprintln(tv, "  • Manage a generated Neovim config (safe user overrides in lua/nvimwiz/user.lua)")
	fmt.Fprintln(tv, "  • Persist your choices so you can rerun it to update plugins/config")
	fmt.Fprintln(tv, "")
	fmt.Fprintln(tv, "Press Ctrl+C anytime to quit.")
	fmt.Fprintln(tv, "")

	form := tview.NewForm()
	form.AddButton("Continue", func() { w.switchTo("system") })
	form.AddButton("Quit", func() { w.app.Stop() })
	form.SetButtonsAlign(tview.AlignRight)

	flex := tview.NewFlex().SetDirection(tview.FlexRow)
	flex.AddItem(tv, 0, 1, false)
	flex.AddItem(form, 3, 0, true)
	return flex
}

func (w *Wizard) systemPage() tview.Primitive {
	w.systemText = tview.NewTextView()
	w.systemText.SetDynamicColors(true)
	w.systemText.SetBorder(true)
	w.systemText.SetTitle("System info")

	w.refreshSystemInfo()

	form := tview.NewForm()
	form.AddButton("Refresh", func() { w.refreshSystemInfo() })
	form.AddButton("Continue", func() { w.switchTo("features") })
	form.AddButton("Back", func() { w.switchTo("welcome") })
	form.AddButton("Quit", func() { w.app.Stop() })
	form.SetButtonsAlign(tview.AlignRight)

	flex := tview.NewFlex().SetDirection(tview.FlexRow)
	flex.AddItem(w.systemText, 0, 1, false)
	flex.AddItem(form, 3, 0, true)
	return flex
}

func (w *Wizard) refreshSystemInfo() {
	info := sysinfo.Collect()

	var b strings.Builder
	fmt.Fprintf(&b, "[::b]Detected system[::-]\n\n")
	fmt.Fprintf(&b, "GOOS:   %s\n", info.GOOS)
	fmt.Fprintf(&b, "GOARCH: %s\n", info.GOARCH)
	if info.PrettyName != "" {
		fmt.Fprintf(&b, "Distro: %s\n", info.PrettyName)
	}
	if info.ID != "" || info.VersionID != "" {
		fmt.Fprintf(&b, "ID: %s   VERSION_ID: %s\n", info.ID, info.VersionID)
	}
	fmt.Fprintf(&b, "WSL: %v\n", info.WSL)

	if w.localBinNote != "" {
		fmt.Fprintf(&b, "\n[::b]PATH note (this session)[::-]\n")
		fmt.Fprintf(&b, "  %s\n", w.localBinNote)
	}
	if w.profileLoadNote != "" {
		fmt.Fprintf(&b, "\n[::b]Profile note[::-]\n")
		fmt.Fprintf(&b, "  %s\n", w.profileLoadNote)
	}

	fmt.Fprintf(&b, "\n[::b]Package managers found[::-]\n")
	if len(info.PackageManagers) == 0 {
		fmt.Fprintf(&b, "  (none detected)\n")
	} else {
		for _, pm := range info.PackageManagers {
			fmt.Fprintf(&b, "  - %s\n", pm)
		}
	}

	fmt.Fprintf(&b, "\n[::b]Tools on PATH (this session)[::-]\n")
	keys := []string{"git", "curl", "tar", "unzip", "sha256sum", "nvim", "rg", "fd", "node", "python3", "go", "java"}
	for _, k := range keys {
		t := info.Tools[k]
		if t.Present {
			val := t.Version
			if val == "" {
				val = t.Path
			}
			fmt.Fprintf(&b, "  [green]✓[-] %-9s %s\n", k, val)
		} else {
			fmt.Fprintf(&b, "  [red]✗[-] %-9s %s\n", k, t.Error)
		}
	}

	w.systemText.SetText(b.String())
}

func (w *Wizard) featuresPage() tview.Primitive {
	form := tview.NewForm()
	form.SetBorder(true)
	form.SetTitle("Choose features")
	form.SetTitleAlign(tview.AlignLeft)
	setFormLabelWidth(form, 30)

	// Settings
	form.AddInputField("Projects dir", w.projectsDir, 60, nil, func(text string) { w.projectsDir = text })

	form.AddInputField("Leader", encodeKeyForUI(w.leader), 20, nil, func(text string) {
		w.leader = decodeKeyFromUI(text)
	})
	form.AddInputField("Local leader", encodeKeyForUI(w.localLeader), 20, nil, func(text string) {
		w.localLeader = decodeKeyFromUI(text)
	})

	configModes := []string{"Managed (writes init.lua)", "Integrate (does not touch init.lua)"}
	modeIndex := 0
	if strings.ToLower(strings.TrimSpace(w.configMode)) == "integrate" {
		modeIndex = 1
	}
	form.AddDropDown("Config mode", configModes, modeIndex, func(option string, index int) {
		if index == 1 {
			w.configMode = "integrate"
		} else {
			w.configMode = "managed"
		}
	})

	// Installs
	form.AddCheckbox("Install Neovim", w.installNeovim, func(v bool) { w.installNeovim = v })
	form.AddCheckbox("Install ripgrep (rg)", w.installRipgrep, func(v bool) { w.installRipgrep = v })
	form.AddCheckbox("Install fd", w.installFd, func(v bool) { w.installFd = v })

	// Config + features
	form.AddCheckbox("Write/update Neovim config", w.writeConfig, func(v bool) { w.writeConfig = v })
	form.AddCheckbox("File tree (nvim-tree)", w.enableTree, func(v bool) { w.enableTree = v })
	form.AddCheckbox("Telescope", w.enableTelescope, func(v bool) { w.enableTelescope = v })
	form.AddCheckbox("LSP baseline (Mason + lspconfig)", w.enableLSP, func(v bool) { w.enableLSP = v })
	form.AddCheckbox("Java (nvim-jdtls)", w.enableJava, func(v bool) { w.enableJava = v })
	form.AddCheckbox("Run :Lazy sync headless", w.runLazySync, func(v bool) { w.runLazySync = v })
	form.AddCheckbox("Require checksums", w.requireChecksums, func(v bool) { w.requireChecksums = v })

	form.AddButton("Save", func() {
		if err := w.saveProfile(); err != nil {
			w.showError("Save failed", err.Error(), nil)
			return
		}
		w.showInfo("Saved", "Saved profile to:\n"+w.profilePath)
	})
	form.AddButton("Next", func() {
		_ = w.saveProfile()
		w.switchTo("summary")
	})
	form.AddButton("Back", func() { w.switchTo("system") })
	form.AddButton("Quit", func() { w.app.Stop() })
	form.SetButtonsAlign(tview.AlignRight)

	help := tview.NewTextView()
	help.SetDynamicColors(true)
	help.SetBorder(true)
	help.SetTitle("Notes")
	fmt.Fprintln(help, "• Installs go to [::b]~/.local/bin[::-] and [::b]~/.local/nvim[::-].")
	fmt.Fprintln(help, "• Config is managed via [::b]lua/nvimwiz/generated/*[-::-]. Your file is [::b]lua/nvimwiz/user.lua[-::-].")
	fmt.Fprintln(help, "• You can rerun nvimwiz to change options; it rewrites only generated files.")
	if w.profilePath != "" {
		fmt.Fprintln(help, "• Profile: [::b]"+w.profilePath+"[-::-]")
	}
	fmt.Fprintln(help, "")
	fmt.Fprintln(help, "Config mode:")
	fmt.Fprintln(help, "  • Managed: nvimwiz writes init.lua (backs up your old init.lua if needed).")
	fmt.Fprintln(help, "  • Integrate: nvimwiz does not touch init.lua. You add: require(\"nvimwiz.loader\")")

	flex := tview.NewFlex()
	flex.AddItem(form, 0, 2, true)
	flex.AddItem(help, 0, 1, false)
	return flex
}

func setFormLabelWidth(form *tview.Form, width int) {
	type labelWidther interface {
		SetLabelWidth(int) *tview.Form
	}
	if f, ok := interface{}(form).(labelWidther); ok {
		f.SetLabelWidth(width)
	}
}

func (w *Wizard) summaryPage() tview.Primitive {
	tv := tview.NewTextView()
	tv.SetDynamicColors(true)
	tv.SetBorder(true)
	tv.SetTitle("Summary")

	opts := w.currentOptions()
	plan := tasks.Plan(opts)

	fmt.Fprintln(tv, "[::b]Planned steps[::-]")
	fmt.Fprintln(tv, "")
	for i, t := range plan {
		fmt.Fprintf(tv, "  %2d. %s\n", i+1, t.Name)
	}
	fmt.Fprintln(tv, "")
	fmt.Fprintf(tv, "Config mode: %s\n", strings.ToLower(strings.TrimSpace(w.configMode)))
	fmt.Fprintf(tv, "Leader: %s   Local leader: %s\n", encodeKeyForUI(w.leader), encodeKeyForUI(w.localLeader))
	fmt.Fprintln(tv, "")
	fmt.Fprintln(tv, "Press Install to run, or Back to adjust options.")

	form := tview.NewForm()
	form.AddButton("Install", func() { w.startInstall() })
	form.AddButton("Back", func() { w.switchTo("features") })
	form.AddButton("Quit", func() { w.app.Stop() })
	form.SetButtonsAlign(tview.AlignRight)

	flex := tview.NewFlex().SetDirection(tview.FlexRow)
	flex.AddItem(tv, 0, 1, false)
	flex.AddItem(form, 3, 0, true)
	return flex
}

func (w *Wizard) installPage() tview.Primitive {
	w.progressView = tview.NewTextView()
	w.progressView.SetDynamicColors(true)
	w.progressView.SetBorder(true)
	w.progressView.SetTitle("Progress")
	w.progressView.SetText("Waiting...")

	w.logView = tview.NewTextView()
	w.logView.SetDynamicColors(true)
	w.logView.SetBorder(true)
	w.logView.SetTitle("Log")
	w.logView.SetScrollable(true)
	w.logView.SetChangedFunc(func() { w.app.Draw() })

	form := tview.NewForm()
	form.AddButton("Cancel", func() {
		if w.cancel != nil {
			w.cancel()
		}
	})
	form.AddButton("Quit", func() {
		if w.cancel != nil {
			w.cancel()
		}
		w.app.Stop()
	})
	form.SetButtonsAlign(tview.AlignRight)

	flex := tview.NewFlex().SetDirection(tview.FlexRow)
	flex.AddItem(w.progressView, 3, 0, false)
	flex.AddItem(w.logView, 0, 1, true)
	flex.AddItem(form, 3, 0, false)
	return flex
}

func (w *Wizard) donePage(summary string) tview.Primitive {
	tv := tview.NewTextView()
	tv.SetDynamicColors(true)
	tv.SetBorder(true)
	tv.SetTitle("Done")

	fmt.Fprintln(tv, "[green]Completed.[-]")
	if summary != "" {
		fmt.Fprintln(tv, "")
		fmt.Fprintln(tv, summary)
	}
	fmt.Fprintln(tv, "")
	fmt.Fprintln(tv, "Next steps:")
	fmt.Fprintln(tv, "  1) Ensure ~/.local/bin is on PATH (for your shell)")
	fmt.Fprintln(tv, "  2) Run: nvim")
	fmt.Fprintln(tv, "  3) In Neovim: :Lazy sync  then  :Mason")
	fmt.Fprintln(tv, "")
	fmt.Fprintln(tv, "Tip: You can rerun nvimwiz anytime to change options and rewrite generated config.")

	form := tview.NewForm()
	form.AddButton("Exit", func() { w.app.Stop() })
	form.AddButton("Back to start", func() { w.switchTo("welcome") })
	form.SetButtonsAlign(tview.AlignRight)

	flex := tview.NewFlex().SetDirection(tview.FlexRow)
	flex.AddItem(tv, 0, 1, false)
	flex.AddItem(form, 3, 0, true)
	return flex
}

func (w *Wizard) currentOptions() tasks.Options {
	enableLSP := w.enableLSP || w.enableJava

	return tasks.Options{
		ProjectsDir:      w.projectsDir,
		Leader:           w.leader,
		LocalLeader:      w.localLeader,
		ConfigMode:       w.configMode,
		InstallNeovim:    w.installNeovim,
		InstallRipgrep:   w.installRipgrep,
		InstallFd:        w.installFd,
		WriteNvimConfig:  w.writeConfig,
		EnableTree:       w.enableTree,
		EnableTelescope:  w.enableTelescope,
		EnableLSP:        enableLSP,
		EnableJava:       w.enableJava,
		RunLazySync:      w.runLazySync,
		RequireChecksums: w.requireChecksums,
	}
}

func (w *Wizard) startInstall() {
	_ = w.saveProfile()

	opts := w.currentOptions()
	plan := tasks.Plan(opts)

	w.switchTo("install")

	logf := func(line string) {
		w.app.QueueUpdateDraw(func() {
			fmt.Fprintln(w.logView, line)
		})
	}
	onProgress := func(done, total int) {
		w.app.QueueUpdateDraw(func() {
			w.progressView.SetText(fmt.Sprintf("%s  %d/%d", progressBar(done, total, 28), done, total))
		})
	}

	ctx, cancel := context.WithCancel(context.Background())
	w.cancel = cancel

	go func() {
		start := time.Now()
		logf(time.Now().Format(time.RFC3339) + " Starting")
		err := tasks.RunAll(ctx, plan, logf, onProgress)
		if err != nil {
			w.app.QueueUpdateDraw(func() {
				w.showError("Install failed", err.Error(), func() { w.switchTo("features") })
			})
			return
		}

		elapsed := time.Since(start).Round(10 * time.Millisecond)
		w.app.QueueUpdateDraw(func() {
			w.pages.RemovePage("done")
			w.pages.AddPage("done", w.donePage("Elapsed: "+elapsed.String()), true, false)
			w.switchTo("done")
		})
	}()
}

func (w *Wizard) showError(title, msg string, onOK func()) {
	modal := tview.NewModal()
	modal.SetTitle(title)
	modal.SetText(msg)
	modal.AddButtons([]string{"OK"})
	modal.SetDoneFunc(func(_ int, _ string) {
		w.pages.RemovePage("modal")
		if onOK != nil {
			onOK()
		}
	})
	w.pages.AddPage("modal", modal, true, true)
}

func (w *Wizard) showInfo(title, msg string) {
	modal := tview.NewModal()
	modal.SetTitle(title)
	modal.SetText(msg)
	modal.AddButtons([]string{"OK"})
	modal.SetDoneFunc(func(_ int, _ string) {
		w.pages.RemovePage("modal")
	})
	w.pages.AddPage("modal", modal, true, true)
}

func (w *Wizard) switchTo(name string) {
	w.pages.SwitchToPage(name)
}

func progressBar(done, total, width int) string {
	if total <= 0 {
		return "[" + strings.Repeat("-", width) + "]"
	}
	if done < 0 {
		done = 0
	}
	if done > total {
		done = total
	}
	filled := int(float64(done) / float64(total) * float64(width))
	if filled > width {
		filled = width
	}
	return "[" + strings.Repeat("#", filled) + strings.Repeat("-", width-filled) + "]"
}

// encodeKeyForUI makes leader keys visible/editable in a single-line input.
// We represent a literal space as <Space> to avoid "blank" confusion.
func encodeKeyForUI(s string) string {
	if s == "" {
		return "<Space>"
	}
	if s == " " {
		return "<Space>"
	}
	return s
}

func decodeKeyFromUI(s string) string {
	st := strings.TrimSpace(s)
	if st == "" {
		return " "
	}
	low := strings.ToLower(st)
	switch low {
	case "<space>", "space":
		return " "
	case "<tab>", "tab":
		return "\t"
	}
	return st
}
