# CLI Performance Metrics

`pulumi-trace-tool` turns Pulumi CLI traces into per-span metrics, so CLI performance can be
measured over time (for example, from CI benchmarks).

## How it works

It builds on the CLI's built-in `--tracing` flag:

```
pulumi <cmd> --tracing file:./cmd.trace
pulumi-trace-tool tocsv -csv cmd.csv cmd.trace
pulumi-trace-tool metrics -csv cmd.csv > metrics.csv
```

First run a command, recording its trace. `tocsv` is a generic converter that turns the trace
into a CSV with one row per span (span annotations become columns); it knows nothing about
Pulumi specifics. `metrics` then applies Pulumi knowledge — locating the top-level span and the
engine span — to interpret the data and compute the metrics below.

At the moment only numbers from the top-level `pulumi` command are included. Building on the
tracing code makes it possible to include numbers from sub-processes (language and resource
plugins, multi-language providers) later.

## Metrics

See [`traces/metric_names.go`](https://github.com/pulumi/pulumi-trace-tool/blob/main/traces/metric_names.go)
for the full list. Some notable ones:

- **`benchmark_phase`** — the phase of a `ProgramTest` run. A single `ProgramTest` runs multiple
  `pulumi` invocations, distinguishing `pulumi-preview-initial`, `pulumi-preview-empty`, and other
  phases.
- **`time_to_engine_ms`** — the interval between the benchmark starting and the span covering
  engine work starting; indicative of init and service-communication overhead.
- **`time_engine_ms`** — the duration of the span covering the engine work.
- **`time_pulumi_api_ms`** — total time spent in requests to the Pulumi Service backend
  (overlapping requests are counted once, not summed).
- **`time_total_ms`** — total command wall time.

## Where it's used

Pulumi runs CLI benchmarks in CI and feeds their traces through this tool. The benchmarks live in:

- [`pulumi/examples` — `performance_test.go`](https://github.com/pulumi/examples/blob/master/misc/test/performance_test.go)
- [`pulumi/templates` — `template_test.go`](https://github.com/pulumi/templates/blob/master/tests/template_test.go)

The helper code lives here in `pulumi-trace-tool`, but could be upstreamed into `pulumi/pulumi`
eventually.
