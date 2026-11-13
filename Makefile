.PHONY: generate test

generate:
	go generate ./...

test: generate
	go test -coverprofile=coverage.out -coverpkg=./... ./...
