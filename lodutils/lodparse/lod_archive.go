package lodparse

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

func LoadLodArchiveMetaFromJson(pathFoJson string) (*LodArchiveMeta, error) {
	jsonFile, err := os.Open(pathFoJson)
	if err != nil {
		return nil, err
	}
	defer jsonFile.Close()

	var lodArchiveMeta LodArchiveMeta
	err = json.NewDecoder(jsonFile).Decode(&lodArchiveMeta)
	if err != nil {
		return nil, err
	}

	return &lodArchiveMeta, nil
}

func LoadLodArchiveMetaFromLodFile(pathToLod string) (*LodArchiveMeta, error) {
	return parseLodFile(pathToLod)
}

type lodArchiveFileIndex map[string]int
type LodArchiveMeta struct {
	ArchiveFilePath string         `json:"file_path"`
	LodType         LodArchiveType `json:"lod_type"`
	NumberOfFiles   int32         `json:"number_of_files"`
	Files           []LodFileMeta `json:"files"`
	filesIndexes    lodArchiveFileIndex
}

func (lam *LodArchiveMeta) indexFiles() {
	lam.filesIndexes = make(lodArchiveFileIndex, len(lam.Files))
	for i, file := range lam.Files {
		lam.filesIndexes[file.Name] = i
	}
}

func (lam *LodArchiveMeta) GetFile(name string) (LodFileMeta, error) {
	if lam.filesIndexes == nil {
		lam.indexFiles()
	}

	if int(lam.NumberOfFiles) != len(lam.filesIndexes) {
		return LodFileMeta{}, errors.New("corrupted index")
	}

	fi, ok := lam.filesIndexes[name]
	if !ok {
		return LodFileMeta{}, errors.New("no such file in archive")
	}

	if fi >= len(lam.Files) {
		return LodFileMeta{}, errors.New("corrupted index")
	}
	return lam.Files[fi], nil
}

func (lam *LodArchiveMeta) ToJSON() ([]byte, error) {
	lamJSON, err := json.Marshal(lam)
	if err != nil {
		return nil, err
	}
	return lamJSON, nil
}

func (lam *LodArchiveMeta) WriteJsonFile(jsonDstDir, filename string) error {
  f, err := os.Create(filepath.Join(jsonDstDir, filename))
  if err != nil {
	  return err
  }
  defer f.Close()
  err = json.NewEncoder(f).Encode(lam)
  if err != nil {
	  return err
  }
  return nil
}

type LodFileMeta struct {
	Name           string `json:"name"`
	Offset         int32  `json:"offset"`
	OriginalSize   int32  `json:"original_size"`
	CompressedSize int32  `json:"compressed_size"`
}

func (lf LodFileMeta) IsCompressed() bool {
	return lf.CompressedSize != 0
}

