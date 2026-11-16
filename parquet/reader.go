package parquet

import (
	"fmt"
	"io"

	"github.com/davidkleiven/colio/colio"
	parquetgo "github.com/parquet-go/parquet-go"
)

type ColumnDecoder interface {
	Read(n int) (int, error)
}

type columnBinding struct {
	parquetIndex int
	decoder      ColumnDecoder
}

// isColumnPath reports whether path is the leaf path of the flat column name.
// Lists encoded with the standard three level layout have the leaf path
// ["name", "list", "element"], lists encoded as a plain repeated field keep the
// flat path.
func isColumnPath(path []string, name string) bool {
	if path[0] != name {
		return false
	}
	return len(path) == 1 ||
		(len(path) == 3 && path[1] == "list" && path[2] == "element")
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
	batch   colio.Columnar
	columns []columnBinding
	Err     error
}

func (r *Reader) Bind(dst colio.Columnar) *Reader {
	if r.Err != nil {
		return r
	}
	r.batch = dst
	chunks := r.file.RowGroups()[0].ColumnChunks()
	for _, col := range dst.Columns() {
		name := col.ParquetName()
		found := false
		for i, path := range r.schema.Columns() {
			if !isColumnPath(path, name) {
				continue
			}

			leaf, ok := r.schema.Lookup(path...)
			if !ok {
				r.Err = fmt.Errorf("parquet column %q not found in schema", name)
				return r
			}

			decoder, err := newDecoder(chunks[i], col.Ptr, leaf.MaxRepetitionLevel > 0)
			if err != nil {
				r.Err = err
				return r
			}
			r.columns = append(r.columns, columnBinding{parquetIndex: i, decoder: decoder})
			found = true
			break
		}

		if !found {
			r.Err = fmt.Errorf("parquet column %q not found", name)
			return r
		}
	}
	return r
}

func (r *Reader) Read[T colio.Columnar](process func(T, int) error) error {
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
