package main

import (
	"fmt"
	"os"
	"regexp"
)

var REGEXP = regexp.MustCompile(`public/(\S+/\S+)`)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Not enough args")
		os.Exit(1)
	}

	if len(os.Args) > 3 {
		fmt.Println("Too many args")
		os.Exit(1)
	}

	arg := os.Args[1]

	match := REGEXP.FindStringSubmatch(arg)
	if match == nil {
		fmt.Println("URL is incorrect")
		os.Exit(1)
	}

	tree, err := GetFiles(match[1])
	if err != nil {
		panic(err)
	}

	DownloadFiles(".", tree)
}
