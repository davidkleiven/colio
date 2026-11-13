package clickhouse

import (
	"context"
	"fmt"
	"io"
	"slices"
	"time"

	"github.com/ClickHouse/ch-go"
	"github.com/ClickHouse/ch-go/proto"
	"github.com/davidkleiven/colio/pkg"
)

type columnBinding struct {
	name  string
	store func(proto.ColResult) error
}

func NewReader(client *ch.Client, query string) *Reader {
	return &Reader{client: client, query: query}
}

type Reader struct {
	client  *ch.Client
	query   string
	batch   pkg.Columnar
	columns []columnBinding
	Err     error
}

func (r *Reader) Bind(dst pkg.Columnar) *Reader {
	if r.Err != nil {
		return r
	}
	r.batch = dst
	for _, col := range dst.Columns() {
		store, err := newStore(col.Ptr)
		if err != nil {
			r.Err = fmt.Errorf("clickhouse column %q: %w", col.Name, err)
			return r
		}
		r.columns = append(r.columns, columnBinding{name: col.ClickHouseName(), store: store})
	}
	return r
}

func (r *Reader) Read[T pkg.Columnar](ctx context.Context, process func(T, int) error) error {
	if r.Err != nil {
		return r.Err
	}

	typedBatch, ok := r.batch.(T)
	if !ok {
		var zero T
		return fmt.Errorf("Incompatible types: %T", zero)
	}

	var results proto.Results
	err := r.client.Do(ctx, ch.Query{
		Body:   r.query,
		Result: results.Auto(),
		OnResult: func(_ context.Context, block proto.Block) error {
			if block.Rows == 0 {
				return nil
			}
			if err := r.decode(results); err != nil {
				return err
			}
			if process == nil {
				return nil
			}
			return process(typedBatch, block.Rows)
		},
	})
	if err != nil {
		return fmt.Errorf("clickhouse: %w", err)
	}
	return io.EOF
}

func (r *Reader) decode(results proto.Results) error {
	for _, binding := range r.columns {
		i := slices.IndexFunc(results, func(col proto.ResultColumn) bool {
			return col.Name == binding.name
		})
		if i < 0 {
			return fmt.Errorf("clickhouse column %q not found", binding.name)
		}
		if err := binding.store(results[i].Data); err != nil {
			return fmt.Errorf("clickhouse column %q: %w", binding.name, err)
		}
	}
	return nil
}

func newStore(dst any) (func(proto.ColResult) error, error) {
	switch ptr := dst.(type) {
	case *[]string:
		return copyValues(ptr), nil
	case *[]int32:
		return copyValues(ptr), nil
	case *[]int64:
		return copyValues(ptr), nil
	case *[]float32:
		return copyValues(ptr), nil
	case *[]float64:
		return copyValues(ptr), nil
	case *[]time.Time:
		return copyValues(ptr), nil
	case *[]int:
		return func(col proto.ColResult) error {
			values, ok := col.(valueColumn[int64])
			if !ok {
				return fmt.Errorf("cannot decode %s into %T", col.Type(), ptr)
			}
			n := values.Rows()
			*ptr = resize(*ptr, n)
			for i := range n {
				(*ptr)[i] = int(values.Row(i))
			}
			return nil
		}, nil
	default:
		return nil, fmt.Errorf("unsupported type %T", dst)
	}
}

type valueColumn[T any] interface {
	Rows() int
	Row(i int) T
}

func copyValues[T any](dst *[]T) func(proto.ColResult) error {
	return func(col proto.ColResult) error {
		values, ok := col.(valueColumn[T])
		if !ok {
			return fmt.Errorf("cannot decode %s into %T", col.Type(), dst)
		}
		n := values.Rows()
		*dst = resize(*dst, n)
		for i := range n {
			(*dst)[i] = values.Row(i)
		}
		return nil
	}
}

func resize[T any](s []T, n int) []T {
	if cap(s) >= n {
		return s[:n]
	}
	return make([]T, n)
}
