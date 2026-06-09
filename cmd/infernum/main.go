package main

import (
	"github.com/joelhelbling/infernum/internal/cli"
)

// Set via ldflags at build time
var version = "dev"

func main() {
	cli.Version = version
	cli.Execute()
}
