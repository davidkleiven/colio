package parquet

import (
	"fmt"
	"io"

	"github.com/davidkleiven/colio/pkg"
	parquetgo "github.com/parquet-go/parquet-go"
)

type ColumnDecoder interface {
	Read(n int) (int, error)
}

type columnBinding struct {
	parquetIndex int
	decoder      ColumnDecoder
}

func NewReader(r io.ReaderAt, size int64) *Reader {
	reader := &Reader{}

	file, err := parquetgo.OpenFile(r, size)
	if err != nil {
		reader.Err = fmt.Errorf("open parquet: %w", err)
		return reader
	}

	reader.file = file
	reader.schema = file.Schema()

	return reader
}

type Reader struct {
	file    *parquetgo.File
	schema  *parquetgo.Schema
	batch   pkg.Columnar
	columns []columnBinding
	Err     error
}

func (r *Reader) Bind(dst pkg.Columnar) *Reader {
	if r.Err != nil {
		return r
	}
	r.batch = dst
	chunks := r.file.RowGroups()[0].ColumnChunks()
	for _, col := range dst.Columns() {
		name := col.ParquetName()
		found := false
		for i, path := range r.schema.Columns() {
			if len(path) == 1 && path[0] == name {
				decoder, err := newDecoder(chunks[i], col.Ptr)
				if err != nil {
					r.Err = err
					return r
				}
				r.columns = append(r.columns, columnBinding{parquetIndex: i, decoder: decoder})
				found = true
				break
			}

			if found {
				break
			}
		}

		if !found {
			r.Err = fmt.Errorf("parquet column %q not found", name)
			return r
		}
	}
	return r
}

func (r *Reader) Read[T pkg.Columnar](process func(T, int) error) error {
	if r.Err != nil {
		return r.Err
	}

	typedBatch, ok := r.batch.(T)
	if !ok {
		var zero T
		return fmt.Errorf("Incompatible types: %T", zero)
	}

	n := r.batch.Cap()
	for {
		for _, col := range r.columns {
			read, err := col.decoder.Read(n)
			if err != nil && err != io.EOF {
				return err
			}
			if read != n {
				n = read
			}
		}
		if n == 0 {
			return io.EOF
		}
		if process != nil {
			if err := process(typedBatch, n); err != nil {
				return fmt.Errorf("process: %w", err)
			}
		}
	}
}
