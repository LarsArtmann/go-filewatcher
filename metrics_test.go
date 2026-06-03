package filewatcher

import (
	"sync/atomic"
	"testing"
	"time"
)

func TestPrometheusCollector_CountersAndGauges(t *testing.T) {
	t.Parallel()

	var callCount atomic.Int32

	statsFn := func() Stats {
		callCount.Add(1)

		return Stats{
			WatchCount:        42,
			IsWatching:        true,
			IsClosed:          false,
			EventsProcessed:   100,
			EventsFilteredOut: 10,
			ErrorsEncountered: 2,
			WatchErrors:       0,
			Uptime:            5 * time.Second,
			WatchLimit:        8192,
			WatchBudgetUsed:   0.005,
		}
	}

	collector := NewPrometheusCollector(statsFn)

	if collector == nil {
		t.Fatal("expected non-nil collector")
	}

	counters := collector.Counters()
	assertLen(t, "counters", len(counters), 4)

	// Verify a specific counter value
	for _, counter := range counters {
		if counter.Name == "filewatcher_events_processed_total" {
			assertEqual(t, "events_processed_total", counter.Value, uint64(100))
		}

		if counter.Name == "filewatcher_errors_encountered_total" {
			assertEqual(t, "errors_encountered_total", counter.Value, uint64(2))
		}
	}

	gauges := collector.Gauges()
	assertLen(t, "gauges", len(gauges), 6)

	// Verify gauge values
	for _, g := range gauges {
		switch g.Name {
		case "filewatcher_watch_count":
			assertEqual(t, "watch_count", g.Value, 42.0)
		case "filewatcher_is_watching":
			assertEqual(t, "is_watching", g.Value, 1.0)
		case "filewatcher_is_closed":
			assertEqual(t, "is_closed", g.Value, 0.0)
		case "filewatcher_uptime_seconds":
			assertEqual(t, "uptime_seconds", g.Value, 5.0)
		}
	}

	// StatsFn should have been called at least twice (once for counters, once for gauges)
	if callCount.Load() < 2 {
		t.Errorf("statsFn called %d times, want >= 2", callCount.Load())
	}
}

func TestPrometheusCollector_NilStatsFunc(t *testing.T) {
	t.Parallel()

	// nil stats function should not panic
	collector := NewPrometheusCollector(nil)

	counters := collector.Counters()
	assertLen(t, "counters (nil stats)", len(counters), 4)

	for _, c := range counters {
		if c.Value != 0 {
			t.Errorf("counter %s = %d, want 0 with nil stats", c.Name, c.Value)
		}
	}
}

func TestPrometheusCollector_BoolConversion(t *testing.T) {
	t.Parallel()

	statsFn := func() Stats {
		return Stats{
			IsWatching: true,
			IsClosed:   true,
		}
	}

	collector := NewPrometheusCollector(statsFn)
	gauges := collector.Gauges()

	for _, g := range gauges {
		if (g.Name == "filewatcher_is_watching" || g.Name == "filewatcher_is_closed") && g.Value != 1 {
			t.Errorf("gauge %s = %v, want 1 for true", g.Name, g.Value)
		}
	}
}
