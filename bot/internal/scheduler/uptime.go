package scheduler

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"bot/internal/serverping"
)

const pingInterval = 10 * time.Second

func StartUptimePingScheduler(startTime time.Time, ctx context.Context) {
	ticker := time.NewTicker(pingInterval)

	go func() {
		for {
			select {
			case <-ticker.C:
				fmt.Println("Time now: " + time.Now().Format("12:00:00"))
				timeSinceStart := time.Since(startTime)
				fmt.Println("Time since start: " + strconv.FormatInt(timeSinceStart.Milliseconds(), 10))
				serverping.SendUptime(timeSinceStart.Milliseconds())
			case <-ctx.Done():
				ticker.Stop()
				fmt.Println("Scheduler shutting down.")
				return
			}
		}
	}()
}
