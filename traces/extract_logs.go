package traces

import (
	"fmt"

	"sourcegraph.com/sourcegraph/appdash"
)

func ExtractLogs(inputFilePath string) error {
	memStore, err := readMemoryStore(inputFilePath)
	if err != nil {
		return err
	}

	traces, err := memStore.Traces(appdash.TracesOpts{})
	if err != nil {
		return err
	}

	err = walkTraces(traces, func(tr *appdash.Trace) error {
		if isEngineLogTrace(tr) {
			var msg, time string
			msg = ""
			time = ""
			for _, ann := range tr.Annotations {
				if ann.Key == "Msg" {
					msg = msg + string(ann.Value)
				}
				if ann.Key == "Time" {
					time = string(ann.Value)
				}
			}
			fmt.Printf("%s\t%s\t%s\n", inputFilePath, time, msg)
		}
		return nil
	})

	if err != nil {
		return err
	}

	return nil
}
