package tasks

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"nvimwiz/internal/env"
)

func ensureLocalBin(logf func(string)) error {
	bin, err := env.LocalBin()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(bin, 0o755); err != nil {
		return err
	}
	if changed, _, err := env.EnsureLocalBinInPath(); err == nil && changed {
		logf("[INFO] Added ~/.local/bin to PATH for this session")
	}
	logf(fmt.Sprintf("[INFO] Ensured %s", bin))
	return nil
}

func userPaths() (home string, localBin string, localNvimRoot string, configRoot string, err error) {
	home, err = os.UserHomeDir()
	if err != nil {
		return "", "", "", "", err
	}
	localBin = filepath.Join(home, ".local", "bin")
	localNvimRoot = filepath.Join(home, ".local", "nvim")
	configRoot = filepath.Join(home, ".config", "nvim")
	return
}

func runCmd(ctx context.Context, logf func(string), name string, args ...string) error {
	cmd := exec.CommandContext(ctx, name, args...)
	out, err := cmd.CombinedOutput()
	if len(out) > 0 {
		for _, line := range strings.Split(strings.TrimRight(string(out), "\n"), "\n") {
			logf("[CMD] " + line)
		}
	}
	return err
}

func lookPath(name string) (string, bool) {
	p, err := exec.LookPath(name)
	return p, err == nil
}

func isLinux() bool  { return runtime.GOOS == "linux" }
func isDarwin() bool { return runtime.GOOS == "darwin" }

// copyFile copies src to dst and sets mode if mode!=0.
func copyFile(src, dst string, mode os.FileMode) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() { _ = out.Close() }()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}

	if mode != 0 {
		_ = os.Chmod(dst, mode)
	}
	return nil
}

func timeoutCtx(parent context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(parent, 10*time.Minute)
}

var errUnsupported = errors.New("unsupported platform for this installer")
