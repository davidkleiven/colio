package colio

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMakeTemplateDataFlattensEmbedded(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "x.go")
	if err := os.WriteFile(src, []byte(`package p

type Base struct {
	ID   int64  `+"`parquet:\"id\"`"+`
	Name string `+"`parquet:\"name\"`"+`
}

type Child struct {
	Base
	Age int `+"`parquet:\"age\"`"+`
}
`), 0644); err != nil {
		t.Fatal(err)
	}
	data := MakeTemplateData(src)
	var child TemplateData
	for _, d := range data {
		if d.StructName == "Child" {
			child = d
		}
	}
	if len(child.Fields) != 3 {
		t.Fatalf("got %d fields: %+v", len(child.Fields), child.Fields)
	}
	if child.Fields[0].Name != "ID" || child.Fields[1].Name != "Name" || child.Fields[2].Name != "Age" {
		t.Fatalf("unexpected fields: %+v", child.Fields)
	}
}
