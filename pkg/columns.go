package pkg

import (
	"reflect"
	"strings"
)

type Columnar interface {
	Columns() []Column
	Cap() int
}

type Column struct {
	Name string
	Tags string
	Ptr  any
}

func (c *Column) ParquetName() string {
	return tagged(c, "parquet")
}

func (c *Column) ClickHouseName() string {
	return tagged(c, "ch")
}

func tagged(c *Column, key string) string {
	tag := reflect.StructTag(c.Tags)
	// Options may follow the column name, e.g. `parquet:"age,list"`, only the
	// part before the first comma identifies the column.
	name, _, _ := strings.Cut(tag.Get(key), ",")
	if name == "" {
		return c.Name
	}
	return name
}
