package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/netscrn/homm3utils/lodutils/lodextract"
	"github.com/netscrn/homm3utils/lodutils/lodparse"
)

func main() {
	pathToLod, dstDir := validateAndGetArgs()

	lodArchive, err := lodparse.ParseLodFile(pathToLod)
	if err != nil {
		panic(err.Error())
	}

	err = lodextract.ExtractLodFiles(lodArchive, dstDir)
	if err != nil {
		panic(err.Error())
	}

	fmt.Print("Done")
}

func validateAndGetArgs() (pathToLod string, dstDir string) {
	if len(os.Args) < 3 {
		panic("\n Wrong arguments count \nfirst argument should be path to .lod file. \nsecond argument should be a path to output dir")
	}

	pathToLod = os.Args[1]
	if len(pathToLod) == 0 {
		panic("first argument is empty, should be path to .lod file")
	}
	pathToLod, err := filepath.Abs(pathToLod)
	if err != nil {
		panic(err)
	}
	if _, err := os.Stat(pathToLod); errors.Is(err, os.ErrNotExist) {
		panic("no file exists by path: " + pathToLod)
	}

	dstDir = os.Args[2]
	if len(dstDir) == 0 {
		panic("second argument is empty, should be a path to output dir")
	}
	dstDirInfo, err := os.Stat(dstDir)
	if errors.Is(err, os.ErrNotExist) {
		panic("no directory exists by path: " + pathToLod)
	}
	if !dstDirInfo.IsDir() {
		panic(dstDir + ": is not directory")
	}

	return pathToLod, dstDir
}