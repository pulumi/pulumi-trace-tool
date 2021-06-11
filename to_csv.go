package main

import (
	"bufio"
	"encoding/csv"
	"os"

	"sourcegraph.com/sourcegraph/appdash"
)

func toCsv(inputTraceFile, outputCsvFile string) error {

	annotationNames, err := detectAnnotationNames(inputTraceFile)
	if err != nil {
		return err
	}

	return writeTracesCsv(annotationNames, inputTraceFile, outputCsvFile)
}

func writeTracesCsv(annotationNames []string, inputTraceFile, outputCsvFile string) error {
	f, err := os.Create(outputCsvFile)
	if err != nil {
		return err
	}
	defer f.Close()

	out := bufio.NewWriter(f)
	w := csv.NewWriter(out)

	var columns []string
	for _, a := range annotationNames {
		columns = append(columns, a)
	}

	err = w.Write(columns)
	if err != nil {
		return err
	}

	i := 0
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

		return w.Write(values)
	}

	if err := walkTracesFromFile(inputTraceFile, writeTrace); err != nil {
		return err
	}

	w.Flush()
	return nil
}

func detectAnnotationNames(inputTraceFile string) ([]string, error) {
	annotations := make(map[string]int)

	detectAnnotations := func(t *appdash.Trace) error {
		for k, _ := range t.Span.Annotations.StringMap() {
			annotations[k] = annotations[k] + 1
		}
		return nil
	}

	if err := walkTracesFromFile(inputTraceFile, detectAnnotations); err != nil {
		return nil, err
	}

	var res []string
	for k, _ := range annotations {
		res = append(res, k)
	}
	return res, nil
}
