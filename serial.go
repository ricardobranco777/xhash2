package main

import (
	"os"
	"path/filepath"
)

// MD5All reads all the files in the file tree rooted at root and returns a map
// from file path to the MD5 sum of the file's contents.  If the directory walk
// fails or any read operation fails, MD5All returns an error.
func MD5All_serial(root string) (map[string][]byte, error) {
	m := make(map[string][]byte)
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() {
			return nil
		}
		sum, err := sumFile(path)
		if err != nil {
			return err
		}
		m[path] = sum
		return nil
	})
	if err != nil {
		return nil, err
	}
	return m, nil
}
