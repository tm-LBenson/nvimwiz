package env

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func CreateNvimAppLauncher(appName string) (string, error) {
	appName = strings.TrimSpace(appName)
	if appName == "" {
		return "", fmt.Errorf("app name is empty")
	}
	if runtime.GOOS == "windows" {
		return "", fmt.Errorf("launcher is not supported on windows")
	}

	lb, err := LocalBin()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(lb, 0o755); err != nil {
		return "", err
	}

	path := filepath.Join(lb, appName)
	if fi, err := os.Stat(path); err == nil && fi.IsDir() {
		return "", fmt.Errorf("launcher path is a directory")
	}

	script := "#!/usr/bin/env sh\nexport NVIM_APPNAME=\"" + appName + "\"\nexec nvim \"$@\"\n"
	if err := os.WriteFile(path, []byte(script), 0o755); err != nil {
		return "", err
	}
	_ = os.Chmod(path, 0o755)
	return path, nil
}
