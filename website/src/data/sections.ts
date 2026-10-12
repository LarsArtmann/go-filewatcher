import type { ComparisonItem, ComparisonMatrix, StepCard, UseCase } from "./types";

export const steps: StepCard[] = [
  {
    code: `w, _ := filewatcher.New(paths,
    filewatcher.WithExtensions(".go"),
)`,
    desc: "Initialize with paths and options. Functional options configure filters, middleware, debounce.",
    step: "1",
    stepColor: "accent",
    title: "Create",
  },
  {
    code: `events, _ := w.Watch(ctx)`,
    desc: "Call Watch(ctx) to start. Returns a read-only event channel. Goroutines handle the rest.",
    step: "2",
    stepColor: "accent",
    title: "Watch",
  },
  {
    code: `// filters + middleware applied automatically`,
    desc: "Events pass through your filter chain and middleware. 17+ filters, 18 middleware, composable.",
    step: "3",
    stepColor: "amber",
    title: "Pipeline",
  },
  {
    code: `for ev := range events {
    fmt.Println(ev.Op, ev.Path)
}`,
    desc: "Range over the channel. Each event carries path, op, timestamp, size, modtime, and optional hash.",
    step: "4",
    stepColor: "amber",
    title: "Consume",
  },
];

export const comparisons: ComparisonItem[] = [
  {
    accent: false,
    cons: [
      "No recursive watching",
      "Manual filter logic",
      "No debounce",
      "ENOSPC crashes the watcher",
      "No NFS/FUSE support",
    ],
    pros: ["Low-level control", "No abstraction overhead"],
    variant: "Raw fsnotify",
  },
  {
    accent: true,
    cons: [],
    pros: [
      "Automatic recursion + new directories",
      "17+ composable filters with AND/OR/NOT",
      "18 production-grade middleware",
      "Global + per-path debouncing",
      "Self-healing + inotify budget awareness",
      ".gitignore-aware walking",
      "Prometheus + OpenTelemetry built in",
    ],
    variant: "go-filewatcher",
  },
  {
    accent: false,
    cons: [
      "Limited or no middleware",
      "Few or no filters",
      "No observability hooks",
      "No resilience features",
    ],
    pros: ["Some convenience over raw fsnotify"],
    variant: "Other wrappers",
  },
];

export const comparisonMatrix: ComparisonMatrix = {
  columns: ["Raw fsnotify", "Other wrappers", "go-filewatcher"],
  rows: [
    { feature: "Recursive watching", values: ["no", "partial", "yes"] },
    { feature: "Built-in filters", values: ["no", "partial", "yes"] },
    { feature: "Middleware chains", values: ["no", "no", "yes"] },
    { feature: "Debouncing", values: ["no", "partial", "yes"] },
    { feature: ".gitignore-aware", values: ["no", "no", "yes"] },
    { feature: "ENOSPC resilience", values: ["no", "no", "yes"] },
    { feature: "NFS/FUSE polling", values: ["no", "no", "yes"] },
    { feature: "Prometheus + OTel", values: ["no", "no", "yes"] },
    { feature: "Self-healing", values: ["no", "no", "yes"] },
  ],
};

export const useCases: UseCase[] = [
  {
    desc: "Dev servers, build systems, and live-reload tools that trigger on file changes",
    icon: "bolt",
    title: "Hot Reload",
  },
  {
    desc: "Tail and process log files in real-time with debouncing and rate limiting",
    icon: "server",
    title: "Log Monitoring",
  },
  {
    desc: "Watch for changes and trigger pipelines, tests, or deployments automatically",
    icon: "refresh",
    title: "CI/CD Triggers",
  },
];
