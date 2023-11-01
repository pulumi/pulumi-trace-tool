// Helpers in this file are used in `pulumi/examples` and
// `pulumi/templates` to tag and extract trace files out of benchmark
// programs.

package traces

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pulumi/pulumi/pkg/v3/testing/integration"
	"github.com/pulumi/pulumi/sdk/v3/go/common/util/contract"
)

// Env variable specifying the folder where trace output should go. If
// this variable is omitted, functions in this module treat tracing as
// disabled, return empty flags or env vars.
const TRACING_DIR_ENV_VAR = "PULUMI_TRACING_DIR"

// Retrieves `TRACING_DIR_ENV_VAR` to find where trace files should go.
func TracingDir() string {
	return os.Getenv(TRACING_DIR_ENV_VAR)
}

// Check if tracing is enabled via the `TRACING_DIR_ENV_VAR`.
func IsTracingEnabled() bool {
	return TracingDir() != ""
}

// Metadata about a benchmark used for tracking performance with `pulumi --tracing`.
type Benchmark struct {
	// Name of the benchmark (normally use the folder name)
	Name string

	// Primary provider (such as aws) the benchmark is testing
	Provider string

	// Primary runtime set in Pulumi.yaml
	Runtime string

	// Main programming language in the benchmark
	Language string

	// Repository where the benchmark is found, such as `pulumi/templates`
	Repository string

	// How often to sample memory stats
	MemstatsPollInterval time.Duration
}

// Creates a minimal benchmark configuration with default parameters.
func NewBenchmark(name string) Benchmark {
	return Benchmark{
		Name:                 name,
		MemstatsPollInterval: 100 * time.Millisecond,
	}
}

// Computes a list of `K=V` environment variables that will inform
// `pulumi --tracing` how to tag the data it produces.
func (benchmark *Benchmark) Env() []string {
	if !IsTracingEnabled() {
		return []string{}
	}

	env := []string{}

	if benchmark.Name != "" {
		env = append(env, fmt.Sprintf("PULUMI_TRACING_TAG_BENCHMARK_NAME=%s", benchmark.Name))
	}

	if benchmark.Repository != "" {
		env = append(env, fmt.Sprintf("PULUMI_TRACING_TAG_REPO=%s", benchmark.Repository))
	}

	if benchmark.Provider != "" {
		env = append(env, fmt.Sprintf("PULUMI_TRACING_TAG_BENCHMARK_PROVIDER=%s", benchmark.Provider))
	}

	if benchmark.Runtime != "" {
		env = append(env, fmt.Sprintf("PULUMI_TRACING_TAG_BENCHMARK_RUNTIME=%s", benchmark.Runtime))
	}

	if benchmark.Language != "" {
		env = append(env, fmt.Sprintf("PULUMI_TRACING_TAG_BENCHMARK_LANGUAGE=%s", benchmark.Language))
	}

	if benchmark.MemstatsPollInterval > 0 {
		env = append(env, fmt.Sprintf("PULUMI_TRACING_MEMSTATS_POLL_INTERVAL=%v", benchmark.MemstatsPollInterval))
	}

	return env
}

// Ensures `ProgramTest` uses appropriate `--tracing` options.
func (benchmark *Benchmark) ProgramTestOptions() integration.ProgramTestOptions {
	if !IsTracingEnabled() {
		return integration.ProgramTestOptions{}
	}

	dir := TracingDir()

	return integration.ProgramTestOptions{
		Env: benchmark.Env(),
		Tracing: fmt.Sprintf("file:%s",
			filepath.Join(dir, fmt.Sprintf("%s-{command}.trace", benchmark.Name))),
	}
}

// Computes `--tracing` option to pass to `pulumi` CLI. The
// `commandName` flag is used to differentiate steps such as
// `pulumi-preview`.
func (benchmark *Benchmark) CommandArgs(commandName string) []string {
	opts := benchmark.ProgramTestOptions()
	if opts.Tracing != "" {
		return []string{"--tracing", strings.ReplaceAll(opts.Tracing, "{command}", commandName)}
	}
	return []string{}
}

// If tracing is enabled, finds all *.trace files in `TracingDir` and
// computes metrics, producing `TracingDir/metrics.parquet.snappy`.
func ComputeMetrics() error {
	if !IsTracingEnabled() {
		return nil
	}

	wrap := func(err error) error {
		return fmt.Errorf("ComputeMetrics() error: %w", err)
	}

	dir := TracingDir()

	cwd, err := os.Getwd()
	if err != nil {
		return wrap(err)
	}

	defer contract.IgnoreError(os.Chdir(cwd))

	err = os.Chdir(dir)
	if err != nil {
		return wrap(err)
	}

	files, err := os.ReadDir(".")
	if err != nil {
		return wrap(err)
	}

	var traceFiles []string
	for _, f := range files {
		if strings.HasSuffix(f.Name(), ".trace") {
			traceFiles = append(traceFiles, f.Name())
		}
	}

	if err := ToCsv(traceFiles, "traces.csv", "filename"); err != nil {
		return wrap(err)
	}

	if err := Metrics("traces.csv", "filename", NewParquetFileMetricsSink("metrics.parquet.snappy")); err != nil {
		return wrap(err)
	}

	return nil
}
