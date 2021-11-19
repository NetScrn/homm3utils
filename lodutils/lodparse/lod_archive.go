package lodparse

import (
	"encoding/json"
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

type LodArchiveMeta struct {
	ArchiveFilePath string         `json:"file_path"`
	LodType         LodArchiveType `json:"lod_type"`
	NumberOfFiles   int32          `json:"number_of_files"`
	Files           []LodFile      `json:"files"`
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

type LodFile struct {
	Name           string `json:"name"`
	Offset         int32  `json:"offset"`
	OriginalSize   int32  `json:"original_size"`
	CompressedSize int32  `json:"compressed_size"`
}

func (lf LodFile) IsCompressed() bool {
	return lf.CompressedSize != 0
}

