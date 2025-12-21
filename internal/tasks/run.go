package tasks

import (
	"context"
	"fmt"
	"time"
)

func RunAll(ctx context.Context, tasks []Task, logf func(string), onProgress func(done, total int)) error {
	if onProgress == nil {
		onProgress = func(int, int) {}
	}
	total := len(tasks)
	done := 0
	onProgress(done, total)

	for _, t := range tasks {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		logf("")
		logf(fmt.Sprintf("[INFO] %s", t.Name))
		start := time.Now()

		if err := t.Run(ctx, logf); err != nil {
			logf(fmt.Sprintf("[ERR] %s: %v", t.Name, err))
			return err
		}

		logf(fmt.Sprintf("[OK] %s (%.2fs)", t.Name, time.Since(start).Seconds()))
		done++
		onProgress(done, total)
	}

	return nil
}
