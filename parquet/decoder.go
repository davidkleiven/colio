package parquet

import (
	"fmt"
	"io"
	"reflect"
	"strings"
	"time"

	parquetgo "github.com/parquet-go/parquet-go"
)

type ValueReader[T any] func([]T) (int, error)

func (r ValueReader[T]) ReadValues(dst []T) (int, error) {
	return r(dst)
}

type Decoder[T any] struct {
	pages  parquetgo.Pages
	page   parquetgo.Page
	reader ValueReader[T]
	dst    *[]T
	read   func(parquetgo.Page) ValueReader[T]
}

func (d *Decoder[T]) Read(n int) (int, error) {
	out := (*d.dst)[:n]
	total := 0

	for total < n {
		if d.page == nil {
			page, err := d.pages.ReadPage()
			if err != nil {
				return total, err
			}
			d.page = page
			d.reader = d.read(page)
		}

		count, err := d.reader.ReadValues(out[total:])
		total += count

		if err == io.EOF {
			parquetgo.Release(d.page)
			d.page = nil
			d.reader = nil
			continue
		}

		if err != nil {
			return total, err
		}
	}

	return total, nil
}

func newTypedDecoder[T any](
	chunk parquetgo.ColumnChunk,
	dst *[]T,
	read func(parquetgo.Page) ValueReader[T],
) *Decoder[T] {
	return &Decoder[T]{
		pages: chunk.Pages(),
		dst:   dst,
		read:  read,
	}
}

func newDecoder(chunk parquetgo.ColumnChunk, dst any, repeated bool) (ColumnDecoder, error) {
	if isListDst(dst) {
		if !repeated {
			return nil, fmt.Errorf("cannot decode list %T from a non-repeated parquet column", dst)
		}
		return newListDecoder(chunk, dst)
	}
	if repeated {
		return nil, fmt.Errorf("cannot decode repeated parquet column into %T", dst)
	}

	switch chunk.Type().Kind() {
	case parquetgo.Int32:
		return newTypedDecoder(
			chunk, dst.(*[]int32),
			func(p parquetgo.Page) ValueReader[int32] {
				r := p.Values().(parquetgo.Int32Reader)
				return r.ReadInt32s
			},
		), nil

	case parquetgo.Int64:
		return newInt64Decoder(chunk, dst)
	case parquetgo.Float:
		return newTypedDecoder(
			chunk, dst.(*[]float32),
			func(p parquetgo.Page) ValueReader[float32] {
				r := p.Values().(parquetgo.FloatReader)
				return r.ReadFloats
			},
		), nil

	case parquetgo.Double:
		return newTypedDecoder(
			chunk, dst.(*[]float64),
			func(p parquetgo.Page) ValueReader[float64] {
				r := p.Values().(parquetgo.DoubleReader)
				return r.ReadDoubles
			},
		), nil
	case parquetgo.ByteArray:
		return newTypedDecoder(
			chunk, dst.(*[]string),
			func(p parquetgo.Page) ValueReader[string] {
				r := p.Values()
				values := make([]parquetgo.Value, p.NumValues())
				return func(dst []string) (int, error) {
					buf := values[:min(len(values), len(dst))]
					n, err := r.ReadValues(buf)

					for i := range n {
						dst[i] = string(buf[i].ByteArray())
					}

					return n, err
				}
			},
		), nil

	}

	return nil, fmt.Errorf("unsupported physical type: %s", chunk.Type())
}

func newInt64Decoder(chunk parquetgo.ColumnChunk, dst any) (ColumnDecoder, error) {
	switch typedDst := dst.(type) {
	case *[]int64:
		return newTypedDecoder(
			chunk, typedDst,
			func(p parquetgo.Page) ValueReader[int64] {
				r := p.Values().(parquetgo.Int64Reader)
				return r.ReadInt64s
			},
		), nil
	case *[]int:
		return newTypedDecoder(
			chunk, typedDst,
			func(p parquetgo.Page) ValueReader[int] {
				r := p.Values().(parquetgo.Int64Reader)
				values := make([]int64, p.NumValues())
				return func(dst []int) (int, error) {
					buf := values[:min(len(values), len(dst))]
					n, err := r.ReadInt64s(buf)
					for i := range n {
						dst[i] = int(values[i])
					}
					return n, err
				}
			},
		), nil
	case *[]time.Time:
		convert := timeConverter(chunk.Type())
		return newTypedDecoder(
			chunk, typedDst,
			func(p parquetgo.Page) ValueReader[time.Time] {
				r := p.Values().(parquetgo.Int64Reader)
				values := make([]int64, p.NumValues())
				return func(dst []time.Time) (int, error) {
					buf := values[:min(len(values), len(dst))]
					n, err := r.ReadInt64s(buf)
					for i := range n {
						dst[i] = convert(values[i])
					}
					return n, err
				}
			},
		), nil
	default:
		return nil, fmt.Errorf("Unsupported type %T", typedDst)
	}
}

