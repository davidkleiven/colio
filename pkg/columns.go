package pkg

import "reflect"

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
	tag := reflect.StructTag(c.Tags)
	if v := tag.Get("parquet"); v != "" {
		return v
	}
	return c.Name
}
