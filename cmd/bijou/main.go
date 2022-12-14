package main

import (
	"flag"
	"fmt"
	"os"
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
		return
	}

	if *printVersion {
		fmt.Println(version)
		os.Exit(0)
	}

	for _, arg := range args {
		fmt.Println(arg)
	}
}
