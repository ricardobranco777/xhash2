package main

import (
	"bufio"
	"crypto"
	_ "crypto/md5"
	_ "crypto/sha1"
	_ "crypto/sha256"
	_ "crypto/sha512"
	"errors"
	"flag"
	"fmt"
	_ "golang.org/x/crypto/sha3"
	"hash"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type info struct {
	sum  []byte
	hash crypto.Hash
	hash.Hash
}

type hashes map[crypto.Hash]*info

// A result is the product of reading and summing a file using MD5.
type result struct {
	path   string
	hashes hashes
	err    error
}

var progname string

var opts struct {
	all bool
}

var algorithms = map[crypto.Hash]*struct {
	name  string
	check bool
}{
	crypto.MD5:        {name: "MD5"},
	crypto.SHA1:       {name: "SHA1"},
	crypto.SHA224:     {name: "SHA224"},
	crypto.SHA256:     {name: "SHA256"},
	crypto.SHA384:     {name: "SHA384"},
	crypto.SHA512:     {name: "SHA512"},
	crypto.SHA512_224: {name: "SHA512-224"},
	crypto.SHA512_256: {name: "SHA512-256"},
	crypto.SHA3_224:   {name: "SHA3-224"},
	crypto.SHA3_256:   {name: "SHA3-256"},
	crypto.SHA3_384:   {name: "SHA3-384"},
	crypto.SHA3_512:   {name: "SHA3-512"},
}

func getHashes() hashes {
	hashes := make(hashes)

	for algorithm, h := range algorithms {
		if !h.check {
			continue
		}
		hashes[algorithm] = &info{hash: algorithm}
	}

	return hashes
}

func sumSmallFileF(f *os.File) (hashes, error) {
	var wg sync.WaitGroup

	data, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}

	hashes := getHashes()

	for algorithm := range hashes {
		wg.Add(1)
		go func(algorithm crypto.Hash) {
			defer wg.Done()
			h := hashes[algorithm].hash.New()
			h.Write(data)
			hashes[algorithm].sum = h.Sum(nil)
		}(algorithm)
	}

	wg.Wait()
	return hashes, nil
}

func sumFileF(f *os.File) (hashes, error) {
	var wg sync.WaitGroup
	var writers []io.Writer
	var pipeWriters []*io.PipeWriter

	hashes := getHashes()

	for algorithm := range hashes {
		pr, pw := io.Pipe()
		writers = append(writers, pw)
		pipeWriters = append(pipeWriters, pw)
		wg.Add(1)
		go func(algorithm crypto.Hash) {
			defer wg.Done()
			h := hashes[algorithm].hash.New()
			if _, err := io.Copy(h, pr); err != nil {
				panic(err)
			}
			hashes[algorithm].sum = h.Sum(nil)
		}(algorithm)
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer func() {
			for _, pw := range pipeWriters {
				pw.Close()
			}
		}()

		// build the multiwriter for all the pipes
		mw := io.MultiWriter(writers...)

		// copy the data into the multiwriter
		if _, err := io.Copy(mw, f); err != nil {
			panic(err)
		}
	}()

	wg.Wait()
	return hashes, nil
}

func sumFile(path string, info fs.FileInfo) (hashes, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	if info.Size() < 1e6 {
		return sumSmallFileF(file)
	} else {
		return sumFileF(file)
	}
}

func sumFilesFromArgs(_unused string, WalkFn filepath.WalkFunc) error {
	for _, path := range flag.Args() {
		// XXX: Use os.Lstat()
		info, err := os.Stat(path)
		if err := WalkFn(path, info, err); err != nil {
			return err
		}
	}
	return nil
}

func sumFilesFromFile(filename string, WalkFn filepath.WalkFunc) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	reader := bufio.NewReader(file)

	for {
		line, _, err := reader.ReadLine()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}
		path := string(line)
		// XXX: Use os.Lstat()
		info, err := os.Stat(path)
		if err := WalkFn(path, info, err); err != nil {
			return err
		}
	}

	return nil
}

// sumFiles starts goroutines to walk the directory tree at root and digest each
// regular file.  These goroutines send the results of the digests on the result
// channel and send the result of the walk on the error channel.  If done is
// closed, sumFiles abandons its work.
func sumFiles(done <-chan struct{}, root string) (<-chan result, <-chan error) {
	// For each regular file, start a goroutine that sums the file and sends
	// the result on c.  Send the result of the walk on errc.
	c := make(chan result)
	errc := make(chan error, 1)
	go func() {
		var wg sync.WaitGroup
		//f := sumFilesFromFile
		// f := sumFilesFromArgs
		f := filepath.WalkDir
		err := f(root, func(path string, entry fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			info, err := entry.Info()
			if err != nil {
				return err
			}
			if !info.Mode().IsRegular() {
				return nil
			}
			wg.Add(1)
			go func() {
				sums, err := sumFile(path, info)
				select {
				case c <- result{path, sums, err}:
				case <-done:
				}
				wg.Done()
			}()
			// Abort the walk if done is closed.
			select {
			case <-done:
				return errors.New("walk canceled")
			default:
				return nil
			}
		})
		// Walk has returned, so all calls to wg.Add are done.  Start a
		// goroutine to close c once all the sends are done.
		go func() {
			wg.Wait()
			close(c)
		}()
		// No select needed here, since errc is buffered.
		errc <- err
	}()
	return c, errc
}

// MD5All reads all the files in the file tree rooted at root and prints the
// the MD5 sum of the file's contents.  If the directory walk
// fails or any read operation fails, MD5All returns an error.  In that case,
// MD5All does not wait for inflight read operations to complete.
func MD5All(root string) error {
	// MD5All closes the done channel when it returns; it may do so before
	// receiving all the values from c and errc.
	done := make(chan struct{})
	defer close(done)

	c, errc := sumFiles(done, root)

	for r := range c {
		if r.err != nil {
			return r.err
		}
		for algorithm, _ := range r.hashes {
			fmt.Printf("%s(%s) = %x\n", algorithms[algorithm].name, r.path, r.hashes[algorithm].sum)
		}
	}
	if err := <-errc; err != nil {
		return err
	}
	return nil
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s DIRECTORY\n", os.Args[0])
	os.Exit(1)
}

func init() {
	progname = filepath.Base(os.Args[0])

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS] [-s STRING...]|[FILE... DIRECTORY...]\n\n", progname)
		flag.PrintDefaults()
	}

	flag.BoolVar(&opts.all, "all", false, "all algorithms (except others specified, if any)")

	for _, h := range algorithms {
		flag.BoolVar(&h.check, strings.ToLower(h.name), false, h.name+" algorithm")
	}
}

func main() {
	flag.Parse()

	// TODO: Support default algorithm
	if opts.all {
		for _, h := range algorithms {
			h.check = !h.check
		}
	}

	err := MD5All(flag.Args()[0])
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
