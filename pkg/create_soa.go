package pkg

import (
	"bytes"
	"embed"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"io"
	"text/template"
)

//go:embed assets/*
var assets embed.FS

func MustLoadSoaTempl() *template.Template {
	f := Must(assets.Open("assets/soa.tmpl"))
	defer f.Close()
	content := Must(io.ReadAll(f))
	return Must(template.New("soa").Parse(string(content)))
}

type Field struct {
	Name string
	Type string
	Tags string
}

type TemplateData struct {
	PackageName string
	StructName  string
	Fields      []Field
}

func WriteSoa(w io.Writer, data TemplateData) error {
	temp := MustLoadSoaTempl()
	return temp.Execute(w, data)
}

func MakeTemplateData(filename string) []TemplateData {
	fset := token.NewFileSet()

	file := Must(parser.ParseFile(fset, filename, nil, parser.ParseComments))

	var result []TemplateData

	for _, decl := range file.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok || gen.Tok != token.TYPE {
			continue
		}

		for _, spec := range gen.Specs {
			ts, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}

			st, ok := ts.Type.(*ast.StructType)
			if !ok {
				continue
			}

			data := TemplateData{
				PackageName: file.Name.Name,
				StructName:  ts.Name.Name,
			}

			for _, field := range st.Fields.List {
				if len(field.Names) == 0 {
					continue
				}

				var typeBuf bytes.Buffer
				err := printer.Fprint(&typeBuf, fset, field.Type)
				Assert(err == nil, fmt.Sprintf("print field type: %v", err))

				tags := ""
				if field.Tag != nil {
					tags = field.Tag.Value
				}

				for _, name := range field.Names {
					if !name.IsExported() {
						continue
					}

					data.Fields = append(data.Fields, Field{
						Name: name.Name,
						Type: typeBuf.String(),
						Tags: tags,
					})
				}
			}

			result = append(result, data)
		}
	}

	return result
}
