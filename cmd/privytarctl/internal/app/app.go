// Package app is the main package for the application.
package app

import (
	"flag"
	"fmt"
	"io"
	"os"

	"git.sr.ht/~jamesponddotco/privytar/cmd/privytarctl/internal/meta"
	"git.sr.ht/~jamesponddotco/xstd-go/xerrors"
)

// ErrConfigPathRequired is returned when the user fails to provide a path to
// the configuration file.
const ErrConfigPathRequired xerrors.Error = "missing configuration file; use the --config flag to provide one"

// Usage returns the usage information for the application.
func Usage(w io.Writer) {
	text := `NAME:
   %s - %s

USAGE:
   %s [global options] command

VERSION:
   %s

COMMANDS:
   start         start the server for the Privytar service
   stop          stop the server for the Privytar service

GLOBAL OPTIONS:
   --config value, -c value  path to configuration file
   --help, -h                show help
   --version, -v             print the version
`

	fmt.Fprintf(w, text, meta.Name, meta.Description, meta.Name, meta.Version)
}

// Run is the entry point for the application.
func Run() int {
	var (
		configFlag  = flag.String("config", "", "path to configuration file")
		helpFlag    = flag.Bool("help", false, "show help")
		versionFlag = flag.Bool("version", false, "print the version")
	)

	flag.Parse()

	if *helpFlag {
		Usage(os.Stdout)

		return 0
	}

	if *versionFlag {
		fmt.Fprintf(os.Stdout, "%s\n", meta.Version)

		return 0
	}

	if flag.NArg() < 1 {
		Usage(os.Stderr)

		return 1
	}

	switch flag.Arg(0) {
	case "start":
		if err := StartAction(*configFlag); err != nil {
			fmt.Fprintf(os.Stderr, "error: %s\n", err)
		}

		return 0
	case "stop":
		if err := StopAction(*configFlag); err != nil {
			fmt.Fprintf(os.Stderr, "error: %s\n", err)
		}

		return 0
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", flag.Arg(0))

		Usage(os.Stderr)

		return 1
	}
}
