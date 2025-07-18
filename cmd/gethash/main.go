package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/mamaart/nxcd/internal/core"
)

func main() {
	flag.Parse()
	repo := flag.Arg(0)
	if repo == "" {
		fmt.Fprintln(os.Stderr, "missing repo")
		flag.Usage()
		os.Exit(1)
	}
	cli, err := core.Client()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	hash, err := core.GetSha(cli, repo)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Print(hash)
}
