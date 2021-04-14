package main

import (
	"flag"
	"fmt"
)

func main() {
	path := flag.String("repo-root", ".", "path to root directory of the repository to generate config for")

	flag.Parse()

	fmt.Println(*path)
}
