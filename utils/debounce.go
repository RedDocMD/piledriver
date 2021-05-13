package utils

import (
	"time"
)

// DebounceEvents removes identical events less than some threshold.
func DebounceEvents(input, output chan Event) {
	const threshold = 500 // ms

	// There is one map for each Category of Event
	var lastEvent [MaxEventCategory]map[string]time.Time
	for i := range lastEvent {
		lastEvent[i] = make(map[string]time.Time)
	}

	for event := range input {
		cat := event.Category
		path := event.Path
		currTime := event.Timestamp
		lastTime, ok := lastEvent[cat][path]
		if !ok {
			lastEvent[cat][path] = currTime
			output <- event
		} else {
			timeGap := currTime.Sub(lastTime)
			if timeGap.Milliseconds() > threshold {
				lastEvent[cat][path] = currTime
				output <- event
			}
		}
	}
}
