package main

// A result is the product of reading and summing a file using MD5.
type result struct {
	path string
	sum  []byte
	err  error
}
