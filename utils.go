package main

import (
	"crypto"
	_ "crypto/md5"
	_ "crypto/sha1"
	_ "crypto/sha256"
	_ "crypto/sha512"
	_ "golang.org/x/crypto/sha3"
	"hash"
	"io"
	"os"
)

var algorithms = map[string]struct {
	hash crypto.Hash
	hash.Hash
}{
	"MD5":        {hash: crypto.MD5},
	"SHA1":       {hash: crypto.SHA1},
	"SHA224":     {hash: crypto.SHA224},
	"SHA256":     {hash: crypto.SHA256},
	"SHA384":     {hash: crypto.SHA384},
	"SHA512":     {hash: crypto.SHA512},
	"SHA512-224": {hash: crypto.SHA512_224},
	"SHA512-256": {hash: crypto.SHA512_256},
	"SHA3-224":   {hash: crypto.SHA3_224},
	"SHA3-256":   {hash: crypto.SHA3_256},
	"SHA3-384":   {hash: crypto.SHA3_384},
	"SHA3-512":   {hash: crypto.SHA3_512},
}

func sumFile(path string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	h := algorithms[algorithm].hash.New()
	if _, err := io.Copy(h, file); err != nil {
		return nil, err
	}

	return h.Sum(nil), nil
}
