package main

import (
	"os"

	"github.com/janczizikow/bijou/pkg/cli"
)

func getVersion() string {
	if v := os.Getenv("BIJOU_VERSION"); v != "" {
		return v
	}
	return "dev"
}

func main() {
	app := cli.New(getVersion())
	os.Exit(app.Run(os.Args[1:]))
}
