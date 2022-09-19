package main

import (
	"flag"

	tr "github.com/pulumi/pulumi-trace-tool/traces"
)

func summaryCommand(flags *flag.FlagSet, args []string) error {
	var filenameColumn string
	flags.StringVar(&filenameColumn, "filenamecolumn", "filename", "Column name to write trace filename to")

	if err := flags.Parse(args); err != nil {
		return err
	}

	traceFiles := flags.Args()

	return tr.Summary(traceFiles, filenameColumn)
}
