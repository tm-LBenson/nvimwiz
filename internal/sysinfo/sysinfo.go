package sysinfo

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

type ToolInfo struct {
	Present bool
	Path    string
	Version string
	Error   string
}

type Info struct {
	GOOS            string
	GOARCH          string
	WSL             bool
	PrettyName      string
	ID              string
	VersionID       string
	PackageManagers []string
	Tools           map[string]ToolInfo
}

func Collect() Info {
	i := Info{
		GOOS:   runtime.GOOS,
		GOARCH: runtime.GOARCH,
		WSL:    detectWSL(),
		Tools:  map[string]ToolInfo{},
	}

	readOSRelease(&i)

	i.PackageManagers = detectPackageManagers()
	for _, tool := range []string{"git", "curl", "tar", "unzip", "sha256sum", "nvim", "rg", "fd", "node", "python3", "go", "java"} {
		i.Tools[tool] = detectTool(tool)
	}

	return i
}

func detectWSL() bool {
	if os.Getenv("WSL_INTEROP") != "" || os.Getenv("WSL_DISTRO_NAME") != "" {
		return true
	}
	// /proc is Linux-only, but it's fine to just return false if missing
	b, err := os.ReadFile("/proc/sys/kernel/osrelease")
	if err == nil && strings.Contains(strings.ToLower(string(b)), "microsoft") {
		return true
	}
	return false
}

func readOSRelease(i *Info) {
	b, err := os.ReadFile("/etc/os-release")
	if err != nil {
		return
	}
	lines := strings.Split(string(b), "\n")
	for _, ln := range lines {
		ln = strings.TrimSpace(ln)
		if ln == "" || strings.HasPrefix(ln, "#") {
			continue
		}
		kv := strings.SplitN(ln, "=", 2)
		if len(kv) != 2 {
			continue
		}
		k := kv[0]
		v := strings.Trim(kv[1], `"'`)
		switch k {
		case "PRETTY_NAME":
			i.PrettyName = v
		case "ID":
			i.ID = v
		case "VERSION_ID":
			i.VersionID = v
		}
	}
}

func detectPackageManagers() []string {
	var pms []string
	for _, pm := range []string{"dnf", "apt-get", "pacman", "zypper", "brew"} {
		if _, err := exec.LookPath(pm); err == nil {
			pms = append(pms, pm)
		}
	}
	return pms
}

func detectTool(name string) ToolInfo {
	path, err := exec.LookPath(name)
	if err != nil {
		return ToolInfo{Present: false, Error: "not found"}
	}
	ti := ToolInfo{Present: true, Path: path}

	// best-effort version lines
	switch name {
	case "git":
		ti.Version = runFirstLine(name, "--version")
	case "curl":
		ti.Version = runFirstLine(name, "--version")
	case "tar":
		ti.Version = runFirstLine(name, "--version")
	case "unzip":
		// unzip -v writes multiple lines
		ti.Version = runFirstLine(name, "-v")
	case "sha256sum":
		ti.Version = runFirstLine(name, "--version")
	case "nvim":
		ti.Version = runFirstLine(name, "--version")
	case "rg":
		ti.Version = runFirstLine(name, "--version")
	case "fd":
		ti.Version = runFirstLine(name, "--version")
	case "node":
		ti.Version = runFirstLine(name, "--version")
	case "python3":
		ti.Version = runFirstLine(name, "--version")
	case "go":
		ti.Version = runFirstLine(name, "version")
	case "java":
		// java -version writes to stderr; capture both
		ti.Version = runFirstLineStderrOk(name, "-version")
	default:
		ti.Version = filepath.Base(path)
	}

	// trim to something short-ish
	ti.Version = strings.TrimSpace(ti.Version)
	return ti
}

func runFirstLine(bin string, args ...string) string {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, bin, args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	_ = cmd.Run()
	line := strings.SplitN(out.String(), "\n", 2)[0]
	return strings.TrimSpace(line)
}

func runFirstLineStderrOk(bin string, args ...string) string {
	return runFirstLine(bin, args...)
}
