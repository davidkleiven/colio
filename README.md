# colio

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

colio reads columnar data — Parquet files and ClickHouse queries — directly into
**struct-of-arrays** (SoA) batches instead of one struct per row.

A plain `[]Person` costs one heap allocation per row and scatters a record across
memory. colio keeps each field in its own contiguous slice, so a batch of a million
rows is a handful of slices and the reader hands it to your callback without touching
a single row.

```go
type PersonSoa struct {
    Name  []string
    Age   []int64
    Tags  [][]string
}
```

Your source struct stays an ordinary AoS struct; colio generates the SoA twin from it
and both directions (`ToAos`) are available when you need the familiar shape.

## Install

```sh
go get github.com/davidkleiven/colio
go get -tool github.com/davidkleiven/colio/cmd/soa@latest
```

## Define the record and generate the SoA type

Write your struct with the tags of the backend you read from (`parquet:"…"` or
`ch:"…"`), and add one `go:generate` directive pointing at the generator.

```go
package data

import "time"

//go:generate go tool soa

type Person struct {
	Name        string    `parquet:"name" ch:"name"`
	Age         int64     `parquet:"age" ch:"age"`
	Tags        []string  `parquet:"tags,list" ch:"tags"`
	CreatedAt   time.Time `parquet:"createdAt" ch:"created_at"`
}
```

Then run it through the Go tool:

```sh
go generate ./...
```

This writes `person_soa.go` next to the source file, containing `PersonSoa` with
`Allocate`, `Columns`, `Cap`, `ToAos` and `Delete`. Keep the directive and the
generated file in the same package and re-run `go generate ./...` whenever the struct
changes — the generator reads `GOFILE`/`GOPACKAGE` from the environment, so it is
meant to be invoked by `go generate`, not by running the binary by hand.

To generate only some structs, pass the flags through the directive:

See [Parquet example](examples/code_gen_example_test.go) for a complete example where
a parquet file are read as columns.


## Limitations

- Parquet: a row that spans two legacy data pages (data page v1) loses the elements
  that continue into the second page. Data page v2, which every current writer
  produces, is fully supported.
- ClickHouse: `Array(Nullable(T))` is not supported — a Go slice cannot hold NULLs.
  The same applies to nullable scalar columns.
- Timestamps are returned in UTC from Parquet and in the local zone from ClickHouse,
  matching what the underlying drivers do.

## Benchmarks
<!-- BENCHMARKS:START -->
| Benchmark | Iterations | ns/op | B/op | allocs/op | 
| --- | ---: | ---: | ---: | ---: |
BenchmarkColVsHighlevelApi/high-level-api-8 | 1 | 1418948767 | 1263505912 | 6003491 |
BenchmarkColVsHighlevelApi/col-based-api-8 | 7 | 150725878 | 2784216 | 1474 |
BenchmarkAosVsSoa/soa_read-8 | 15 | 74602760 | 92666712 | 2003232 |
BenchmarkAosVsSoa/aos_read-8 | 1 | 1405769912 | 990258304 | 12002981 |
<!-- BENCHMARKS:END -->

## Development

```sh
make generate # regenerate the *_soa.go files
make test     # generate + unit tests + integration tests (Docker required)
make bench    # benchmarks
```
