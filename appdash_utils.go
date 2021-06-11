package main

import (
	"bufio"
	// "flag"
	// "fmt"
	// "log"
	"os"

	"sourcegraph.com/sourcegraph/appdash"
)

func walkTrace(trace *appdash.Trace, onTrace func(x *appdash.Trace) error) error {
	err := onTrace(trace)
	if err != nil {
		return err
	}
	err = walkTraces(trace.Sub, onTrace)
	if err != nil {
		return err
	}
	return nil
}

func walkTraces(traces []*appdash.Trace, onTrace func(x *appdash.Trace) error) error {
	for _, t := range traces {
		err := walkTrace(t, onTrace)
		if err != nil {
			return err
		}
	}
	return nil
}

func walkTracesFromFile(file string, onTrace func(x *appdash.Trace) error) error {
	memStore, err := readMemoryStore(file)
	if err != nil {
		return err
	}

	traces, err := memStore.Traces(appdash.TracesOpts{})
	if err != nil {
		return err
	}

	return walkTraces(traces, onTrace)
}

func writeMemoryStore(filepath string, memStore *appdash.MemoryStore) error {
	f, err := os.Create(filepath)
	if err != nil {
		return err
	}
	return memStore.Write(f)
}

func readMemoryStore(filePath string) (*appdash.MemoryStore, error) {
	memStore := appdash.NewMemoryStore()

	inputFile, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}

	_, err = memStore.ReadFrom(bufio.NewReader(inputFile))
	if err != nil {
		return nil, err
	}

	return memStore, nil
}
