package main

import (
	"os"

	"github.com/janczizikow/bijou/pkg/cli"
)

const version = "1.0.0"

func main() {
	app := cli.New(version)
	os.Exit(app.Run(os.Args[1:]))
}
