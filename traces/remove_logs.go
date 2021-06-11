package traces

import (
	"sourcegraph.com/sourcegraph/appdash"
)

func RemoveLogs(inputFilePath, outputFilePath string) error {
	memStore, err := readMemoryStore(inputFilePath)
	if err != nil {
		return err
	}

	traces, err := memStore.Traces(appdash.TracesOpts{})
	if err != nil {
		return err
	}

	if outputFilePath != "" {
		newStore := appdash.NewMemoryStore()

		walkTraces(traces, func(tr *appdash.Trace) error {
			if !isEngineLogTrace(tr) {
				newStore.Collect(tr.ID, tr.Annotations...)
			}
			return nil
		})

		err = writeMemoryStore(outputFilePath, newStore)
		if err != nil {
			return err
		}
	}

	return nil
}
