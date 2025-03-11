package main

import (
	"fmt"
	"os"

	"github.com/davidkohl/gobelix/idefix/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
