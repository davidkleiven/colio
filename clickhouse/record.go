package clickhouse

import "time"

//go:generate go run ../cmd/soa

type Person struct {
	Name        string    `ch:"name"`
	Age         int64     `ch:"age"`
	NumChildren int32     `ch:"num_children"`
	Height      float64   `ch:"height"`
	Weight      float32   `ch:"weight"`
	CreatedAt   time.Time `ch:"created_at"`
}
