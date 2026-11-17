package colio

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
	f := Must(assets.Open("assets/soa.go.tmpl"))
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

	structs := map[string]*ast.StructType{}
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
			if st, ok := ts.Type.(*ast.StructType); ok {
				structs[ts.Name.Name] = st
			}
		}
	}

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
				Fields:      flattenFields(fset, structs, st),
			}
			result = append(result, data)
		}
	}

	return result
}

func flattenFields(fset *token.FileSet, structs map[string]*ast.StructType, st *ast.StructType) []Field {
	var fields []Field
	for _, field := range st.Fields.List {
		var typeBuf bytes.Buffer
		err := printer.Fprint(&typeBuf, fset, field.Type)
		Assert(err == nil, fmt.Sprintf("print field type: %v", err))
		typeStr := typeBuf.String()

		var tags string
		if field.Tag != nil {
			tags = field.Tag.Value
		}

		if len(field.Names) == 0 {
			name := embeddedTypeName(field.Type)
			if est, ok := structs[name]; ok {
				fields = append(fields, flattenFields(fset, structs, est)...)
			}
			continue
		}

		for _, name := range field.Names {
			if !name.IsExported() {
				continue
			}
			fields = append(fields, Field{
				Name: name.Name,
				Type: typeStr,
				Tags: tags,
			})
		}
	}
	return fields
}

func embeddedTypeName(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return embeddedTypeName(t.X)
	case *ast.SelectorExpr:
		return t.Sel.Name
	}
	return ""
}
