
BINS = bounded parallel serial

all: $(BINS)

% : %.go utils.go
	go build $< utils.go

clean:
	rm $(BINS)
