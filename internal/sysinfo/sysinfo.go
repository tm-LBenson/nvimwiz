package sysinfo

import (
	"bufio"
	"bytes"
	"os"
	"os/exec"
	"runtime"
	"strings"
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
	PrettyName      string
	ID              string
	VersionID       string
	WSL             bool
	PackageManagers []string
	Tools           map[string]ToolInfo
}

func Collect() Info {
	info := Info{
		GOOS:   runtime.GOOS,
		GOARCH: runtime.GOARCH,
		WSL:    isWSL(),
		Tools:  map[string]ToolInfo{},
	}
	readOSRelease(&info)
	info.PackageManagers = detectPackageManagers()
	for _, k := range []string{"git", "curl", "tar", "unzip", "sha256sum", "nvim", "rg", "fd", "node", "python3", "go", "java"} {
		info.Tools[k] = detectTool(k)
	}
	return info
}

func detectPackageManagers() []string {
	candidates := []string{"dnf", "apt", "pacman", "zypper", "brew"}
	out := []string{}
	for _, c := range candidates {
		if _, err := exec.LookPath(c); err == nil {
			out = append(out, c)
		}
	}
	return out
}

func detectTool(name string) ToolInfo {
	p, err := exec.LookPath(name)
	if err != nil {
		return ToolInfo{Present: false, Error: err.Error()}
	}
	ti := ToolInfo{Present: true, Path: p}
	cmd := versionCmd(name)
	if len(cmd) > 0 {
		b, err := exec.Command(cmd[0], cmd[1:]...).CombinedOutput()
		if err == nil {
			ti.Version = strings.TrimSpace(firstLine(b))
		}
	}
	return ti
}

func versionCmd(name string) []string {
	switch name {
	case "nvim":
		return []string{"nvim", "--version"}
	case "rg":
		return []string{"rg", "--version"}
	case "fd":
		return []string{"fd", "--version"}
	case "git":
		return []string{"git", "--version"}
	case "go":
		return []string{"go", "version"}
	case "python3":
		return []string{"python3", "--version"}
	case "node":
		return []string{"node", "--version"}
	case "java":
		return []string{"java", "-version"}
	default:
		return nil
	}
}

func firstLine(b []byte) string {
	b = bytes.ReplaceAll(b, []byte("\r\n"), []byte("\n"))
	s := string(b)
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		return s[:i]
	}
	return s
}

func readOSRelease(info *Info) {
	f, err := os.Open("/etc/os-release")
	if err != nil {
		return
	}
	defer f.Close()
	m := map[string]string{}
	s := bufio.NewScanner(f)
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if line == "" {
			continue
		}
		idx := strings.Index(line, "=")
		if idx <= 0 {
			continue
		}
		k := line[:idx]
		v := strings.Trim(line[idx+1:], "\" ")
		m[k] = v
	}
	info.PrettyName = m["PRETTY_NAME"]
	info.ID = m["ID"]
	info.VersionID = m["VERSION_ID"]
}

func isWSL() bool {
	if os.Getenv("WSL_DISTRO_NAME") != "" || os.Getenv("WSL_INTEROP") != "" {
		return true
	}
	b, err := os.ReadFile("/proc/version")
	if err != nil {
		return false
	}
	low := strings.ToLower(string(b))
	return strings.Contains(low, "microsoft") || strings.Contains(low, "wsl")
}
