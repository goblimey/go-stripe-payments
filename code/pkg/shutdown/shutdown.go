package shutdown

import (
	"os"
	"time"
)

// timeToShutdown uses the given start time to work out the duration to wait before shutting
// down the app.  In production the given time should be the current time.  In test any time
// can be supplied.
func timeToShutdown(startTime time.Time) time.Duration {
	// Figure out the duration from now to one second before midnight at the end of today.
	// There's an edge case - if we are already within one second of midnight we should find the
	// duration to midnight at the end of tomorrow.

	shutdownTime := time.Date(startTime.Year(), startTime.Month(), startTime.Day(), 23, 59, 59, 0, startTime.Location())
	if shutdownTime.Before(startTime) {
		// It's close to midnight.  Pause until the end of the next day.
		shutdownTime = shutdownTime.AddDate(0, 0, 1)
	}

	// Calculate the duration to shutdown time.
	duration := shutdownTime.Sub(startTime)

	return duration

}

// pauseAndShutdown waits until jut before midnight and then shuts down the app.  It's
// intended that it runs as a goroutine.
func PauseAndShutdown(now time.Time) {
	d := timeToShutdown(now)
	time.Sleep(d)
	os.Exit(0)
}
