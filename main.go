package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	tr "github.com/pulumi/pulumi-trace-tool/traces"
)

type command struct {
	name string
	run  func(*flag.FlagSet, []string) error
}

func toCsvCommand(flags *flag.FlagSet, args []string) error {
	var outputCsvFile, filenameColumn string

	flags.StringVar(&outputCsvFile, "csv", "", "Path where to write the CSV output file")
	flags.StringVar(&filenameColumn, "filenamecolumn", "tracefile", "Column name to write trace filename to")

	if err := flags.Parse(args); err != nil {
		return err
	}

	traceFiles := flags.Args()

	return tr.ToCsv(traceFiles, outputCsvFile, filenameColumn)
}

func removeLogsCommand(flags *flag.FlagSet, args []string) error {
	var inputFilePath, outputFilePath string

	flags.StringVar(&inputFilePath, "from", "", "Path to the trace file")
	flags.StringVar(&outputFilePath, "to", "", "Path where to write the filtered output trace file")

	if err := flags.Parse(args); err != nil {
		return err
	}

	return tr.RemoveLogs(inputFilePath, outputFilePath)
}

func extractLogsCommand(flags *flag.FlagSet, args []string) error {
	if err := flags.Parse(args); err != nil {
		return err
	}

	for _, f := range flags.Args() {
		if err := tr.ExtractLogs(f); err != nil {
			return err
		}
	}

	return nil
}

func metricsCommand(flags *flag.FlagSet, args []string) error {
	var csvFile, filenameColumn string

	flags.StringVar(&csvFile, "csv", "", "CSV file with data to aggreate into metrics")
	flags.StringVar(&filenameColumn, "filenamecolumn", "tracefile", "Column name where trace filename was recorded")

	if err := flags.Parse(args); err != nil {
		return err
	}

	return tr.Metrics(csvFile, filenameColumn)
}

var commands map[string]command = map[string]command{
	"tocsv":       command{"tocsv", toCsvCommand},
	"removelogs":  command{"removelogs", removeLogsCommand},
	"extractlogs": command{"extractlogs", extractLogsCommand},
	"metrics":     command{"metrics", metricsCommand},
}

func main() {
	exitCannotParseSubcommand := func() {
		var commandNames []string
		for name, _ := range commands {
			commandNames = append(commandNames, name)
		}

		fmt.Printf("expected one of the subcommands: %s\n",
			strings.Join(commandNames, ", "))

		os.Exit(1)
	}

	if len(os.Args) < 2 {
		exitCannotParseSubcommand()
	}

	cmd, gotCmd := commands[os.Args[1]]

	if !gotCmd {
		exitCannotParseSubcommand()
	}

	err := cmd.run(flag.NewFlagSet(cmd.name, flag.ExitOnError), os.Args[2:])
	if err != nil {
		log.Fatal(err)
	}
}
