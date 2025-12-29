package ui

import (
	"context"
	"os/exec"
	"sort"
	"strings"
	"time"

	"nvimwiz/internal/install"
)

func (w *Wizard) refreshInstallStatusAsync() {
	if w.installStatusRunning {
		return
	}
	if !w.installStatusLast.IsZero() && time.Since(w.installStatusLast) < 5*time.Second {
		return
	}

	w.installStatusRunning = true
	w.installStatusLast = time.Now()

	ids := w.installFeatureIDs()
	go func(ids []string) {
		res := map[string]install.ToolStatus{}
		for _, id := range ids {
			ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
			st, ok := install.StatusForFeature(ctx, id)
			cancel()
			if ok {
				res[id] = st
			}
		}

		w.app.QueueUpdateDraw(func() {
			if w.installStatus == nil {
				w.installStatus = map[string]install.ToolStatus{}
			}
			for id, st := range res {
				w.installStatus[id] = st
			}
			w.installStatusRunning = false
			if w.currentCategory == "Install" {
				w.renderFeatureTable()
			}
		})
	}(ids)
}

func (w *Wizard) installFeatureIDs() []string {
	ids := []string{}
	for id, f := range w.cat.Features {
		if strings.EqualFold(f.Category, "Install") {
			ids = append(ids, id)
		}
	}
	sort.Strings(ids)
	return ids
}

func (w *Wizard) installEnabledLabel(featureID string) string {
	if w.installStatus != nil {
		if st, ok := w.installStatus[featureID]; ok {
			if !st.Present {
				return "Install"
			}
			if st.LatestOK && st.CurrentOK && strings.TrimSpace(st.CurrentVersion) != "" && strings.TrimSpace(st.LatestVersion) != "" && st.CurrentVersion != st.LatestVersion {
				return "Update"
			}
			return "Installed"
		}
	}

	cmd := installCommandForFeature(featureID)
	if cmd == "" {
		return "Install"
	}
	if _, err := exec.LookPath(cmd); err != nil {
		return "Install"
	}
	return "Installed"
}

func (w *Wizard) installDetailsLines(featureID string) []string {
	enabled := w.p.Features[featureID]
	labelEnabled := w.installEnabledLabel(featureID)

	present := false
	path := ""
	cur := ""
	curOK := false
	latest := ""
	latestOK := false
	err := ""

	if w.installStatus != nil {
		if st, ok := w.installStatus[featureID]; ok {
			present = st.Present
			path = st.Path
			cur = st.CurrentVersion
			curOK = st.CurrentOK
			latest = st.LatestVersion
			latestOK = st.LatestOK
			err = st.Error
		}
	}

	if path == "" {
		cmd := installCommandForFeature(featureID)
		if cmd != "" {
			if p, e := exec.LookPath(cmd); e == nil {
				present = true
				path = p
			}
		}
	}

	curDisp := "unknown"
	if curOK && strings.TrimSpace(cur) != "" {
		curDisp = cur
	}
	latestDisp := "unknown"
	if latestOK && strings.TrimSpace(latest) != "" {
		latestDisp = latest
	}

	lines := []string{}
	if !present || strings.TrimSpace(path) == "" {
		lines = append(lines, "Installed path: not found")
	} else {
		lines = append(lines, "Installed path: "+path)
	}
	lines = append(lines, "Current version: "+curDisp)
	lines = append(lines, "Latest version: "+latestDisp)

	if !latestOK && strings.TrimSpace(err) != "" {
		errLine := strings.TrimSpace(err)
		r := []rune(errLine)
		if len(r) > 120 {
			errLine = string(r[:120]) + "..."
		}
		lines = append(lines, "Latest check: "+errLine)
	}

	apply := ""
	if !enabled {
		apply = "Apply: skipped. This tool will not be installed or updated."
	} else if !present {
		apply = "Apply: will download and install the latest release."
	} else if labelEnabled == "Update" {
		apply = "Apply: will download and install the latest release (update)."
	} else {
		apply = "Apply: will check for the latest release and skip the download if you're already up to date."
	}

	lines = append(lines, "")
	lines = append(lines, apply)
	return lines
}

func installCommandForFeature(featureID string) string {
	switch featureID {
	case "install.neovim":
		return "nvim"
	case "install.ripgrep":
		return "rg"
	case "install.fd":
		return "fd"
	default:
		return ""
	}
}
