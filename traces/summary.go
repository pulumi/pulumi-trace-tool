package traces

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
)

func Summary(traceFiles []string, filenameColumn string) error {
	tempCsv, err := os.CreateTemp("", "pulumi-decoded-traces")
	if err != nil {
		return err
	}
	defer noErr(os.Remove(tempCsv.Name()))

	tempCsv2, err := os.CreateTemp("", "pulumi-metrics")
	if err != nil {
		return err
	}
	defer noErr(os.Remove(tempCsv2.Name()))

	if err := ToCsv(traceFiles, tempCsv.Name(), filenameColumn); err != nil {
		return fmt.Errorf("Failed converting trace files to CSV: %w", err)
	}

	w, err := os.Create(tempCsv2.Name())
	if err != nil {
		return err
	}
	if err := Metrics(tempCsv.Name(), filenameColumn, NewCsvMetricsSink(w)); err != nil {
		return fmt.Errorf("Failed to compute metrics: %w", err)
	}
	if err := w.Close(); err != nil {
		return err
	}

	f, err := os.Open(tempCsv2.Name())
	if err != nil {
		return err
	}
	defer ensureClosed(f)

	csvReader := csv.NewReader(f)
	csvWriter := csv.NewWriter(os.Stdout)

	return csvSelectColumns([]string{
		benchmark_name,
		benchmark_phase,
		time_total_ms,
		time_pulumi_api_ms,
		time_to_engine_ms,
		time_language_runtime_run_ms,
		time_patch_checkpoint_ms,
		time_get_required_plugins_ms,
		time_register_resource_ms,
		time_resource_provider_configure_ms,
	}, csvReader, csvWriter)
}

func noErr(err error) {
	if err != nil {
		panic(err)
	}
}

func ensureClosed(x io.Closer) {
	noErr(x.Close())
}

func csvSelectColumns(columns []string, reader *csv.Reader, writer *csv.Writer) error {
	header, err := reader.Read()
	if err == io.EOF {
		return nil
	}
	if err != nil {
		return err
	}

	colIndex := func(col string) int {
		for i, h := range header {
			if h == col {
				return i
			}
		}
		return -1
	}

	indices := []int{}
	for _, c := range columns {
		i := colIndex(c)
		if i == -1 {
			return fmt.Errorf("Unknown column: %s", c)
		}
		indices = append(indices, i)
	}

	selected := func(row []string) []string {
		result := []string{}
		for _, i := range indices {
			result = append(result, row[i])
		}
		return result
	}

	if err := writer.Write(selected(header)); err != nil {
		return err
	}

	for {
		record, err := reader.Read()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		if err := writer.Write(selected(record)); err != nil {
			return err
		}

		writer.Flush()
	}
}
