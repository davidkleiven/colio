package parquet

import (
	"bytes"
	"io"
	"testing"
	"time"

	parquetgo "github.com/parquet-go/parquet-go"
	"github.com/parquet-go/parquet-go/compress/zstd"
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
			CreatedAt:   time.Now().UTC(),
		},
		{
			Name:        "Susan",
			Age:         48,
			NumChildren: 1,
			Height:      168.0,
			Weight:      68.0,
			CreatedAt:   time.Now().UTC(),
		},
	}

	for _, test := range []struct {
		Name string
		Opts []parquetgo.WriterOption
	}{
		{
			Name: "none",
		},
		{
			Name: "zstd",
			Opts: []parquetgo.WriterOption{parquetgo.Compression(&zstd.Codec{})},
		},
	} {
		t.Run(test.Name, func(t *testing.T) {
			var buf bytes.Buffer
			pr := parquetgo.NewGenericWriter[Person](&buf, test.Opts...)
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

		})
	}

}

func BenchmarkAosVsSoa(b *testing.B) {
	num := 2_000_000

	rows := make([]Person, num)
	names := []string{"John", "Axel", "Susan"}
	for i := range rows {
		rows[i] = Person{
			Name:        names[i%len(names)],
			Age:         i % 120,
			NumChildren: int32(i % 4),
			Height:      float64(i % 210),
			Weight:      float32(i % 120),
			CreatedAt:   time.Now().UTC(),
		}
	}

	var buf bytes.Buffer
	pw := parquetgo.NewGenericWriter[Person](&buf)
	n, err := pw.Write(rows)
	if err != nil {
		b.Fatal(err)
	}
	if n != num {
		b.Fatalf("Expected %d rows written, wrote %d", num, n)
	}
	pw.Close()

	soaResult := PersonSoa{}
	soaResult.Allocate(num)

	b.Run("soa read", func(b *testing.B) {
		for range b.N {
			err := NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len())).Bind(&soaResult).Read[*PersonSoa](nil)
			if err != io.EOF {
				b.Fatal(err)
			}
		}
	})

	b.Run("aos read", func(b *testing.B) {
		for range b.N {
			_, err := parquetgo.Read[Person](bytes.NewReader(buf.Bytes()), int64(buf.Len()))
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

func TestDelete(t *testing.T) {
	soa := PersonSoa{}
	soa.Allocate(2)
	soa.Name[0] = "John"
	soa.Name[1] = "Susan"
	soa.Delete(0)

	persons := soa.ToAos()
	require.Len(t, persons, 1)
	require.Equal(t, persons[0].Name, "Susan")
}
