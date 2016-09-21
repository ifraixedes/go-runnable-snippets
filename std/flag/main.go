package main

import (
	"flag"
	"fmt"
	"os"
	"time"
)

func main() {
	fmt.Println(snippetDesc)

	var defaultFlagSet = flag.NewFlagSet("", flag.ContinueOnError)
	var subcmdFlagSet = flag.NewFlagSet("subcmd", flag.ContinueOnError)

	// empty flag set flags definitions
	var (
		defaultFSName = defaultFlagSet.String(
			"name", "defaultName", "a string flag which specifies some name",
		)
		defaultFSDuration = defaultFlagSet.Duration(
			"wait", time.Duration(0),
			"a duration flag which specifieds some waiting time duration",
		)
	)

	// non empty flag set flags definitions
	var (
		subcmdFSName = subcmdFlagSet.String(
			"name", "defaultName", "a string flag which specifies some name",
		)
		subcmdFSDuration = subcmdFlagSet.Duration(
			"wait", time.Duration(0),
			"a duration flag which specifieds some waiting time duration",
		)
	)

	var (
		defaultFlags []string
		subcmdFlags  []string
		subcmdArgs   bool
	)

	for _, f := range os.Args[1:] {
		if f == "subcmd" {
			subcmdArgs = true
			continue
		}

		if subcmdArgs {
			subcmdFlags = append(subcmdFlags, f)
		} else {
			defaultFlags = append(defaultFlags, f)
		}
	}

	if err := defaultFlagSet.Parse(defaultFlags); err != nil {
		fmt.Sprintf("empty flag set error parsing: %+v", err)
		os.Exit(1)
	}

	if err := subcmdFlagSet.Parse(subcmdFlags); err != nil {
		fmt.Sprintf("non-empty flag set error parsing: %+v", err)
		os.Exit(1)
	}

	fmt.Printf(
		outputMsg, *defaultFSName, defaultFSDuration.String(), *subcmdFSName,
		subcmdFSDuration.String(),
	)
}

const (
	snippetDesc = "This snippet shows a basic example in how to use the flag " +
		"standard package avoiding to use the default set of command-line flags " +
		"and using two flag sets, one for the program command-line flags and " +
		"another one for a subcommand-line flags\n"
	outputMsg = `The flags values are:
	Empty flat set:
		- name: %s
		- wait: %s
	Non-empty flat set:
		- name: %s
		- wait: %s
`
)
