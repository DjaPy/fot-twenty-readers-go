package service

import (
	"os"
	"path/filepath"
)

type PathForFile struct {
	baseDir      string
	outFile      string
	templatesDir string
}

func getPathForFile() PathForFile {
	currentFile, err := os.Executable()
	if err != nil {
		panic(err)
	}
	baseDir := filepath.Dir(currentFile)
	outFile := filepath.Join(baseDir, "..", "graph_of_reading_of_the_psalter.xlsx")
	templatesDir := filepath.Join(baseDir, "entrypoints", "templates")
	var p PathForFile
	p.baseDir = baseDir
	p.outFile = outFile
	p.templatesDir = templatesDir
	return p
}
