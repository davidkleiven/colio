package parquet

//go:generate go run ../cmd/soa

type Person struct {
	Name        string  `parquet:"name"`
	Age         int     `parquet:"age"`
	NumChildren int32   `parquet:"numChildren"`
	Height      float64 `parquet:"height"`
	Weight      float32 `parquet:"weight"`
}
