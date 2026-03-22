package main

import (
	"os"

	"github.com/mauricioTechDev/propcheck-ai/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
