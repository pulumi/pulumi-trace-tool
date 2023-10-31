package traces

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"sort"
	"strings"
	"time"

	"github.com/pulumi/pulumi-trace-tool/intervals"
)

type MetricsSink struct {
	writeMetrics func(data []map[string]string) error
}

func NewCsvMetricsSink(writer io.Writer) MetricsSink {
	return MetricsSink{
		func(data []map[string]string) error {
			return writeMetricsToCsvWriter(data, writer)
		},
	}
}

func NewParquetFileMetricsSink(filePath string) MetricsSink {
	return MetricsSink{
		func(data []map[string]string) error {
			records, err := parseParquetRecords(data)
			if err != nil {
				return err
			}
			return writeParquetRecords(filePath, records)
		},
	}
}

func Metrics(csvFile string, filenameColumn string, sink MetricsSink) error {
	aliases := metricAliases()

	invAliases := make(map[string]string)
	for k, vs := range aliases {
		for _, v := range vs {
			invAliases[v] = k
		}
	}

	files := make(map[string]string)

	// find set of source files in the data
	err := readLargeCsvFile(csvFile, func(row map[string]string) error {
		f := row[filenameColumn]
		if f != "" {
			files[f] = f
		}
		return nil
	})
	if err != nil {
		return err
	}

	var metrics []map[string]string

	for f := range files {

		var apiOverhead, engDuration time.Duration
		var engStart time.Time
		var haveEngStart bool
		var pulumiApiEndpoint string

		miscMetrics := map[string]*intervals.TimeTracker{}
		for _, metric := range metricsAccumulators() {
			miscMetrics[metric] = &intervals.TimeTracker{}
		}

		precomputeMetricsFromRow := func(row map[string]string) error {
			if row[filenameColumn] != f {
				return nil
			}

			for rowName, metric := range metricsAccumulators() {
				if row["Name"] == rowName {
					iv, err := spanInterval(row)
					if err != nil {
						return err
					}
					if err := miscMetrics[metric].Track(iv); err != nil {
						return err
					}
				}
			}

			if row["Name"] == "pulumi-plan" {
				t0, err := spanStart(row)
				if err != nil {
					return err
				}
				engStart = t0
				haveEngStart = true

				dur, err := spanDuration(row)
				if err != nil {
					return err
				}
				engDuration = dur
			}

			if row["api"] != "" {
				pulumiApiEndpoint = row["api"]
				dur, err := spanDuration(row)
				if err != nil {
					return err
				}
				apiOverhead += dur
			}

			return nil
		}

		precomputeTolerant := tolerateFaults(csvFile+"#precompute", precomputeMetricsFromRow)
		if err := readLargeCsvFile(csvFile, precomputeTolerant); err != nil {
			return err
		}

		emitMetricsFromRow := func(row map[string]string) error {
			if row[filenameColumn] != f {
				return nil
			}

			// Detect the all-encompassing span collected from the
			// top-level `pulumi` invocation.

			if row["Name"] == "pulumi" {
				m := make(map[string]string)

				t0, err := spanStart(row)
				if err != nil {
					return err
				}

				m[benchmark_start] = row["Span.Start"]

				// this is coming from `pulumi` CLI process, not a plugin
				m[pulumi_process] = "pulumi"

				// copy labels if found in aliases
				for k, v := range row {
					col, includeCol := invAliases[k]
					if includeCol {
						m[col] = v
					}
				}

				// infer benchmark phase; example inputs:
				//
				// filename=aws-go-s3-folder-pulumi-update-initial.trace
				// benchmark_name=aws-go-s3-folder
				m[benchmark_phase] = ""

				f := path.Base(row[filenameColumn])
				if strings.HasPrefix(f, m[benchmark_name]+"-") {
					s := strings.TrimPrefix(f, m[benchmark_name]+"-")
					if strings.HasSuffix(s, ".trace") {
						s = strings.TrimSuffix(s, ".trace")
						m[benchmark_phase] = s
					}
				}

				// use pre-computed things here
				m[time_engine_ms] = ms(engDuration)
				m[pulumi_api] = pulumiApiEndpoint
				m[time_pulumi_api_ms] = ms(apiOverhead)

				for k, v := range miscMetrics {
					m[k] = ms(v.TimeTaken())
				}

				if haveEngStart {
					m[time_to_engine_ms] = ms(engStart.Sub(t0))
				} else {
					m[time_to_engine_ms] = ""
				}

				metrics = append(metrics, m)
			}

			return nil
		}

		emitTolerant := tolerateFaults(csvFile, emitMetricsFromRow)
		if err := readLargeCsvFile(csvFile, emitTolerant); err != nil {
			return err
		}
	}

	if err := sink.writeMetrics(metrics); err != nil {
		return err
	}

	return nil
}

