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

func Metrics(csvFile string, filenameColumn string, writer io.Writer) error {

	f, err := os.Open(csvFile)
	if err != nil {
		return err
	}

	defer f.Close()

	r := bufio.NewReader(f)

	cr := csv.NewReader(r)

	header, err := cr.Read()
	if err != nil {
		return err
	}

	inColumns := make(map[string]int)
	outColumns := make(map[string]int)

	cw := csv.NewWriter(writer)
	for i, h := range header {
		inColumns[h] = i
		outColumns[filenameColumn] = len(outColumns) - 1
		outColumns["time_ms"] = len(outColumns) - 1

		if strings.HasPrefix(h, "MemStats.") ||
			strings.HasPrefix(h, "benchmark_") ||
			strings.HasPrefix(h, "runtime.") ||
			h == "repo" ||
			h == "os.Args" ||
			h == "Span.Start" ||
			h == "Span.End" {
			outColumns[h] = len(outColumns) - 1
		}
	}

	outColumnNames := make([]string, len(outColumns))
	for k, j := range outColumns {
		outColumnNames[j] = k
	}

	if err := cw.Write(outColumnNames); err != nil {
		return err
	}
	cw.Flush()

	for {
		record, err := cr.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		if record[inColumns["Name"]] == "pulumi" {

			values := make([]string, len(outColumns))

			for k, j := range outColumns {
				values[j] = record[inColumns[k]]
			}

			spanStart, err := parseTime(values[outColumns["Span.Start"]])
			if err != nil {
				return err
			}
			spanEnd, err := parseTime(values[outColumns["Span.End"]])
			if err != nil {
				return err
			}
			dur := spanEnd.Sub(spanStart)

			values[outColumns["time_ms"]] = fmt.Sprintf("%v",
				int64(dur/time.Millisecond))

			cw.Write(values)
			cw.Flush()
		}

	}

	return nil
}

func parseTime(str string) (time.Time, error) {
	return time.Parse(time.RFC3339, str)
}
