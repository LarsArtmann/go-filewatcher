import type { Feature } from "./types";

export const features: Feature[] = [
  {
    icon: "bolt",
    title: "Zero Boilerplate",
    desc: "Start watching in 5 lines of code. Sensible defaults, automatic recursion, context-aware cancellation.",
  },
  {
    icon: "filter",
    title: "17+ Built-in Filters",
    desc: "Extensions, globs, regex, size, age, content hash, gitignore, generated-code detection. Compose with AND/OR/NOT.",
  },
  {
    icon: "layers",
    title: "18 Middleware",
    desc: "Logging, recovery, rate limiting, circuit breaker, metrics, batching, error correlation, exponential backoff.",
  },
  {
    icon: "refresh",
    title: "Resilient by Default",
    desc: "Self-healing watches, inotify budget awareness, graceful ENOSPC handling, and .gitignore-aware walking.",
  },
  {
    icon: "chart",
    title: "Full Observability",
    desc: "Built-in Stats(), Prometheus collector, OpenTelemetry tracing middleware, and structured debug logging.",
  },
  {
    icon: "server",
    title: "NFS/FUSE Friendly",
    desc: "Optional polling mode supplements OS-native events for network filesystems, Docker volumes, and FUSE mounts.",
  },
];
