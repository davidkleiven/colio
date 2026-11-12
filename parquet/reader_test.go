package parquet

import (
	"bytes"
	"io"
	"testing"

	parquetgo "github.com/parquet-go/parquet-go"
	"github.com/stretchr/testify/require"
)

func TestReadAllInBatches(t *testing.T) {
	data := []Person{
		{
			Name:        "John",
			Age:         52,
			NumChildren: 2,
			Height:      180.0,
			Weight:      82.0,
		},
		{
			Name:        "Susan",
			Age:         48,
			NumChildren: 1,
			Height:      168.0,
			Weight:      68.0,
		},
	}

	var buf bytes.Buffer
	pr := parquetgo.NewGenericWriter[Person](&buf)
	n, err := pr.Write(data)
	require.NoError(t, err)
	require.Equal(t, n, 2)
	pr.Close()

	// Read as columns
	var result PersonSoa
	reader := NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len())).Bind(&result)
	err = reader.Read[*PersonSoa](nil)
	require.ErrorIs(t, err, io.EOF)
	require.Equal(t, 0, result.Cap())

	count := 0
	var got []Person
	result.Allocate(1)
	err = reader.Read(func(batch *PersonSoa, _ int) error {
		count++
		got = append(got, batch.ToAos()...)
		return nil
	})
	require.ErrorIs(t, err, io.EOF)
	require.Equal(t, 2, count)
	require.Equal(t, data, got)

	// When batch size is larger than the file, process is called once
	count = 0
	var big PersonSoa
	big.Allocate(10)
	err = NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len())).Bind(&big).
		Read(func(_ *PersonSoa, _ int) error { count++; return nil })
	require.ErrorIs(t, err, io.EOF)
	require.Equal(t, 1, count)
	require.Equal(t, data, big.ToAos()[:2])
}
