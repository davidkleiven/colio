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
	return tagged(c, "parquet")
}

func (c *Column) ClickHouseName() string {
	return tagged(c, "ch")
}

func tagged(c *Column, key string) string {
	tag := reflect.StructTag(c.Tags)
	if v := tag.Get(key); v != "" {
		return v
	}
	return c.Name
}
