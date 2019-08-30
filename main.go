package main

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
)

var algorithm string
var method string
var numDigesters = 20

func init() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS] DIRECTORY\n\n", os.Args[0])
		flag.PrintDefaults()
	}

	for h := range algorithms {
		flag.StringVar(&algorithm, strings.ToLower(h), "MD5", h+" algorithm")
	}
	for _, s := range []string{"serial", "bounded", "parallel"} {
		flag.StringVar(&method, s, "", s)
	}
}

func main() {
	flag.Parse()

	// Calculate the MD5 sum of all files under the specified directory,
	// then print the results sorted by path name.

	var f func(string) (map[string][]byte, error)

	if len(flag.Args()) != 1 {
		flag.PrintDefaults()
	}

	switch method {
	case "serial":
		f = MD5All_serial
	case "parallel":
		f = MD5All_parallel
	case "bounded":
		if _, ok := os.LookupEnv("DIGESTERS"); ok {
			var err error
			numDigesters, err = strconv.Atoi(os.Getenv("DIGESTERS"))
			if err != nil || numDigesters < 1 {
				fmt.Fprintf(os.Stderr, "Invalid value for DIGESTERS\n")
				os.Exit(1)
			}
		}
		f = MD5All_bounded
	default:
		panic("Invalid method")
	}

	m, err := f(os.Args[2])
	if err != nil {
		fmt.Println(err)
		return
	}
	var paths []string
	for path := range m {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	for _, path := range paths {
		fmt.Printf("%x  %s\n", m[path], path)
	}
}
