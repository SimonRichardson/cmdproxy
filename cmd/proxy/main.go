package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"strings"
	"text/tabwriter"
)

var version = "dev"

func usage() {
	fmt.Fprintf(os.Stderr, "USAGE\n")
	fmt.Fprintf(os.Stderr, "  %s <mode> [flags]\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "MODES\n")
	fmt.Fprintf(os.Stderr, "  forward      Forwarding agent\n")
	fmt.Fprintf(os.Stderr, "  agents       Create agents for forwarding\n")
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "VERSION\n")
	fmt.Fprintf(os.Stderr, "  %s (%s)\n", version, runtime.Version())
	fmt.Fprintf(os.Stderr, "\n")
}

func usageFor(fs *flag.FlagSet, name string) func() {
	return func() {
		fmt.Fprintf(os.Stderr, "USAGE\n")
		fmt.Fprintf(os.Stderr, "  %s\n", name)
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "FLAGS\n")

		writer := tabwriter.NewWriter(os.Stderr, 0, 2, 2, ' ', 0)
		fs.VisitAll(func(f *flag.Flag) {
			fmt.Fprintf(writer, "\t-%s %s\t%s\n", f.Name, f.DefValue, f.Usage)
		})
		writer.Flush()

		fmt.Fprintf(os.Stderr, "\n")
	}
}

type command func([]string) error

func (c command) Run(args []string) {
	if err := c(args); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

const (
	defaultAPIPort      = 7650
	defaultAgentAPIPort = 8080
)

var (
	defaultAPIAddr = fmt.Sprintf("tcp://0.0.0.0:%d", defaultAPIPort)
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	var cmd command
	switch strings.ToLower(os.Args[1]) {
	case "forward":
		cmd = runForward
	case "agents":
		cmd = runAgents
	default:
		usage()
		os.Exit(1)
	}

	cmd.Run(os.Args[2:])
}
