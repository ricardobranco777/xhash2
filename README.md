# xhash2

Improved code for the MD5 examples in [Go Concurrency Patterns: Pipelines and cancellation](https://blog.golang.org/pipelines)

## Main changes:
- Use `io.Copy` instead of `ioutil.ReadFile` to handle very big files.

## TODO
- Support other hashes like SHA-2, SHA-2, Blake2, etc.
- Support options used by other standard utilities.
- Handle stdin.
- Add tests
