package lodextract

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/netscrn/homm3utils/lodutils/lodparse"
)

const defaultConcurrencyLevel = 4

func ExtractLodFiles(lodArchiveRef *lodparse.LodArchive, dstDir string) error {
	lodArchive := *lodArchiveRef

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
			lodFileReader, err := os.Open(lodArchive.FilePath)
			if err != nil {
				panic(fmt.Errorf("can't open lod archive(%s): %w", lodArchive.FilePath, err))
			}
			defer lodFileReader.Close()

			for _, file := range lodArchive.Files[start:end] {
				err = ExtractFile(file, dstDir, lodFileReader)
				if err != nil {
					panic(fmt.Errorf("can't extract lod archive(%s) file(%s): %w", lodArchive.FilePath, file.Name, err))
				}
			}
			wg.Done()
		}()
	}
	wg.Wait()

	return nil
}

func ExtractFile(file lodparse.LodFile, dstDir string, lodFileReader *os.File) error {
	fmt.Printf("Processing: %s\n", file.Name)
	var fileSize int32
	if file.IsCompressed() {
		fileSize = file.CompressedSize
	} else {
		fileSize = file.OriginalSize
	}

	_, err := lodFileReader.Seek(int64(file.Offset), io.SeekStart)
	if err != nil {
		return fmt.Errorf("can't seek on lod archive: %w", err)
	}

	fileBuf := make([]byte, fileSize)
	_, err = lodFileReader.Read(fileBuf)
	if err != nil {
		return fmt.Errorf("can't read lod archive: %w", err)
	}

	if file.IsCompressed() {
		err = writeCompressedFile(fileBuf, file, dstDir)
	} else {
		err = writeFile(fileBuf, file, dstDir)
	}
	return err
}

func writeCompressedFile(data []byte, fileMeta lodparse.LodFile, dstDir string) error {
	br := bytes.NewReader(data)
	zr, err := zlib.NewReader(br)
	if err != nil {
		return fmt.Errorf("can't write compressed lod file: %w", err)
	}
	defer zr.Close()
	file, err := os.Create(filepath.Join(dstDir, fileMeta.Name))
	if err != nil {
		return fmt.Errorf("can't write compressed lod file: %w", err)
	}
	defer file.Close()
	_, err = io.Copy(file, zr)
	if err != nil {
		return fmt.Errorf("can't write compressed lod file: %w", err)
	}
	return nil
}

func writeFile(data []byte, fileMeta lodparse.LodFile, dstDir string) error {
	err := os.WriteFile(filepath.Join(dstDir, fileMeta.Name), data,0644)
	if err != nil {
		return fmt.Errorf("can't write raw lod file: %w", err)
	}
	return nil
}
