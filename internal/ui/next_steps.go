package ui

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/rivo/tview"

	"nvimwiz/internal/env"
	"nvimwiz/internal/nvimcfg"
)

func (w *Wizard) showNextSteps() {
	if w == nil || w.app == nil || w.pages == nil {
		return
	}

	w.pages.RemovePage("next_steps")

	target := strings.ToLower(strings.TrimSpace(w.p.Target))
	safe := target != "default"
	appName := w.p.EffectiveAppName()
	cfgDir, _ := nvimcfg.ConfigDirForProfile(w.p)
	localBin, _ := env.LocalBin()

	canLauncher := safe && runtime.GOOS != "windows"
	launcherName := ""
	launcherPath := ""
	launcherExists := false
	if canLauncher && appName != "" && appName != "nvim" {
		launcherName = appName
		if localBin != "" {
			launcherPath = filepath.Join(localBin, launcherName)
			if st, err := os.Stat(launcherPath); err == nil && !st.IsDir() {
				launcherExists = true
			}
		}
	}

	render := func(status string) string {
		lines := []string{}
		if safe {
			lines = append(lines, "You are using a safe build.")
		} else {
			lines = append(lines, "You are managing your system Neovim config.")
		}
		if cfgDir != "" {
			lines = append(lines, "Config dir: "+cfgDir)
		}

		lines = append(lines, "")

		if safe {
			lines = append(lines, "Launch:")
			if launcherExists {
				lines = append(lines, "  "+launcherName)
				if launcherPath != "" {
					lines = append(lines, "  (launcher: "+launcherPath+")")
				}
			} else {
				lines = append(lines, "  NVIM_APPNAME="+appName+" nvim")
				if canLauncher && launcherPath != "" {
					lines = append(lines, "")
					lines = append(lines, "Recommended:")
					lines = append(lines, "  Create a launcher command so you can run: "+launcherName)
					lines = append(lines, "  It will be written to: "+launcherPath)
					if localBin != "" {
						lines = append(lines, "  If the command is not found, add "+localBin+" to PATH.")
					}
				}
			}
		} else {
			lines = append(lines, "Launch:")
			lines = append(lines, "  nvim")
			if strings.ToLower(strings.TrimSpace(w.p.ConfigMode)) == "integrate" {
				lines = append(lines, "")
				lines = append(lines, "Integrate mode:")
				lines = append(lines, "  Add require(\"nvimwiz.loader\") to your init.lua")
			}
		}

		if status != "" {
			lines = append(lines, "")
			lines = append(lines, status)
		}
		return strings.Join(lines, "\n")
	}

	modal := tview.NewModal()
	modal.SetTitle("Next steps")
	modal.SetText(render(""))

	buttons := []string{"Close"}
	if canLauncher && launcherName != "" {
		buttons = []string{"Create launcher", "Close"}
	}
	modal.AddButtons(buttons)
	modal.SetDoneFunc(func(_ int, label string) {
		if label == "Close" {
			w.pages.RemovePage("next_steps")
			if w.applyButtons != nil {
				w.app.SetFocus(w.applyButtons)
			} else {
				w.app.SetFocus(w.pages)
			}
			return
		}
		if label == "Create launcher" {
			path, err := env.CreateNvimAppLauncher(appName)
			if err != nil {
				modal.SetText(render("Launcher error: " + err.Error()))
				w.app.SetFocus(modal)
				return
			}
			launcherExists = true
			launcherPath = path
			modal.SetText(render("Launcher created: " + path))
			w.app.SetFocus(modal)
		}
	})

	w.pages.AddPage("next_steps", modal, true, true)
	w.app.SetFocus(modal)
}
