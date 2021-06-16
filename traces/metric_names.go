package traces

const benchmark_name string = "benchmark_name"
const time_total_ms string = "time_total_ms"
const pulumi_api = "pulumi_api"

const time_log_overhead_ms = "time_log_overhead_ms"
const time_pulumi_api_ms = "time_pulumi_api_ms"
const time_to_engine_ms = "time_to_engine_ms"

// Duration of the span marked `pulumi-plan`. This span seem to cover
// plan and/or update operations.
const time_engine_ms string = "time_engine_ms"

// Phases include things like pulumi-update-initial as defined by ProgramTest.
const benchmark_phase string = "benchmark_phase"

// Process such as `pulumi` or `pulumi-resource-aws` the data is coming from.
const pulumi_process string = "pulumi_process"

// Maps canonical column names as they should appear in `merics.csv`
// to a list of possible aliases how they appear in the `traces.csv`
// files and binary traces.
func metricAliases() map[string][]string {
	return map[string][]string{
		benchmark_name:          {"benchmark_name"},
		"benchmark_provider":    {"benchmark_provider", "benchmark_cloud"},
		"benchmark_repo":        {"repo"},
		"benchmark_runtime":     {"benchmark_runtime"},
		"benchmark_language":    {"benchmark_language"},
		"mem_frees":             {"MemStats.Frees"},
		"mem_heap_alloc_max":    {"MemStats.HeapAlloc.Max"},
		"mem_heap_idle_max":     {"MemStats.HeapIdle.Max"},
		"mem_heap_inuse_max":    {"MemStats.HeapInuse.Max"},
		"mem_heap_objects_max":  {"MemStats.HeapObjects.Max"},
		"mem_heap_released_max": {"MemStats.HeapReleased.Max"},
		"mem_heap_sys_max":      {"MemStats.HeapSys.Max"},
		"mem_mallocs":           {"MemStats.Mallocs"},
		"mem_num_gc":            {"MemStats.NumGC"},
		"mem_pause_total_ns":    {"MemStats.PauseTotalNs"},
		"mem_stack_in_use_max":  {"MemStats.StackInuse.Max"},
		"mem_stack_sys_max":     {"MemStats.StackSys.Max"},
		"mem_sys_max":           {"MemStats.Sys.Max"},
		"mem_total_alloc":       {"MemStats.TotalAlloc"},
		"pulumi_version":        {"pulumi_version"},
		"pulumi_commandline":    {"os.Args"},
		"runner_arch":           {"runtime.GOARCH"},
		"runner_num_cpu":        {"runtime.NumCPU"},
		"runner_os":             {"runtime.GOOS"},
	}
}
