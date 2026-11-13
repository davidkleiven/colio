.PHONY: generate test

generate:
	go generate ./...

test: generate
	go test -coverprofile=coverage.out -coverpkg=./... ./...

bench: generate
	go test -bench=. -benchmem -run=^$$ ./...
