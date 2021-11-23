package lodparse

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/netscrn/homm3utils/internal/binread"
)

// 0x00444f4c is the special value that each lod archive starts with
const lodArchiveHeader uint32 = 0x00444f4c

func parseLodFile(pathToLod string) (*LodArchiveMeta, error) {
	lodArchiveMeta := LodArchiveMeta{
		ArchiveFilePath: pathToLod,
	}

	lodFileReader, err := os.Open(pathToLod)
	if err != nil {
		return nil, fmt.Errorf("can't open lod archive(%s): %w", pathToLod, err)
	}
	defer lodFileReader.Close()

	var lodHeader uint32
	err = binread.ReadUint32(lodFileReader, &lodHeader)
	if err != nil {
		return nil, fmt.Errorf("can't read header of lod archive (%s): %w", pathToLod, err)
	}
	if lodHeader != lodArchiveHeader {
		errorMessage := fmt.Sprintf("invalid lod archive(%s), wrong header value", pathToLod)
		return nil, errors.New(errorMessage)
	}

	var lodType uint32
	err = binread.ReadUint32(lodFileReader, &lodType)
	if err != nil {
		return nil, fmt.Errorf("can't read type of lod archive(%s): %w", pathToLod, err)
	}
	lodArchiveMeta.LodType = LodArchiveType(lodType)

	err = binread.ReadUint32(lodFileReader, &lodArchiveMeta.NumberOfFiles)
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
	lodArchiveMeta.indexFiles()

	return &lodArchiveMeta, nil
}

func readLodFiles(laf *os.File, numberOfFiles uint32) ([]LodFileMeta, error) {
	lodFiles := make([]LodFileMeta, 0, numberOfFiles)

	var fi uint32
	for fi = 0; fi < numberOfFiles; fi++ {
		lodFile := LodFileMeta{}

		name, err := binread.ReadAvailableChars(laf, 16)
		if err != nil {
			return nil, fmt.Errorf("error reading offset of [%d]: %w", fi, err)
		}
		lodFile.Name = name

		err = binread.ReadUint32(laf, &lodFile.Offset)
		if err != nil {
			return nil, fmt.Errorf("error reading offset of [%d, %s]: %w", fi, name, err)
		}
		err = binread.ReadUint32(laf, &lodFile.OriginalSize)
		if err != nil {
			return nil, fmt.Errorf("error reading originalSize of [%d, %s]: %w", fi, name, err)
		}
		_, err = laf.Seek(4, io.SeekCurrent)
		if err != nil {
			return nil, fmt.Errorf("can't seek unknown lod data part of [%d, %s]: %w", fi, name, err)
		}
		err = binread.ReadUint32(laf, &lodFile.CompressedSize)
		if err != nil {
			return nil, fmt.Errorf("error reading compressedSize of [%d, %s]: %w", fi, name, err)
		}

		lodFiles = append(lodFiles, lodFile)
	}

	return lodFiles, nil
}
