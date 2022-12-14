package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/janczizikow/bijou/pkg/cli"
)

const version = "1.0.0"

func main() {
	printVersion := flag.Bool("version", false, "print bijou version")
	flag.Parse()

	args := os.Args[1:]
	if len(args) == 0 {
		fmt.Println("Usage: bijou [options] [file/dir/glob ...]")
		fmt.Printf("\noptions:\n")
		flag.PrintDefaults()
		os.Exit(1)
	}

	if *printVersion {
		fmt.Println(version)
		os.Exit(0)
	}

	os.Exit(cli.Run(args))
}
