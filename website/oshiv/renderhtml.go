package main

import (
	"os"
	"text/template"
)

func main() {
	version := os.Args[1]
	templateFile := os.Args[2]

	template, err := template.New(templateFile).ParseFiles(templateFile)
	if err != nil {
		panic(err)
	}

	var indexFile *os.File
	indexFile, err = os.Create("index.html")
	if err != nil {
		panic(err)
	}

	err = template.Execute(indexFile, version)
	if err != nil {
		panic(err)
	}
}
