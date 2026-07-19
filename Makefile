.PHONY: build test vet fmt run

build:
	go build -o drift-check .

test:
	go test ./...

vet:
	go vet ./...

fmt:
	gofmt -l .

run: build
	./drift-check
