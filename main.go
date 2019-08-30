package main

import (
	"fmt"
	"os"
	"sort"
)

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s [-serial|-parallel|-bounded] DIRECTORY\n", os.Args[0])
	os.Exit(1)
}

func main() {
	// Calculate the MD5 sum of all files under the specified directory,
	// then print the results sorted by path name.

	var f func(string) (map[string][]byte, error)

	if len(os.Args) != 3 {
		usage()
	}

	switch os.Args[1] {
	case "-serial":
		f = MD5All_serial
	case "-parallel":
		f = MD5All_parallel
	case "-bounded":
		f = MD5All_bounded
	default:
		usage()
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
