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

type LodArchive struct {
	FilePath      string
	LodType       LodArchiveType
	NumberOfFiles int32
	Files         []LodFile
}

type LodFile struct {
	Name           string
	Offset         int32
	OriginalSize   int32
	CompressedSize int32
}
func (lf LodFile) IsCompressed() bool {
	return lf.CompressedSize != 0
}

func ParseLodFile(pathToLod string) (*LodArchive, error) {
	lodArchive := LodArchive{
		FilePath: pathToLod,
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
	lodArchive.LodType = LodArchiveType(lodType)

	err = readInt32(lodFileReader, &lodArchive.NumberOfFiles)
	if err != nil {
		return nil, fmt.Errorf("can't read numberOfFiles of lod archive(%s): %w", pathToLod, err)
	}
	if lodArchive.NumberOfFiles == 0 {
		return nil, errors.New("lod archive is empty")
	}

	_, err = lodFileReader.Seek(80, io.SeekCurrent)
	if err != nil {
		return nil, fmt.Errorf("can't seek on lod archive(%s): %w", pathToLod, err)
	}

	lodArchive.Files, err = readLodFiles(lodFileReader, lodArchive.NumberOfFiles)
	if err != nil {
		return nil, fmt.Errorf("can't read files of lod archive(%s): %w", pathToLod, err)
	}

	return &lodArchive, nil
}

func readInt32(r io.Reader, o *int32) error {
	return binary.Read(r, binary.LittleEndian, o)
}

func readLodFiles(r *os.File, numberOfFiles int32) ([]LodFile, error) {
	lodFiles := make([]LodFile, 0, numberOfFiles)

	var fi int32
	for fi = 0; fi < numberOfFiles; fi++ {
		lodFile := LodFile{}

		name, err := readLodArchiveFileName(r)
		if err != nil {
			return nil, fmt.Errorf("error reading offset of [%d]: %w", fi, err)
		}
		lodFile.Name = name

		err = readInt32(r, &lodFile.Offset)
		if err != nil {
			return nil, fmt.Errorf("error reading offset of [%d, %s]: %w", fi, name, err)
		}
		err = readInt32(r, &lodFile.OriginalSize)
		if err != nil {
			return nil, fmt.Errorf("error reading originalSize of [%d, %s]: %w", fi, name, err)
		}
		_, err = r.Seek(4, io.SeekCurrent)
		if err != nil {
			return nil, fmt.Errorf("can't seek unknown lod data part of [%d, %s]: %w", fi, name, err)
		}
		err = readInt32(r, &lodFile.CompressedSize)
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
