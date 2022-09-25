package main

import (
	"flag"
	"fmt"
	"os"
	"regexp"
)

var REGEXP = regexp.MustCompile(`public/(\S+/\S+)`)

var (
	help = flag.Bool("help", false, "Display this message")
	h    = flag.Bool("h", false, "Display this message (shorthand)")

	output = flag.String("output", ".", "Output folder")
	o      = flag.String("o", ".", "Output folder (shorthand)")
)

func main() {
	flag.Parse()

	if *help || *h {
		flag.Usage()
		os.Exit(0)
	}

	arg := flag.Arg(0)

	match := REGEXP.FindStringSubmatch(arg)
	if match == nil {
		fmt.Println("URL is incorrect")
		os.Exit(1)
	}

	tree, err := GetFiles(match[1])
	if err != nil {
		panic(err)
	}

	var path string

	if *output != "." {
		path = *output
	}

	if *o != "." {
		path = *o
	}

	DownloadFiles(path, tree)
}
