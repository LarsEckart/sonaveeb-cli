package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/lars/sonaveeb-cli/cmd"
)

type exitCoder interface {
	error
	ExitCode() int
}

func main() {
	if err := cmd.Run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		var ec exitCoder
		if errors.As(err, &ec) {
			if msg := err.Error(); msg != "" {
				fmt.Fprintln(os.Stderr, msg)
			}
			os.Exit(ec.ExitCode())
		}

		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
