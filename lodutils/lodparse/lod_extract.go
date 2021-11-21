package lodparse

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"sync"
)

const defaultConcurrencyLevel = 4

func ExtractLodFiles(lodArchive *LodArchiveMeta, dstDir string) error {
	var concurrencyLevel = defaultConcurrencyLevel
	if len(os.Args) > 3 {
		i, err := strconv.Atoi(os.Args[3])
		if err == nil {
			concurrencyLevel = i
		}
	}
	if concurrencyLevel > int(lodArchive.NumberOfFiles) {
		concurrencyLevel = int(lodArchive.NumberOfFiles)
	}

	var wg sync.WaitGroup
	wg.Add(concurrencyLevel)

	batchSize := int(lodArchive.NumberOfFiles) / concurrencyLevel
	remainingFiles := int(lodArchive.NumberOfFiles) % concurrencyLevel
	for l := 1; l <= concurrencyLevel; l++ {
		start := batchSize * (l-1)
		end   := batchSize * l
		if (l == concurrencyLevel) && (remainingFiles != 0) {
			end += remainingFiles
		}

		go func() {
			defer wg.Done()
			lodFileReader, err := os.Open(lodArchive.ArchiveFilePath)
			if err != nil {
				panic(fmt.Errorf("can't open lod archive(%s): %w", lodArchive.ArchiveFilePath, err))
			}
			defer lodFileReader.Close()

			for _, file := range lodArchive.Files[start:end] {
				err = ExtractFile(file, lodFileReader, dstDir)
				if err != nil {
					panic(fmt.Errorf("can't extract lod archive(%s) file(%s): %w", lodArchive.ArchiveFilePath, file.Name, err))
				}
			}
		}()
	}
	wg.Wait()

	return nil
}

func ExtractFile(file LodFile, lodFileReader *os.File, dstDir string) error {
	fmt.Printf("Processing: %s\n", file.Name)
	var fsize int32
	if file.IsCompressed() {
		fsize = file.CompressedSize
	} else {
		fsize = file.OriginalSize
	}

	_, err := lodFileReader.Seek(int64(file.Offset), io.SeekStart)
	if err != nil {
		return fmt.Errorf("can't seek on lod archive: %w", err)
	}

	fb := make([]byte, fsize)
	_, err = lodFileReader.Read(fb)
	if err != nil {
		return fmt.Errorf("can't read lod archive: %w", err)
	}
	var fbr = ioutil.NopCloser(bytes.NewReader(fb))
	defer fbr.Close()

	if file.IsCompressed() {
		fbr, err = zlib.NewReader(fbr)
		if err != nil {
			return fmt.Errorf("can't create zlib reader during decompressng lod file(%s): %w", file.Name, err)
		}
	}

	return writeFile(file, fbr, dstDir)
}

func writeFile(fileMeta LodFile, bufReader io.Reader, dstDir string) error {
	file, err := os.Create(filepath.Join(dstDir, fileMeta.Name))
	if err != nil {
		return fmt.Errorf("can't create lod file: %w", err)
	}
	defer file.Close()
	_, err = io.Copy(file, bufReader)
	if err != nil {
		return fmt.Errorf("can't write lod file: %w", err)
	}
	return nil
}
