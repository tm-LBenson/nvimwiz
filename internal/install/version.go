package install

import (
	"context"
	"os/exec"
	"strings"
	"time"
)

func normalizeVersion(v string) string {
	v = strings.TrimSpace(v)
	v = strings.TrimPrefix(v, "v")
	return v
}

func installedCommandVersion(ctx context.Context, command string, args ...string) (version string, path string, ok bool) {
	path, err := exec.LookPath(command)
	if err != nil {
		return "", "", false
	}

	if ctx == nil {
		ctx = context.Background()
	}
	if _, has := ctx.Deadline(); !has {
		cctx, cancel := context.WithTimeout(ctx, 2*time.Second)
		defer cancel()
		ctx = cctx
	}

	cmd := exec.CommandContext(ctx, path, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", path, false
	}

	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) >= 2 {
			return normalizeVersion(fields[1]), path, true
		}
		break
	}

	return "", path, false
}
