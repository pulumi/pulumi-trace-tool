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

		err = walkTraces(traces, func(tr *appdash.Trace) error {
			if !isEngineLogTrace(tr) {
				err = newStore.Collect(tr.ID, tr.Annotations...)
				if err != nil {
					return err
				}
			}
			return nil
		})
		if err != nil {
			return err
		}

		err = writeMemoryStore(outputFilePath, newStore)
		if err != nil {
			return err
		}
	}

	return nil
}
