package traces

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/tobgu/qframe"
)

func Summary(traceFiles []string, filenameColumn string) error {
	tempCsv, err := ioutil.TempFile("", "pulumi-decoded-traces")
	if err != nil {
		return err
	}

	defer func() {
		if err := os.Remove(tempCsv.Name()); err != nil {
			panic(err)
		}
	}()

	tempCsv2, err := ioutil.TempFile("", "pulumi-metrics")
	if err != nil {
		return err
	}

	defer func() {
		if err := os.Remove(tempCsv2.Name()); err != nil {
			panic(err)
		}
	}()

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

	fr := qframe.ReadCSV(f)

	defer func() {
		if err := f.Close(); err != nil {
			panic(err)
		}
	}()

	return fr.Select(
		"benchmark_name",
		"benchmark_phase",
		"time_total_ms",
		"time_pulumi_api_ms",
		"time_to_engine_ms",
		"time_language_runtime_run_ms",
		"time_patch_checkpoint_ms",
		"time_get_required_plugins_ms",
		"time_register_resource_ms",
		"time_resource_provider_configure",
		"time_resource_provider_create",
	).ToCSV(os.Stdout)
}
