// Transcodes metrics into Parquet so that we can easily query them as
// an external table in Redshift Spectrum.

package traces

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/xitongsys/parquet-go-source/local"
	"github.com/xitongsys/parquet-go/parquet"
	"github.com/xitongsys/parquet-go/writer"
)

// TestCommandStats is a collection of data related to running a single command during a test.
type ParquetRecord struct {
	Benchmark_language    *string `parquet:"name=benchmark_language, type=BYTE_ARRAY, convertedtype=UTF8"`
	Benchmark_name        *string `parquet:"name=benchmark_name, type=BYTE_ARRAY, convertedtype=UTF8"`
	Benchmark_phase       *string `parquet:"name=benchmark_phase, type=BYTE_ARRAY, convertedtype=UTF8"`
	Benchmark_provider    *string `parquet:"name=benchmark_provider, type=BYTE_ARRAY, convertedtype=UTF8"`
	Benchmark_repo        *string `parquet:"name=benchmark_repo, type=BYTE_ARRAY, convertedtype=UTF8"`
	Benchmark_start       *string `parquet:"name=benchmark_start, type=BYTE_ARRAY, convertedtype=UTF8"`
	Mem_frees             *int64  `parquet:"name=mem_frees, type=INT64"`
	Mem_heap_alloc_max    *int64  `parquet:"name=mem_heap_alloc_max, type=INT64"`
	Mem_heap_idle_max     *int64  `parquet:"name=mem_heap_idle_max, type=INT64"`
	Mem_heap_inuse_max    *int64  `parquet:"name=mem_heap_inuse_max, type=INT64"`
	Mem_heap_objects_max  *int64  `parquet:"name=mem_heap_objects_max, type=INT64"`
	Mem_heap_released_max *int64  `parquet:"name=mem_heap_released_max, type=INT64"`
	Mem_heap_sys_max      *int64  `parquet:"name=mem_heap_sys_max, type=INT64"`
	Mem_mallocs           *int64  `parquet:"name=mem_mallocs, type=INT64"`
	Mem_num_gc            *int64  `parquet:"name=mem_num_gc, type=INT64"`
	Mem_pause_total_ns    *int64  `parquet:"name=mem_pause_total_ns, type=INT64"`
	Mem_stack_in_use_max  *int64  `parquet:"name=mem_stack_in_use_max, type=INT64"`
	Mem_stack_sys_max     *int64  `parquet:"name=mem_stack_sys_max, type=INT64"`
	Mem_sys_max           *int64  `parquet:"name=mem_sys_max, type=INT64"`
	Mem_total_alloc       *int64  `parquet:"name=mem_total_alloc, type=INT64"`
	Pulumi_api            *string `parquet:"name=pulumi_api, type=BYTE_ARRAY, convertedtype=UTF8"`
	Pulumi_commandline    *string `parquet:"name=pulumi_commandline, type=BYTE_ARRAY, convertedtype=UTF8"`
	Pulumi_process        *string `parquet:"name=pulumi_process, type=BYTE_ARRAY, convertedtype=UTF8"`
	Pulumi_version        *string `parquet:"name=pulumi_version, type=BYTE_ARRAY, convertedtype=UTF8"`
	Runner_arch           *string `parquet:"name=runner_arch, type=BYTE_ARRAY, convertedtype=UTF8"`
	Runner_num_cpu        *int64  `parquet:"name=runner_num_cpu, type=INT64"`
	Runner_os             *string `parquet:"name=runner_os, type=BYTE_ARRAY, convertedtype=UTF8"`
	Time_engine_ms        *int64  `parquet:"name=time_engine_ms, type=INT64"`
	Time_log_overhead_ms  *int64  `parquet:"name=time_log_overhead_ms, type=INT64"`
	Time_pulumi_api_ms    *int64  `parquet:"name=time_pulumi_api_ms, type=INT64"`
	Time_to_engine_ms     *int64  `parquet:"name=time_to_engine_ms, type=INT64"`
	Time_total_ms         *int64  `parquet:"name=time_total_ms, type=INT64"`
}

