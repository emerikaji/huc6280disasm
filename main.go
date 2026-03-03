// from huc6280disasm
// original 26 april 2013
// modified 3 february 2025
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/emerikaji/huc6280disasm/op"
)

func errorf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format, args...)
	os.Exit(1)
}

// command-line options
var (
	useStack = flag.Bool("stack", false, "follow stack for tam/tma values (may fix some broken disassemblies but breaks if some subroutine breaks the push/pop system (TODO add jsr and rts))")
)

func usage() {
	fmt.Fprintf(os.Stderr, "usage: %s [-stack] ROM\n", os.Args[0])
	flag.PrintDefaults()
	os.Exit(1)
}

func main() {
	flag.Usage = usage
	flag.Parse()
	if flag.NArg() != 1 {
		usage()
	}

	filename := flag.Arg(0)

	bytes, err := os.ReadFile(filename)
	if err != nil {
		errorf("error reading input file %s: %v", filename, err)
	}
	if len(bytes) < 0x2000 {
		errorf("given input file %s does not provide a complete interrupt vector table (this restriction may be lifted in the future)", filename)
	}
	if len(bytes) >= 0x1F0000 {
		errorf("given input file %s too large (this restriction may be lifted in the future)", filename)
	}

	run := op.NewRunner(bytes, *useStack)

	if label, err := run.StartDisassembly(); err != nil {
		errorf("internal error: could not get physical address for %s vector (meaning something is up with the paging or the game actually does have the vector outside page 7): %v\n", label, err)
	}
}
