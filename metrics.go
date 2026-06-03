package filewatcher

// PrometheusCollector is a Prometheus metrics collector for the watcher.
// It implements the prometheus.Collector interface:
//
//	type Collector interface {
//	    Describe(chan<- *Desc)
//	    Collect(chan<- Metric)
//	}
//
// Users register it with their prometheus.Registry:
//
//	import "github.com/prometheus/client_golang/prometheus"
//
//	collector := filewatcher.NewPrometheusCollector(watcher.Stats)
//	prometheus.MustRegister(collector)
//
// The collector does not depend on prometheus/client_golang — it is a
// standalone type that can be adapted by users to any metrics system
// (Prometheus, OpenTelemetry, Datadog, etc.) without adding a dependency
// to this library.
//
// The StatsFunc is called on each Collect invocation to retrieve the
// current Stats snapshot. This allows the collector to be reused across
// watchers by passing different StatsFunc closures.
type PrometheusCollector struct {
	stats    StatsFunc
	describe []string // metric help text
}

// StatsFunc is a function that returns a Stats snapshot. It is called
// on each Prometheus scrape to get the latest values.
type StatsFunc func() Stats

// NewPrometheusCollector returns a Prometheus collector for a watcher's
// stats. The provided function should return the current Stats (e.g.,
// `watcher.Stats` or a closure that calls it).
func NewPrometheusCollector(stats StatsFunc) *PrometheusCollector {
	if stats == nil {
		stats = func() Stats { return Stats{} }
	}

	return &PrometheusCollector{
		stats: stats,
		describe: []string{
			"Total events that passed all filters",
			"Events filtered out (dropped by filters)",
			"Errors encountered during processing",
			"Watch add failures (ENOSPC, permission denied, etc.)",
		},
	}
}

// CounterMetric is a generic counter descriptor for a Prometheus metric.
// It carries the name, help text, and value extraction from Stats.
//
// Users adapt this to their metrics library by iterating over the
// PrometheusCollector.Counters() method, which returns []CounterMetric.
type CounterMetric struct {
	Name  string
	Help  string
	Value uint64
}

// GaugeMetric is a generic gauge descriptor for a Prometheus metric.
type GaugeMetric struct {
	Name  string
	Help  string
	Value float64
}

// Counters returns the current counter values from the watcher stats.
// Callers can map these to their metrics library of choice.
func (c *PrometheusCollector) Counters() []CounterMetric {
	stats := c.stats()

	return []CounterMetric{
		{
			Name:  "filewatcher_events_processed_total",
			Help:  c.describe[0],
			Value: stats.EventsProcessed,
		},
		{
			Name:  "filewatcher_events_filtered_out_total",
			Help:  c.describe[1],
			Value: stats.EventsFilteredOut,
		},
		{
			Name:  "filewatcher_errors_encountered_total",
			Help:  c.describe[2],
			Value: stats.ErrorsEncountered,
		},
		{
			Name:  "filewatcher_watch_errors_total",
			Help:  c.describe[3],
			Value: stats.WatchErrors,
		},
	}
}

// Gauges returns the current gauge values from the watcher stats.
func (c *PrometheusCollector) Gauges() []GaugeMetric {
	stats := c.stats()

	return []GaugeMetric{
		{
			Name:  "filewatcher_watch_count",
			Help:  "Number of paths currently being watched",
			Value: float64(stats.WatchCount),
		},
		{
			Name:  "filewatcher_is_watching",
			Help:  "1 if the watcher is currently running, 0 otherwise",
			Value: boolToFloat(stats.IsWatching),
		},
		{
			Name:  "filewatcher_is_closed",
			Help:  "1 if the watcher has been closed, 0 otherwise",
			Value: boolToFloat(stats.IsClosed),
		},
		{
			Name:  "filewatcher_uptime_seconds",
			Help:  "Time in seconds since the watcher was started",
			Value: stats.Uptime.Seconds(),
		},
		{
			Name:  "filewatcher_watch_limit",
			Help:  "System inotify watch limit (0 if unknown)",
			Value: float64(stats.WatchLimit),
		},
		{
			Name:  "filewatcher_watch_budget_used_ratio",
			Help:  "Percentage of inotify budget used (0.0-1.0)",
			Value: stats.WatchBudgetUsed,
		},
	}
}

func boolToFloat(b bool) float64 {
	if b {
		return 1
	}

	return 0
}
