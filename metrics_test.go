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
	if len(counters) != 4 {
		t.Errorf("expected 4 counters, got %d", len(counters))
	}

	// Verify a specific counter value
	for _, counter := range counters {
		if counter.Name == "filewatcher_events_processed_total" && counter.Value != 100 {
			t.Errorf("events_processed_total = %d, want 100", counter.Value)
		}

		if counter.Name == "filewatcher_errors_encountered_total" && counter.Value != 2 {
			t.Errorf("errors_encountered_total = %d, want 2", counter.Value)
		}
	}

	gauges := collector.Gauges()
	if len(gauges) != 6 {
		t.Errorf("expected 6 gauges, got %d", len(gauges))
	}

	// Verify gauge values
	for _, g := range gauges {
		switch g.Name {
		case "filewatcher_watch_count":
			if g.Value != 42 {
				t.Errorf("watch_count = %v, want 42", g.Value)
			}
		case "filewatcher_is_watching":
			if g.Value != 1 {
				t.Errorf("is_watching = %v, want 1", g.Value)
			}
		case "filewatcher_is_closed":
			if g.Value != 0 {
				t.Errorf("is_closed = %v, want 0", g.Value)
			}
		case "filewatcher_uptime_seconds":
			if g.Value != 5 {
				t.Errorf("uptime_seconds = %v, want 5", g.Value)
			}
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
	if len(counters) != 4 {
		t.Errorf("expected 4 counters even with nil stats, got %d", len(counters))
	}

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
