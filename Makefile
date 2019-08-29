
BIN = bounded parallel serial

all: $(BIN)

% : %.go utils.go
	go build $< utils.go

clean:
	rm -f $(BIN)