// Map span name to the duration sum counter.
func metricsAccumulators() map[string]string {
	return map[string]string{
		"pulumi":                         time_total_ms,
		"/pulumirpc.Engine/Log":          time_log_overhead_ms,
		"api/patchCheckpoint":            time_patch_checkpoint_ms,
		"/pulumirpc.LanguageRuntime/Run": time_language_runtime_run_ms,

		"/pulumirpc.LanguageRuntime/GetRequiredPlugins": time_get_required_plugins_ms,
		"/pulumirpc.ResourceMonitor/RegisterResource":   time_register_resource_ms,
		"/pulumirpc.ResourceProvider/Configure":         time_resource_provider_configure_ms,
		"/pulumirpc.ResourceProvider/Create":            time_resource_provider_create_ms,
	}
}

func spanStart(row map[string]string) (time.Time, error) {
	spanStart, err := parseTime(row["Span.Start"])
	if err != nil {
		err = fmt.Errorf("Failed to parse Span.Start time: %w", err)
		return time.Time{}, err
	}
	return spanStart, nil
}

func spanEnd(row map[string]string) (time.Time, error) {
	spanEnd, err := parseTime(row["Span.End"])
	if err != nil {
		err = fmt.Errorf("Failed to parse Span.End time: %w", err)
		return time.Time{}, err
	}
	return spanEnd, nil
}

func spanInterval(row map[string]string) (intervals.Interval, error) {
	spanStart, err := spanStart(row)
	if err != nil {
		return intervals.Interval{}, err
	}
	spanEnd, err := spanEnd(row)
	if err != nil {
		return intervals.Interval{}, err
	}
	return intervals.Interval{Start: spanStart, End: spanEnd}, nil
}

func spanDuration(row map[string]string) (time.Duration, error) {
	spanStart, err := spanStart(row)
	if err != nil {
		return 0, err
	}
	spanEnd, err := spanEnd(row)
	if err != nil {
		return 0, err
	}
	dur := spanEnd.Sub(spanStart)
	return dur, nil
}

func readLargeCsvFile(csvFile string, handleRow func(map[string]string) error) error {
	f, err := os.Open(csvFile)
	if err != nil {
		return err
	}
	defer f.Close()

	reader := bufio.NewReader(f)
	csvReader := csv.NewReader(reader)

	header, err := csvReader.Read()
	if err != nil {
		return err
	}

	for {
		record, err := csvReader.Read()

		if err == io.EOF {
			break
		}

		if err != nil {
			return err
		}

		values := make(map[string]string, len(header))

		for i, value := range record {
			values[header[i]] = value
		}

		if err := handleRow(values); err != nil {
			return err
		}
	}

	return nil
}

// Formats a CSV data row as an easy-to-read string to emit in logs.
// Empty fields are omitted, fields are aligned and sorted by name.
//
// Example output:
//
//	        Name: api/startUpdate
//	         api: https://api.pulumi.com
//	    filename: azure-classic-csharp-pulumi-destroy.trace
//	      method: POST
//	        path: /api/stacks/me/test-env2047447553/p-it-..
//	responseCode: 200 OK
//	       retry: false
func prettyPrintRow(indent string, row map[string]string) string {
	var keys []string
	var maxLenKey int
	for k := range row {
		if len(k) > maxLenKey {
			maxLenKey = len(k)
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var buf bytes.Buffer
	for _, k := range keys {
		v := row[k]
		if v != "" {
			fmt.Fprintf(&buf, "%s%*s: %s\n", indent, maxLenKey, k, v)
		}
	}
	return buf.String()
}

// Allows to read CSV files ignoring individual row parse failures,
// and logging them via log.Printf instead of returning an error.
//
// Intended use is to transform this code:
//
//	readLargeCsvFile(csvFile, parseRow)
//
// Into this code:
//
//	readLargeCsvFile(csvFile, tolerateFaults(csvFile, parseRow))
//
// The tolerateFaults code will process every row and return nil error
// and log failed rows, instead of stopping at the first failed row
// and returning an error.
func tolerateFaults(
	csvFile string,
	handleRow func(map[string]string) error,
) func(map[string]string) error {
	return func(row map[string]string) error {
		if err := handleRow(row); err != nil {
			log.Printf("WARN ignoring failure to parse a row from %s\n  Error: %v\n  Data:\n%s",
				csvFile,
				err,
				prettyPrintRow("    ", row))
		}
		return nil
	}
}

func writeMetricsToCsvWriter(data []map[string]string, writer io.Writer) error {
	csvWriter := csv.NewWriter(writer)
	defer csvWriter.Flush()

	columns := make(map[string]int)

	addColumn := func(name string) {
		_, seen := columns[name]
		if !seen {
			n := len(columns)
			columns[name] = n
		}
	}

	for _, row := range data {
		for k := range row {
			addColumn(k)
		}
	}

	columnNames := make([]string, len(columns))

	for k, j := range columns {
		columnNames[j] = k
	}

	if err := csvWriter.Write(columnNames); err != nil {
		return err
	}

	for _, row := range data {
		values := make([]string, len(columns))

		for k, v := range row {
			values[columns[k]] = v
		}

		if err := csvWriter.Write(values); err != nil {
			return err
		}
	}

	return nil
}

func parseTime(str string) (time.Time, error) {
	return time.Parse(time.RFC3339, str)
}

func ms(dur time.Duration) string {
	return fmt.Sprintf("%v", int64(dur/time.Millisecond))
}
