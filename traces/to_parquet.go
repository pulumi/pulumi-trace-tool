package traces

func ToParquet(inputCsvFile, outputParquetFile string) error {
	var data []map[string]string
	err := readLargeCsvFile(inputCsvFile, func(row map[string]string) error {
		data = append(data, row)
		return nil
	})
	if err != nil {
		return nil
	}

	sink := NewParquetFileMetricsSink(outputParquetFile)
	err = sink.writeMetrics(data)
	if err != nil {
		return err
	}

	return nil
}
