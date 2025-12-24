package ui

import (
	"sort"
	"strings"

	"github.com/rivo/tview"
)

func (w *Wizard) systemInfoView() tview.Primitive {
	tv := tview.NewTextView()
	tv.SetDynamicColors(true)
	tv.SetBorder(true)
	tv.SetTitle("System")

	lines := []string{}

	if strings.TrimSpace(w.envNote) != "" {
		lines = append(lines, w.envNote)
		lines = append(lines, "")
	}

	if strings.TrimSpace(w.sys.PrettyName) != "" {
		lines = append(lines, "OS: "+w.sys.PrettyName)
	} else if strings.TrimSpace(w.sys.GOOS) != "" {
		lines = append(lines, "OS: "+w.sys.GOOS)
	}

	if strings.TrimSpace(w.sys.GOARCH) != "" {
		lines = append(lines, "Arch: "+w.sys.GOARCH)
	}

	if w.sys.WSL {
		lines = append(lines, "WSL: yes")
	} else {
		lines = append(lines, "WSL: no")
	}

	if len(w.sys.PackageManagers) > 0 {
		lines = append(lines, "Package managers: "+strings.Join(w.sys.PackageManagers, ", "))
	}

	if len(w.sys.Tools) > 0 {
		toolNames := make([]string, 0, len(w.sys.Tools))
		for name := range w.sys.Tools {
			toolNames = append(toolNames, name)
		}
		sort.Strings(toolNames)

		for _, name := range toolNames {
			tool := w.sys.Tools[name]
			if !tool.Present {
				continue
			}
			value := tool.Version
			if strings.TrimSpace(value) == "" {
				value = tool.Path
			}
			if strings.TrimSpace(value) == "" {
				continue
			}
			lines = append(lines, name+": "+value)
		}
	}

	if len(lines) == 0 {
		lines = append(lines, "System info not available")
	}

	tv.SetText(strings.Join(lines, "\n"))
	return tv
}
