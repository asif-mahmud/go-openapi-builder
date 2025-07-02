package main

import (
	"embed"
	"fmt"
	"io"

	goopenapibuilder "github.com/asif-mahmud/go-openapi-builder"
)

//go:embed sample
var doc embed.FS

func main() {
	r, e := goopenapibuilder.BuildFromFS(doc)
	if e != nil {
		panic(e)
	}
	d, e := io.ReadAll(r)
	if e != nil {
		panic(e)
	}
	fmt.Println(string(d))
}
