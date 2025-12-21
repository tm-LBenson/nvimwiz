package tasks

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

func runLazySync(ctx context.Context, logf func(string)) error {
	nvimPath, ok := lookPath("nvim")
	if !ok {
		logf("[WARN] nvim not found on PATH; skipping :Lazy sync")
		return nil
	}

	// Run with a reasonable timeout.
	ctx2, cancel := context.WithTimeout(ctx, 8*time.Minute)
	defer cancel()

	cmd := exec.CommandContext(ctx2, nvimPath, "--headless", "+Lazy sync", "+qa")
	out, err := cmd.CombinedOutput()
	if len(out) > 0 {
		for _, ln := range strings.Split(strings.TrimRight(string(out), "\n"), "\n") {
			logf("[NVIM] " + ln)
		}
	}
	if err != nil {
		logf(fmt.Sprintf("[WARN] :Lazy sync failed (continuing): %v", err))
		return nil
	}
	logf("[OK] :Lazy sync completed")
	return nil
}
