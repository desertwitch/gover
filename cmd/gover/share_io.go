package main

import (
	"context"
	"fmt"
	"runtime"

	"github.com/desertwitch/gover/internal/generic/queue"
)

func (app *App) IO(ctx context.Context) error {
	tasker := queue.NewTaskManager()

	queues := app.queueManager.IOManager.GetQueues()

	for _, targetQueue := range queues {
		tasker.Add(
			func(targetQueue *queue.IOTargetQueue) func() {
				return func() {
					_ = app.ioHandler.ProcessTargetQueue(ctx, targetQueue)
				}
			}(targetQueue),
		)
	}

	// go app.IOProgress(ctx)
	if err := tasker.LaunchConcAndWait(ctx, runtime.NumCPU()); err != nil {
		return fmt.Errorf("(app-io) %w", err)
	}

	return nil
}

/* func (app *App) IOProgress(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second) //nolint:mnd
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			progress := app.queueManager.IOManager.Progress()

			// Format time values
			var startTime, etaTime string
			if !progress.StartTime.IsZero() {
				startTime = progress.StartTime.Format("15:04:05")
			} else {
				startTime = "N/A"
			}

			if !progress.ETA.IsZero() {
				etaTime = progress.ETA.Format("15:04:05")
			} else {
				etaTime = "N/A"
			}

			// Calculate time left in minutes
			timeLeftMin := 0.0
			if progress.TimeLeft > 0 {
				timeLeftMin = float64(progress.TimeLeft.Seconds()) / 60 //nolint:mnd
			}

			// Print the progress information
			fmt.Printf("Progress: %.2f%% (%d/%d)\n"+
				"Items: InProgress=%d, Success=%d, Skipped=%d\n"+
				"Time: Started=%v, ETA=%v (%.1f%s left)\n"+
				"Speed: %s\n",
				progress.ProgressPct,
				progress.ProcessedItems,
				progress.TotalItems,
				progress.InProgressItems,
				progress.SuccessItems,
				progress.SkippedItems,
				startTime,
				etaTime,
				timeLeftMin, "min",
				humanize.Bytes(uint64(progress.TransferSpeed))+" / second",
			)

			// If transfer is finished, print completion message and return
			if !progress.IsStarted && progress.FinishTime.After(progress.StartTime) {
				fmt.Printf("Transfer completed at %s. Total time: %s\n",
					progress.FinishTime.Format("15:04:05"),
					progress.FinishTime.Sub(progress.StartTime).Round(time.Second))

				return
			}
		}
	}
}
*/
