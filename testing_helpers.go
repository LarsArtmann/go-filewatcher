package filewatcher

import (
	"time"
)

func testEvent(path string, op Op) Event {
	return Event{Path: path, Op: op, Timestamp: time.Now(), IsDir: false}
}

func testWriteEvent(path string) Event {
	return testEvent(path, Write)
}

func fixedTimeEvent(path string, op Op, hour int) Event {
	return Event{
		Path:      path,
		Op:        op,
		Timestamp: time.Date(2025, 1, 1, hour, 0, 0, 0, time.UTC),
		IsDir:     false,
	}
}
