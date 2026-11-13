package parquet

import (
	"fmt"
	"io"
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

func newDecoder(chunk parquetgo.ColumnChunk, dst any) (ColumnDecoder, error) {
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
		return newTypedDecoder(
			chunk, typedDst,
			func(p parquetgo.Page) ValueReader[time.Time] {
				r := p.Values().(parquetgo.Int64Reader)
				values := make([]int64, p.NumValues())
				t := chunk.Type()
				logicalType := t.LogicalType().String()
				var convert func(int64) time.Time
				switch {
				case strings.Contains(logicalType, parquetgo.Millisecond.TimeUnit().Value.String()):
					convert = func(v int64) time.Time { return time.UnixMilli(v).UTC() }
				case strings.Contains(logicalType, parquetgo.Nanosecond.TimeUnit().Value.String()):
					convert = func(v int64) time.Time { return time.Unix(0, v).UTC() }
				default:
					convert = func(v int64) time.Time { return time.UnixMicro(v).UTC() }
				}
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
