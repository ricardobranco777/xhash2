xhash2: *.go
	go build

test:
	go vet

clean:
	rm -f xhash2
