package install

import (
	"context"
)

type ToolStatus struct {
	Present        bool
	Path           string
	CurrentVersion string
	CurrentOK      bool
	LatestVersion  string
	LatestTag      string
	LatestOK       bool
	Error          string
}

func StatusForFeature(ctx context.Context, featureID string) (ToolStatus, bool) {
	switch featureID {
	case "install.neovim":
		return status(ctx, "nvim", []string{"--version"}, "neovim", "neovim"), true
	case "install.ripgrep":
		return status(ctx, "rg", []string{"--version"}, "BurntSushi", "ripgrep"), true
	case "install.fd":
		return status(ctx, "fd", []string{"--version"}, "sharkdp", "fd"), true
	default:
		return ToolStatus{}, false
	}
}

func status(ctx context.Context, command string, args []string, owner, repo string) ToolStatus {
	if ctx == nil {
		ctx = context.Background()
	}

	cur, path, ok := installedCommandVersion(ctx, command, args...)
	st := ToolStatus{
		Present:        path != "",
		Path:           path,
		CurrentVersion: cur,
		CurrentOK:      ok,
	}

	rel, err := fetchLatestRelease(ctx, owner, repo)
	if err != nil {
		st.Error = err.Error()
		return st
	}
	st.LatestTag = rel.TagName
	st.LatestVersion = normalizeVersion(rel.TagName)
	st.LatestOK = st.LatestVersion != ""
	return st
}
