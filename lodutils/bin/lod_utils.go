package main

import (
	"errors"
	"os"

	"github.com/netscrn/homm3utils/lodutils/lodfilesextractor"
	"github.com/netscrn/homm3utils/lodutils/lodparser"
)

func main() {
	pathToLod, outputDir := validateAndGetArgs()

	lodArchive, err := lodparser.ParseLodFile(pathToLod)
	if err != nil {
		panic(err.Error())
	}

	err = lodfilesextractor.ExtractLodFiles(lodArchive, outputDir)
	if err != nil {
		panic(err.Error())
	}
}

func validateAndGetArgs() (pathToLod string, outputDir string) {
	if len(os.Args) < 3 {
		panic("\n Wrong arguments count \nfirst argument should be path to .lod file. \nsecond argument should be a path to output dir")
	}

	pathToLod = os.Args[1]
	if len(pathToLod) == 0 {
		panic("first argument is empty, should be path to .lod file")
	}
	if _, err := os.Stat(pathToLod); errors.Is(err, os.ErrNotExist) {
		panic("no file exists by path: " + pathToLod)
	}

	outputDir = os.Args[2]
	if len(outputDir) == 0 {
		panic("second argument is empty, should be a path to output dir")
	}
	outputDirInfo, err := os.Stat(outputDir)
	if errors.Is(err, os.ErrNotExist) {
		panic("no directory exists by path: " + pathToLod)
	}
	if !outputDirInfo.IsDir() {
		panic(outputDir + ": is not directory")
	}

	return pathToLod, outputDir
}