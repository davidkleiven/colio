package parquet

import "time"

//go:generate go run ../cmd/soa

type Person struct {
	Name        string    `parquet:"name"`
	Age         int       `parquet:"age"`
	NumChildren int32     `parquet:"numChildren"`
	Height      float64   `parquet:"height"`
	Weight      float32   `parquet:"weight"`
	CreatedAt   time.Time `parquet:"createdAt"`
}

type Event struct {
	Name  string      `parquet:"name"`
	Tags  []int       `parquet:"tags,list"`
	Steps []float64   `parquet:"steps"`
	Times []time.Time `parquet:"times,list"`
}
