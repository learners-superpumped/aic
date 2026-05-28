package main

import (
	"fmt"
	"os"

	"github.com/learners-superpumped/aic/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}
