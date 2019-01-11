package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/justin177/gogroup"
)

const (
	statusError       = 1
	statusHelp        = 2
	statusInvalidFile = 3
)

func validateOne(proc *gogroup.Processor, file string) (validErr *gogroup.ValidationError, err error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return proc.Validate(file, f)
}

func validateAll(proc *gogroup.Processor, files []string) {
	invalid := false
	for _, file := range files {
		validErr, err := validateOne(proc, file)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(statusError)
		}
		if validErr != nil {
			invalid = true
			fmt.Fprintf(os.Stdout, "%s:%d: %s at %s\n", file, validErr.Line,
				validErr.Message, strconv.Quote(validErr.ImportPath))
		}
	}

	if invalid {
		os.Exit(statusInvalidFile)
	}
}

func rewriteOne(proc *gogroup.Processor, file string) error {
	// Get the rewritten file.
	r, err := func() (io.Reader, error) {
		f, err := os.Open(file)
		if err != nil {
			return nil, err
		}
		defer f.Close()
		return proc.Reformat(file, f)
	}()
	if err != nil {
		return err
	}

	if r != nil {
		// Write the result.
		f, err := os.Create(file)
		if err != nil {
			return err
		}
		defer f.Close()
		_, err = io.Copy(f, r)
		if err != nil {
			return err
		}
		fmt.Fprintf(os.Stderr, "Fixed %s\n", file)
	}
	return nil
}

func rewriteAll(proc *gogroup.Processor, files []string) {
	for _, file := range files {
		err := rewriteOne(proc, file)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(statusError)
		}
	}
}

func main() {
	rewrite := false
	sortByName := false
	gr := gogroup.NewGrouper()

	flag.Usage = func() {
		// Hard to get flag to format long usage well, so just put everything here.
		fmt.Fprintln(os.Stderr,
			`group-imports: Enforce import grouping in Go source files.

Exits with status 3 if import grouping is violated.

Usage: group-imports [OPTIONS] FILE...

  -rewrite
      Instead of checking import grouping, rewrite the source files with
      the correct grouping. Default: false.

  -order SPEC[,SPEC...]
      Modify the import grouping strategy by listing the desired groups in
      order. Group specifications include:

      - std: Standard library imports
      - prefix=PREFIX: Imports whose path starts with PREFIX
      - regexp=REGEXP: Imports whose path matchs with REGEXP
      - other: Imports that match no other specification
      - named: Imports that with name

      These groups can be specified in one comma-separated argument, or
      multiple arguments. Default: std,other
`,
		)
	}

	flag.BoolVar(&rewrite, "rewrite", false, "")
	flag.Var(gr, "order", "")
	flag.BoolVar(&sortByName, "sortByName", false, "")

	flag.Parse()
	if flag.NArg() == 0 {
		fmt.Fprintln(os.Stderr, "No file provided.")
		flag.Usage()
		os.Exit(statusHelp)
	}

	proc := gogroup.NewProcessor(gr, sortByName)
	if rewrite {
		rewriteAll(proc, flag.Args())
	} else {
		validateAll(proc, flag.Args())
	}
}
