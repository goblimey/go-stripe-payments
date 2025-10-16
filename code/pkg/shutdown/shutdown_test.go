package shutdown

import (
	"testing"
	"time"
)

// TestTimeToShutdown checks the timeToShutdown function.
func TestTimeToShutdown(t *testing.T) {

	london, tzErr := time.LoadLocation("Europe/London")
	if tzErr != nil {
		t.Error(tzErr)
	}

	var testData = []struct {
		description string
		now         time.Time
		want        time.Duration
	}{
		{"start of day", time.Date(2025, time.February, 14, 0, 0, 0, 0, london),
			(time.Hour * 24) - (time.Second)},
		{"noon", time.Date(2025, time.February, 14, 12, 0, 0, 0, london),
			(time.Hour * 12) - (time.Second)},
		{"late", time.Date(2025, time.February, 14, 23, 59, 0, 0, london),
			(time.Second * 59)},
		// The clocks moved forward on Sunday 30th March at 1am so that day was only
		// 23 hours long.
		{"daylight saving", time.Date(2025, time.March, 30, 0, 0, 0, 0, london),
			(time.Hour * 23) - (time.Second)},
	}

	for _, td := range testData {

		got := timeToShutdown(td.now)
		if got != td.want {
			t.Errorf("%s - want %v got %v", td.description, td.want, got)
			continue
		}
	}
}
