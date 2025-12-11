package scheduler

import (
	"context"
	"fmt"
	"time"

	"bot/internal/serverping"
)

const pingInterval = 15 * time.Second

func StartUptimePingScheduler(startTime time.Time, ctx context.Context) {
	ticker := time.NewTicker(pingInterval)

	go func() {
		for {
			select {
			case <-ticker.C:
				fmt.Println("Time now: " + time.Now().Format("12:00:00"))
				timeSinceStart := time.Since(startTime).String()
				fmt.Println("Time since start: " + timeSinceStart)
				serverping.SendUptime(timeSinceStart)
			case <-ctx.Done():
				ticker.Stop()
				fmt.Println("Scheduler shutting down.")
				return
			}
		}
	}()
}
