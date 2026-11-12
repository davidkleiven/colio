package main

import (
	"bytes"
	"flag"
	"go/format"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/davidkleiven/colio/pkg"
)

func main() {
	source := os.Getenv("GOFILE")
	pkgName := os.Getenv("GOPACKAGE")
	if source == "" {
		log.Fatal("GOFILE not set")
	}

	exclude := flag.String("exclude", "", "Structs to exclude from generation (default none)")
	include := flag.String("include", "", "structs to include (default all)")
	flag.Parse()

	var excludes, includes []string
	if *exclude != "" {
		excludes = strings.Split(*exclude, ",")
	}
	if *include != "" {
		includes = strings.Split(*include, ",")
	}
	data := pkg.MakeTemplateData(source)
	data = slices.DeleteFunc(data, func(d pkg.TemplateData) bool {
		return slices.Contains(excludes, d.StructName) || (len(includes) > 0 && !slices.Contains(includes, d.StructName))
	})

	wd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	var buf bytes.Buffer
	for i, item := range data {
		if i == 0 {
			buf.WriteString("// Generated code by coliosoa. DO NOT EDIT\n\n")
			buf.WriteString("package ")
			buf.WriteString(pkgName)
			buf.WriteString("\nimport \"github.com/davidkleiven/colio/pkg\"")
			buf.WriteString("\n\n")
		}

		pkg.WriteSoa(&buf, item)
		buf.WriteString("\n\n")
	}
	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		log.Fatal(err)
	}
	outputPath := filepath.Join(wd, strings.TrimSuffix(source, ".go")+"_soa.go")
	if err := os.WriteFile(outputPath, formatted, 0644); err != nil {
		log.Fatal(err)
	}
}
