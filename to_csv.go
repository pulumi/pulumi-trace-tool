package main

import (
	"bufio"
	"encoding/csv"
	"os"

	"sourcegraph.com/sourcegraph/appdash"
)

func toCsv(inputTraceFiles []string, outputCsvFile string, filenameColumn string) error {
	annotationNames, err := detectAnnotationNames(inputTraceFiles)
	if err != nil {
		return err
	}

	return writeTracesCsv(annotationNames, inputTraceFiles, outputCsvFile, filenameColumn)
}

func writeTracesCsv(annotationNames []string, inputTraceFiles []string, outputCsvFile, filenameColumn string) error {
	f, err := os.Create(outputCsvFile)
	if err != nil {
		return err
	}
	defer f.Close()

	out := bufio.NewWriter(f)
	w := csv.NewWriter(out)

	var columns []string
	columns = append(columns, annotationNames...)
	if filenameColumn != "" {
		columns = append(columns, filenameColumn)
	}

	err = w.Write(columns)
	if err != nil {
		return err
	}

	i := 0

	for _, inputTraceFile := range inputTraceFiles {
		writeTrace := func(t *appdash.Trace) error {
			i = i + 1
			if i%1024 == 0 {
				w.Flush()
			}
			var values []string
			m := t.Span.Annotations.StringMap()

			for _, a := range annotationNames {
				values = append(values, m[a])
			}

			if filenameColumn != "" {
				values = append(values, inputTraceFile)
			}

			return w.Write(values)
		}

		if err := walkTracesFromFile(inputTraceFile, writeTrace); err != nil {
			return err
		}
	}

	w.Flush()
	return nil
}

func detectAnnotationNames(inputTraceFiles []string) ([]string, error) {
	annotations := make(map[string]int)

	detectAnnotations := func(t *appdash.Trace) error {
		for k, _ := range t.Span.Annotations.StringMap() {
			annotations[k] = annotations[k] + 1
		}
		return nil
	}

	for _, inputTraceFile := range inputTraceFiles {
		if err := walkTracesFromFile(inputTraceFile, detectAnnotations); err != nil {
			return nil, err
		}
	}

	var res []string
	for k, _ := range annotations {
		res = append(res, k)
	}
	return res, nil
}
