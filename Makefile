BIN := subagents
CMD := ./cmd/subagents

.PHONY: build test vet clean

build:
	go build -o $(BIN) $(CMD)

test:
	go test ./...

vet:
	go vet ./...

clean:
	rm -f $(BIN)
