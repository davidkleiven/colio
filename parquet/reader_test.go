package parquet

import (
	"bytes"
	"fmt"
	"io"
	"slices"
	"testing"
	"time"

	"github.com/davidkleiven/colio/colio"
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
				got = append(got, slices.Collect(batch.ToAos())...)
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
			require.Equal(t, data, slices.Collect(big.ToAos())[:2])
		})
	}

}

func TestReadListColumns(t *testing.T) {
	moment := func(ms int64) time.Time { return time.UnixMilli(ms).UTC() }
	data := []Event{
		{Name: "first", Tags: []int{1, 2, 3}, Steps: []float64{1.5}, Times: []time.Time{moment(1000)}},
		{Name: "empty", Tags: []int{}, Steps: nil, Times: []time.Time{}},
		{Name: "second", Tags: nil, Steps: []float64{0.5, 2.5}, Times: []time.Time{moment(3000), moment(4000)}},
	}

	// Parquet does not distinguish nil from empty lists
	want := slices.Clone(data)
	want[1].Steps = []float64{}
	want[2].Tags = []int{}

	var buf bytes.Buffer
	pr := parquetgo.NewGenericWriter[Event](&buf)
	n, err := pr.Write(data)
	require.NoError(t, err)
	require.Equal(t, 3, n)
	pr.Close()

	for _, batch := range []int{1, 3, 10} {
		t.Run(fmt.Sprintf("batch-%d", batch), func(t *testing.T) {
			var got []Event
			var result EventSoa
			result.Allocate(batch)
			err := NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len())).Bind(&result).
				Read(func(soa *EventSoa, n int) error {
					got = append(got, slices.Collect(soa.ToAos())[:n]...)
					return nil
				})
			require.ErrorIs(t, err, io.EOF)
			require.Equal(t, want, got)
		})
	}
}

func TestReadListColumnsAcrossPages(t *testing.T) {
	numRows := 400
	data := make([]Event, numRows)
	for i := range data {
		data[i] = Event{
			Name:  fmt.Sprintf("event-%d", i),
			Tags:  []int{i, i + 1, i + 2},
			Steps: []float64{float64(i), float64(i) / 2},
			Times: []time.Time{time.UnixMilli(int64(i)).UTC()},
		}
	}

	var buf bytes.Buffer
	pr := parquetgo.NewGenericWriter[Event](&buf, parquetgo.PageBufferSize(1024))
	n, err := pr.Write(data)
	require.NoError(t, err)
	require.Equal(t, numRows, n)
	pr.Close()

	// Pages end on row boundaries, so every column contributes whole rows page
	// after page: make sure the file actually spans more than one page.
	file, err := parquetgo.OpenFile(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	require.NoError(t, err)
	for i, chunk := range file.RowGroups()[0].ColumnChunks() {
		pages := chunk.Pages()
		count := 0
		for {
			if _, err := pages.ReadPage(); err != nil {
				break
			}
			count++
		}
		require.NoError(t, pages.Close())
		require.Greater(t, count, 1, "column %d should span multiple pages", i)
	}

	var got []Event
	var result EventSoa
	result.Allocate(64)
	err = NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len())).Bind(&result).
		Read(func(soa *EventSoa, n int) error {
			got = append(got, slices.Collect(soa.ToAos())[:n]...)
			return nil
		})
	require.ErrorIs(t, err, io.EOF)
	require.Equal(t, data, got)
}

// Columns with a shape that does not match the file they are bound to.
type listOnScalarSoa struct {
	Name [][]string
}

func (v *listOnScalarSoa) Columns() []colio.Column {
	return []colio.Column{{Name: "Name", Tags: `parquet:"name"`, Ptr: &v.Name}}
}

func (v *listOnScalarSoa) Cap() int { return len(v.Name) }

type scalarOnListSoa struct {
	Steps []float64
}

func (v *scalarOnListSoa) Columns() []colio.Column {
	return []colio.Column{{Name: "Steps", Tags: `parquet:"steps"`, Ptr: &v.Steps}}
}

func (v *scalarOnListSoa) Cap() int { return len(v.Steps) }

func TestBindListMismatch(t *testing.T) {
	var buf bytes.Buffer
	pr := parquetgo.NewGenericWriter[Event](&buf)
	_, err := pr.Write([]Event{{Name: "a", Tags: []int{1}}})
	require.NoError(t, err)
	pr.Close()
	data, size := buf.Bytes(), int64(buf.Len())

	t.Run("list into scalar column", func(t *testing.T) {
		var dst listOnScalarSoa
		dst.Name = make([][]string, 1)
		err := NewReader(bytes.NewReader(data), size).Bind(&dst).Err
		require.ErrorContains(t, err, "cannot decode list *[][]string from a non-repeated parquet column")
	})

	t.Run("scalar into list column", func(t *testing.T) {
		var dst scalarOnListSoa
		dst.Steps = make([]float64, 1)
		err := NewReader(bytes.NewReader(data), size).Bind(&dst).Err
		require.ErrorContains(t, err, "cannot decode repeated parquet column into *[]float64")
	})
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

	persons := slices.Collect(soa.ToAos())
	require.Len(t, persons, 1)
	require.Equal(t, persons[0].Name, "Susan")
}

var (
	_ colio.Aoser[Person] = (*PersonSoa)(nil)
	_                     = EmbeddedParentSoa{}.Name // Verify parent gets field of child
)
