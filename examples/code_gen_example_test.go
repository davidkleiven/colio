package examples

import (
	"bytes"
	"fmt"
	"slices"
	"time"

	"github.com/davidkleiven/colio/parquet"
	parquetgo "github.com/parquet-go/parquet-go"
)

//go:generate go tool soa

type Measurement struct {
	Id    string    `parquet:"id"`
	Time  time.Time `parquet:"time"`
	Value float64   `parqiet:"value"`
}

func ExampleReader() {
	measurements := []Measurement{
		{
			Id:    "M1",
			Time:  time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			Value: 0.0,
		},
		{
			Id:    "M1",
			Time:  time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC),
			Value: 1.0,
		},
		{
			Id:    "M2",
			Time:  time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			Value: 10.0,
		},
		{
			Id:    "M2",
			Time:  time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC),
			Value: 11.0,
		},
	}

	// Write data to parquet
	var buf bytes.Buffer
	pw := parquetgo.NewGenericWriter[Measurement](&buf)
	pw.Write(measurements)
	pw.Close()

	// Read it back as structure of arrays
	var resultSoa MeasurementSoa
	resultSoa.Allocate(10)

	// Reads allocated amount and then passes the soa into the passed processing function
	// (here we just pass nil since we want to read everything, and 10 is higher than the 4 rows)
	parquet.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len())).Bind(&resultSoa).Read[*MeasurementSoa](nil)

	fmt.Print(resultSoa.Id[0])
	fmt.Print(resultSoa.Id[2])

	// Convert to AOS
	result := slices.Collect(resultSoa.ToAos())
	fmt.Printf("\nId=%q", result[0].Id)

	// Output: M1M2
	// Id="M1"

}
