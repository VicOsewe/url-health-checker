package main

import (
	"fmt"
	"os"

	flagcli "github.com/VicOsewe/url-health-checker/cli/flag"
)

func main() {

	cli := flagcli.New()

	if err := cli.Run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
