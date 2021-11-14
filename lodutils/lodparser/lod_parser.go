package lodparser

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
)

// 0x00444f4c is the special value that each lod archive starts with
const lodArchiveHeader int32 = 0x00444f4c

type LodArchiveType int32
const (
	BaseArchive      LodArchiveType = 0x01
	ExpansionArchive LodArchiveType = 0x02
	Unknown          LodArchiveType = 0xff
)

type LodArchive struct {
	lodType       LodArchiveType
	numberOfFiles int32
	files         []LodFile
}

type LodFile struct {
	name           string
	offset         int32
	originalSize   int32
	compressedSize int32
	fileType       int32
}

func ParseLodFile(lodFile string) (*LodArchive, error) {
	var lodArchive LodArchive

	lodFileReader, err := os.Open(lodFile)
	if err != nil {
		return nil, fmt.Errorf("can't open lod archive(%s): %w", lodFile, err)
	}

	defer func() {
		err = lodFileReader.Close()
		if err != nil {
			fmt.Printf("can't close lod archive(%s)\n", lodFile)
		}
	}()

	var lodHeader int32
	err = readInt32(lodFileReader, &lodHeader)
	if err != nil {
		return nil, fmt.Errorf("can't read header of lod archive (%s): %w", lodFile, err)
	}
	if lodHeader != lodArchiveHeader {
		errorMessage := fmt.Sprintf("invalid lod archive(%s), wrong header value", lodFile)
		return nil, errors.New(errorMessage)
	}

	var lodType int32
	err = readInt32(lodFileReader, &lodType)
	if err != nil {
		return nil, fmt.Errorf("can't read type of lod archive(%s): %w", lodFile, err)
	}
	lodArchive.lodType = LodArchiveType(lodType)

	err = readInt32(lodFileReader, &lodArchive.numberOfFiles)
	if err != nil {
		return nil, fmt.Errorf("can't read numberOfFiles of lod archive(%s): %w", lodFile, err)
	}
	if lodArchive.numberOfFiles == 0 {
		return nil, errors.New("lod archive is empty")
	}

	_, err = lodFileReader.Seek(80, io.SeekCurrent)
	if err != nil {
		return nil, fmt.Errorf("can't seek on lod archive(%s): %w", lodFile, err)
	}

	lodArchive.files, err = readLodFiles(lodFileReader, int(lodArchive.numberOfFiles))
	if err != nil {
		return nil, fmt.Errorf("can't read files of lod archive(%s): %w", lodFile, err)
	}

	return &lodArchive, nil
}

func readInt32(r io.Reader, o *int32) error {
	return binary.Read(r, binary.LittleEndian, o)
}

func readLodFiles(r io.Reader, numberOfFiles int) ([]LodFile, error) {
	var lodFiles []LodFile

	for i := 0; i < numberOfFiles; i++ {
		lodFile := LodFile{}

		name, err := readLodArchiveFileName(r)
		if err != nil {
			return nil, fmt.Errorf("error reading offset of [%d]: %w", i, err)
		}
		lodFile.name = name

		err = readInt32(r, &lodFile.offset)
		if err != nil {
			return nil, fmt.Errorf("error reading offset of [%d, %s]: %w", i, name, err)
		}
		err = readInt32(r, &lodFile.originalSize)
		if err != nil {
			return nil, fmt.Errorf("error reading originalSize of [%d, %s]: %w", i, name, err)
		}
		err = readInt32(r, &lodFile.fileType)
		if err != nil {
			return nil, fmt.Errorf("error reading fileType of [%d, %s]: %w", i, name, err)
		}
		err = readInt32(r, &lodFile.compressedSize)
		if err != nil {
			return nil, fmt.Errorf("error reading compressedSize of [%d, %s]: %w", i, name, err)
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

