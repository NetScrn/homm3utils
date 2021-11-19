package lodparse

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
)

// 0x00444f4c is the special value that each lod archive starts with
const lodArchiveHeader int32 = 0x00444f4c

func parseLodFile(pathToLod string) (*LodArchiveMeta, error) {
	lodArchiveMeta := LodArchiveMeta{
		ArchiveFilePath: pathToLod,
	}

	lodFileReader, err := os.Open(pathToLod)
	if err != nil {
		return nil, fmt.Errorf("can't open lod archive(%s): %w", pathToLod, err)
	}
	defer lodFileReader.Close()

	var lodHeader int32
	err = readInt32(lodFileReader, &lodHeader)
	if err != nil {
		return nil, fmt.Errorf("can't read header of lod archive (%s): %w", pathToLod, err)
	}
	if lodHeader != lodArchiveHeader {
		errorMessage := fmt.Sprintf("invalid lod archive(%s), wrong header value", pathToLod)
		return nil, errors.New(errorMessage)
	}

	var lodType int32
	err = readInt32(lodFileReader, &lodType)
	if err != nil {
		return nil, fmt.Errorf("can't read type of lod archive(%s): %w", pathToLod, err)
	}
	lodArchiveMeta.LodType = LodArchiveType(lodType)

	err = readInt32(lodFileReader, &lodArchiveMeta.NumberOfFiles)
	if err != nil {
		return nil, fmt.Errorf("can't read numberOfFiles of lod archive(%s): %w", pathToLod, err)
	}
	if lodArchiveMeta.NumberOfFiles == 0 {
		return nil, errors.New("lod archive is empty")
	}

	_, err = lodFileReader.Seek(80, io.SeekCurrent)
	if err != nil {
		return nil, fmt.Errorf("can't seek on lod archive(%s): %w", pathToLod, err)
	}

	lodArchiveMeta.Files, err = readLodFiles(lodFileReader, lodArchiveMeta.NumberOfFiles)
	if err != nil {
		return nil, fmt.Errorf("can't read files of lod archive(%s): %w", pathToLod, err)
	}

	return &lodArchiveMeta, nil
}

func readInt32(r io.Reader, o *int32) error {
	return binary.Read(r, binary.LittleEndian, o)
}

func readLodFiles(laf *os.File, numberOfFiles int32) ([]LodFile, error) {
	lodFiles := make([]LodFile, 0, numberOfFiles)

	var fi int32
	for fi = 0; fi < numberOfFiles; fi++ {
		lodFile := LodFile{}

		name, err := readLodArchiveFileName(laf)
		if err != nil {
			return nil, fmt.Errorf("error reading offset of [%d]: %w", fi, err)
		}
		lodFile.Name = name

		err = readInt32(laf, &lodFile.Offset)
		if err != nil {
			return nil, fmt.Errorf("error reading offset of [%d, %s]: %w", fi, name, err)
		}
		err = readInt32(laf, &lodFile.OriginalSize)
		if err != nil {
			return nil, fmt.Errorf("error reading originalSize of [%d, %s]: %w", fi, name, err)
		}
		_, err = laf.Seek(4, io.SeekCurrent)
		if err != nil {
			return nil, fmt.Errorf("can't seek unknown lod data part of [%d, %s]: %w", fi, name, err)
		}
		err = readInt32(laf, &lodFile.CompressedSize)
		if err != nil {
			return nil, fmt.Errorf("error reading compressedSize of [%d, %s]: %w", fi, name, err)
		}

		lodFiles = append(lodFiles, lodFile)
	}

	return lodFiles, nil
}

func readLodArchiveFileName(r io.Reader) (string, error) {
	nameBuf := make([]byte, 16)
	_, err := r.Read(nameBuf)
	if err != nil {
		return "", err
	}

	var nameLen int
	for i, b := range nameBuf {
		if b == 0 {
			nameLen = i
			break
		}
	}

	return string(nameBuf[:nameLen]), nil
}
