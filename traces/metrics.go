package traces

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
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

		var logOverhead, apiOverhead, engDuration time.Duration
		var engStart time.Time
		var haveEngStart bool
		var pulumiApiEndpoint string

		// precompute metrics
		err = readLargeCsvFile(csvFile, func(row map[string]string) error {

			if row[filenameColumn] != f {
				return nil
			}

			if row["Name"] == "/pulumirpc.Engine/Log" {
				dur, err := spanDuration(row)
				if err != nil {
					return err
				}
				logOverhead += dur
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
		})
		if err != nil {
			return err
		}

		// emit metrics
		err = readLargeCsvFile(csvFile, func(row map[string]string) error {

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

				// compute duration
				dur, err := spanDuration(row)
				if err != nil {
					return err
				}
				m[time_total_ms] = ms(dur)

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
				if strings.HasPrefix(row[filenameColumn], m[benchmark_name]+"-") {
					s := strings.TrimPrefix(row[filenameColumn], m[benchmark_name]+"-")
					if strings.HasSuffix(s, ".trace") {
						s = strings.TrimSuffix(s, ".trace")
						m[benchmark_phase] = s
					}
				}

				// use pre-computed things here
				m[time_engine_ms] = ms(engDuration)
				m[pulumi_api] = pulumiApiEndpoint
				m[time_pulumi_api_ms] = ms(apiOverhead)
				m[time_log_overhead_ms] = ms(logOverhead)

				if haveEngStart {
					m[time_to_engine_ms] = ms(engStart.Sub(t0))
				} else {
					m[time_to_engine_ms] = ""
				}

				metrics = append(metrics, m)
			}

			return nil
		})

		if err != nil {
			return err
		}
	}

	if err := sink.writeMetrics(metrics); err != nil {
		return err
	}

	return nil
}

func spanStart(row map[string]string) (time.Time, error) {
	spanStart, err := parseTime(row["Span.Start"])
	if err != nil {
		return time.Time{}, err
	}
	return spanStart, nil
}

func spanDuration(row map[string]string) (time.Duration, error) {
	spanStart, err := spanStart(row)
	if err != nil {
		return 0, err
	}
	spanEnd, err := parseTime(row["Span.End"])
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
