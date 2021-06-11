package main

import (
	"flag"
	//"fmt"
	"log"
	//"os"
)

func main() {
	var inputTraceFile, outputCsvFile string
	flag.StringVar(&inputTraceFile, "trace", "", "Path to the trace file produced by Pulumi")
	flag.StringVar(&outputCsvFile, "csv", "", "Path where to write the CSV output file")
	flag.Parse()

	err := toCsv(inputTraceFile, outputCsvFile)
	if err != nil {
		log.Fatal(err)
	}
}

// func main2() {
// 	var inputFilePath, outputFilePath, logPath string
// 	flag.StringVar(&inputFilePath, "trace", "", "Path to the trace file")
// 	flag.StringVar(&outputFilePath, "out", "", "Path where to write the filtered output trace file")
// 	flag.StringVar(&logPath, "log", "", "Path where to write log output to")
// 	flag.Parse()

// 	memStore, err := readMemoryStore(inputFilePath)
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	traces, err := memStore.Traces(appdash.TracesOpts{})
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	counts := make(map[string]int)

// 	walkTraces(traces, func(x *appdash.Trace) error {
// 		if !isEngineLogTrace(x) {
// 			for _, ann := range x.Annotations {
// 				if ann.Key == "Name" {
// 					count := counts[string(ann.Value)]
// 					counts[string(ann.Value)] = count + 1
// 				}
// 			}
// 		}
// 		return nil
// 	})

// 	//fmt.Printf("memStore: %d traces found as %d\n", i, len(traces))

// 	for k, v := range counts {
// 		fmt.Printf("%s = %d\n", k, v)
// 	}

// 	if logPath != "" {
// 		f, err := os.Create(logPath)
// 		if err != nil {
// 			log.Fatal(err)
// 		}

// 		walkTraces(traces, func(tr *appdash.Trace) error {
// 			if isEngineLogTrace(tr) {
// 				var msg, time string
// 				msg = ""
// 				time = ""
// 				for _, ann := range tr.Annotations {
// 					//fmt.Printf("  %s ==> %s\n", ann.Key, ann.Value)
// 					if ann.Key == "Msg" {
// 						msg = msg + string(ann.Value)
// 					}
// 					if ann.Key == "Time" {
// 						time = string(ann.Value)
// 					}
// 				}
// 				//fmt.Printf("%s\t%s\n", time, msg)
// 				fmt.Fprintf(f, "%s\t%s\n", time, msg)
// 			}
// 			return nil
// 		})
// 	}

// 	if outputFilePath != "" {
// 		newStore := appdash.NewMemoryStore()

// 		walkTraces(traces, func(tr *appdash.Trace) error {
// 			if !isEngineLogTrace(tr) {
// 				newStore.Collect(tr.ID, tr.Annotations...)
// 			}
// 			return nil
// 		})

// 		err = writeMemoryStore(outputFilePath, newStore)
// 		if err != nil {
// 			log.Fatal(err)
// 		}
// 		fmt.Printf("Written %s\n", outputFilePath)
// 	}
// }

// func isEngineLogTrace(trace *appdash.Trace) bool {
// 	for _, ann := range trace.Annotations {
// 		if ann.Key == "Name" && string(ann.Value) == "/pulumirpc.Engine/Log" {
// 			return true
// 		}
// 	}
// 	return false
// }
