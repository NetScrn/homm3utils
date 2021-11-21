package lodparse_test

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/netscrn/homm3utils/lodutils/lodparse"
)

var tempDirPath string

func TestMain(m *testing.M) {
	tdp, err := os.MkdirTemp("", "lodparse_test")
	if err != nil {
		panic("can't create temp dir for tests")
	}
	tempDirPath = tdp
	m.Run()
	err = os.RemoveAll(tempDirPath)
}

func TestLoadLodArchiveMetaFromLodFile(t *testing.T) {
	lam, err := lodparse.LoadLodArchiveMetaFromLodFile(filepath.Join(".", "testdata", "HotA_lng.lod"))
	if err != nil {
		t.Fatalf("Can't load lod archive meta: %s", err.Error())
	}
	lamFile, err := lam.GetFile("advevent.txt")
	if err != nil {
		t.Fatalf("Can't get file in lod archive meta: %s", err.Error())
	}

	if !lam.LodType.IsUnknownType() {
		t.Error("Wrong lod type")
	}

	if int(lam.NumberOfFiles) != len(lam.Files) {
		t.Error("Archives number of filed do not match with extracted number of files meta")
	}

	if lam.NumberOfFiles != 209 {
		t.Error("Wrong number of files in lod archive")
	}

	if int(lamFile.Offset) != 320092 {
		t.Error("Wrong archive file offset")
	}

	if int(lamFile.OriginalSize) != 23895 {
		t.Error("Wrong archive file original size")
	}

	if int(lamFile.CompressedSize) != 8946 {
		t.Error("Wrong archive file original size")
	}
}

func TestExtractFile(t *testing.T) {
	lam, err := lodparse.LoadLodArchiveMetaFromLodFile(filepath.Join(".", "testdata", "HotA_lng.lod"))
	if err != nil {
		t.Fatalf("Can't load lod archive meta: %s", err.Error())
	}
	compressedFile, err := lam.GetFile("AVArnd1.def")
	if err != nil {
		t.Fatalf("Can't get compressed file meta from archive meta: %s", err.Error())
	}
	notCompressedFile, err := lam.GetFile("AVArnd1.msk")
	if err != nil {
		t.Fatalf("Can't get not compressed file meta from archive meta: %s", err.Error())
	}
	lodArchiveReader, err := os.Open(filepath.Join(".", "testdata", "HotA_lng.lod"))
	if err != nil {
		t.Fatalf("Can open lod archive: %s", err.Error())
	}
	defer lodArchiveReader.Close()


	err = lodparse.ExtractFile(compressedFile, lodArchiveReader, tempDirPath)
	if err != nil {
		t.Fatalf("Can't write extracted compressed file: %s", err.Error())
	}
	err = lodparse.ExtractFile(notCompressedFile, lodArchiveReader, tempDirPath)
	if err != nil {
		t.Fatalf("Can't write extracted not compressed file: %s", err.Error())
	}


	originalExtractedCompressedFile, err := os.ReadFile(filepath.Join(".", "testdata", "HotA_lng_files", "AVArnd1.def"))
	testingExtractedCompressedFile, err := os.ReadFile(filepath.Join(tempDirPath, "AVArnd1.def"))
	if !reflect.DeepEqual(originalExtractedCompressedFile, testingExtractedCompressedFile) {
		t.Error("invalid extracted compressed file")
	}

	originalExtractedNotCompressedFile, err := os.ReadFile(filepath.Join(".", "testdata", "HotA_lng_files", "AVArnd1.msk"))
	testingExtractedNotCompressedFile, err := os.ReadFile(filepath.Join(tempDirPath, "AVArnd1.msk"))
	if !reflect.DeepEqual(originalExtractedNotCompressedFile, testingExtractedNotCompressedFile) {
		t.Error("invalid extracted not compressed file")
	}
}