func parseParquetRecord(dataRow map[string]string) (ParquetRecord, error) {
	r := ParquetRecord{}
	var err error

	newStrPtr := func(value string) *string {
		return &value
	}

	newI64Ptr := func(value int64) *int64 {
		return &value
	}

	str := func(out **string, key string) {
		strVal, hasVal := dataRow[key]
		if hasVal {
			*out = newStrPtr(strVal)
		}
	}

	i64 := func(out **int64, key string) {
		strVal, hasVal := dataRow[key]
		if hasVal && strVal != "" {
			if n, e := strconv.ParseInt(strVal, 10, 64); e == nil {
				*out = newI64Ptr(n)
			} else {
				err = fmt.Errorf("Failed to parse integer column %s value %s as an int64: %w",
					key, strVal, e)
			}
		}
	}

	str(&r.Benchmark_language, "benchmark_language")
	str(&r.Benchmark_name, "benchmark_name")
	str(&r.Benchmark_phase, "benchmark_phase")
	str(&r.Benchmark_provider, "benchmark_provider")
	str(&r.Benchmark_repo, "benchmark_repo")
	str(&r.Benchmark_start, "benchmark_start")
	str(&r.Pulumi_api, "pulumi_api")
	str(&r.Pulumi_commandline, "pulumi_commandline")
	str(&r.Pulumi_process, "pulumi_process")
	str(&r.Pulumi_version, "pulumi_version")
	str(&r.Runner_arch, "runner_arch")
	str(&r.Runner_os, "runner_os")

	i64(&r.Mem_frees, "mem_frees")
	i64(&r.Mem_heap_alloc_max, "mem_heap_alloc_max")
	i64(&r.Mem_heap_idle_max, "mem_heap_idle_max")
	i64(&r.Mem_heap_inuse_max, "mem_heap_inuse_max")
	i64(&r.Mem_heap_objects_max, "mem_heap_objects_max")
	i64(&r.Mem_heap_released_max, "mem_heap_released_max")
	i64(&r.Mem_heap_sys_max, "mem_heap_sys_max")
	i64(&r.Mem_mallocs, "mem_mallocs")
	i64(&r.Mem_num_gc, "mem_num_gc")
	i64(&r.Mem_pause_total_ns, "mem_pause_total_ns")
	i64(&r.Mem_stack_in_use_max, "mem_stack_in_use_max")
	i64(&r.Mem_stack_sys_max, "mem_stack_sys_max")
	i64(&r.Mem_sys_max, "mem_sys_max")
	i64(&r.Mem_total_alloc, "mem_total_alloc")
	i64(&r.Runner_num_cpu, "runner_num_cpu")
	i64(&r.Time_engine_ms, "time_engine_ms")
	i64(&r.Time_log_overhead_ms, "time_log_overhead_ms")
	i64(&r.Time_pulumi_api_ms, "time_pulumi_api_ms")
	i64(&r.Time_to_engine_ms, "time_to_engine_ms")
	i64(&r.Time_total_ms, "time_total_ms")

	return r, err
}

func parseParquetRecords(dataRows []map[string]string) ([]ParquetRecord, error) {
	var out []ParquetRecord
	for _, row := range dataRows {
		r, err := parseParquetRecord(row)
		if err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, nil
}

func writeParquetRecords(filePath string, records []ParquetRecord) error {
	if strings.HasSuffix(".parquet.snappy", filePath) {
		return fmt.Errorf("Parquet file path should have the .parquet.snappy extension: %s", filePath)
	}

	fw, err := local.NewLocalFileWriter(filePath)

	if err != nil {
		return err
	}
	defer fw.Close()

	pw, err := writer.NewParquetWriter(fw, new(ParquetRecord), 2)
	if err != nil {
		return err
	}

	pw.RowGroupSize = 128 * 1024 * 1024 //128M
	pw.CompressionType = parquet.CompressionCodec_SNAPPY

	for _, r := range records {
		err := pw.Write(r)
		if err != nil {
			return err
		}
	}

	if err = pw.WriteStop(); err != nil {
		return err
	}

	return nil
}