// timeConverter returns the conversion from the raw int64 values stored in a
// column of type t into time.Time, based on the unit of the timestamp.
func timeConverter(t parquetgo.Type) func(int64) time.Time {
	logicalType := t.LogicalType().String()
	switch {
	case strings.Contains(logicalType, parquetgo.Millisecond.TimeUnit().Value.String()):
		return func(v int64) time.Time { return time.UnixMilli(v).UTC() }
	case strings.Contains(logicalType, parquetgo.Nanosecond.TimeUnit().Value.String()):
		return func(v int64) time.Time { return time.Unix(0, v).UTC() }
	default:
		return func(v int64) time.Time { return time.UnixMicro(v).UTC() }
	}
}

// isListDst reports whether dst points at a column of lists, e.g. *[][]int.
func isListDst(dst any) bool {
	t := reflect.TypeOf(dst).Elem()
	return t.Kind() == reflect.Slice && t.Elem().Kind() == reflect.Slice
}

// newListDecoder dispatches on the element type of the destination list. The
// repetition and definition levels of the column are needed to group the values
// into rows, so the values are read through the generic parquetgo.Value reader
// rather than through the typed readers used for scalar columns.
func newListDecoder(chunk parquetgo.ColumnChunk, dst any) (ColumnDecoder, error) {
	switch typedDst := dst.(type) {
	case *[][]int:
		return makeListDecoder(chunk, typedDst, parquetgo.Int64,
			func(v parquetgo.Value) int { return int(v.Int64()) })

	case *[][]int32:
		return makeListDecoder(chunk, typedDst, parquetgo.Int32,
			func(v parquetgo.Value) int32 { return v.Int32() })

	case *[][]int64:
		return makeListDecoder(chunk, typedDst, parquetgo.Int64,
			func(v parquetgo.Value) int64 { return v.Int64() })

	case *[][]float32:
		return makeListDecoder(chunk, typedDst, parquetgo.Float,
			func(v parquetgo.Value) float32 { return v.Float() })

	case *[][]float64:
		return makeListDecoder(chunk, typedDst, parquetgo.Double,
			func(v parquetgo.Value) float64 { return v.Double() })

	case *[][]string:
		return makeListDecoder(chunk, typedDst, parquetgo.ByteArray,
			func(v parquetgo.Value) string { return string(v.ByteArray()) })

	case *[][]time.Time:
		convert := timeConverter(chunk.Type())
		return makeListDecoder(chunk, typedDst, parquetgo.Int64,
			func(v parquetgo.Value) time.Time { return convert(v.Int64()) })
	}

	return nil, fmt.Errorf("unsupported type %T for repeated parquet column", dst)
}

// makeListDecoder decodes a repeated column into one list per row. A new row
// starts at every value with a repetition level of zero, values that are null
// (which is how parquet represents empty lists) are skipped.
func makeListDecoder[E any](
	chunk parquetgo.ColumnChunk,
	dst *[][]E,
	kind parquetgo.Kind,
	get func(parquetgo.Value) E,
) (ColumnDecoder, error) {
	if chunk.Type().Kind() != kind {
		return nil, fmt.Errorf("cannot decode %s column into %T", chunk.Type(), dst)
	}

	// Pages start and end on row boundaries (data page v2), so the state needed
	// to group the values into rows is scoped to a single page.
	return newTypedDecoder(chunk, dst, func(p parquetgo.Page) ValueReader[[]E] {
		values := p.Values()
		buf := make([]parquetgo.Value, max(1, min(int(p.NumValues()), 4096)))
		var (
			partial []E // elements of the row currently being built
			open    bool
			i, n    int
			eof     bool
		)

		commit := func(out [][]E) {
			if partial == nil {
				partial = []E{} // parquet does not distinguish nil from empty lists
			}
			out[0] = partial
			partial = nil
		}

		return func(out [][]E) (int, error) {
			rows := 0
			for rows < len(out) {
				if i == n {
					if eof {
						// shortcut: elements of a row that continue into the next
						// page (legacy data page v1 files) are dropped, only the
						// number of rows is preserved. Buffer the row across
						// pages when such files show up.
						if !open {
							return rows, io.EOF
						}
						commit(out[rows:])
						rows++
						open = false
						continue
					}
					var err error
					n, err = values.ReadValues(buf)
					i = 0
					if err == io.EOF {
						eof = true
					} else if err != nil {
						return rows, err
					}
					continue
				}

				v := buf[i]
				i++
				if v.RepetitionLevel() == 0 {
					if open {
						commit(out[rows:])
						rows++
					}
					open = true
				}
				if open && !v.IsNull() {
					partial = append(partial, get(v))
				}
			}
			return rows, nil
		}
	}), nil
}
