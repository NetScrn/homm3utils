package defparse

import (
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/netscrn/homm3utils/internal/binread"
)

type OutFilesMeta struct {
	BlocksMeta []DefBlockMeta `json:"blocks_meta"`
	DefType    uint32         `json:"def_type"`
	Format     uint32         `json:"format"`
}
type DefBlockMeta struct {
	Id        uint32     `json:"block_id"`
	DefImages []DefImage `json:"images"`
}
type DefImage struct {
	Name      string    `json:"name"`
	Offset    uint32    `json:"offset"`
}
type ImageMeta struct {
	Size       uint32
	Format     uint32
	FullWight  uint32
	FullHeight uint32
	Width      uint32
	Height     uint32
	LeftMargin int32
	TopMargin  int32
}

func ExtractDef(defPath, outDir string) error {
	defFile, err := os.Open(defPath)
	if err != nil {
		return fmt.Errorf("can't read def file: %w", err)
	}
	defer defFile.Close()

	defType, _, _, defBlocksCount, err := readDefMeta(defFile)
	if err != nil {
		return fmt.Errorf("can't read def header: %w", err)
	}

	palette, err := readDefPalette(defFile)
	if err != nil {
		return fmt.Errorf("can't read def palette: %w", err)
	}

	defBlocksMeta, err := readDefBlocksMeta(defFile, defBlocksCount)
	if err != nil {
		return fmt.Errorf("can't read def blocks meta: %w", err)
	}

	defName := filepath.Base(strings.TrimSuffix(defPath, filepath.Ext(defPath)))
	defOutDir := filepath.Join(outDir, defName)
	err = extractBlocksContent(defFile, defBlocksMeta, *palette, defOutDir, defType)
	if err != nil {
		return fmt.Errorf("can't extract def blocks content: %w", err)
	}

	return nil
}

func readDefMeta(defFile *os.File) (defType, width, height, blocks uint32, err error) {
	err = binread.ReadUint32(defFile, &defType)
	if err != nil {
		return 0, 0, 0, 0, fmt.Errorf("can't read def type: %w", err)
	}

	err = binread.ReadUint32(defFile, &width)
	if err != nil {
		return 0, 0, 0, 0, fmt.Errorf("can't read def width: %w", err)
	}

	err = binread.ReadUint32(defFile, &height)
	if err != nil {
		return 0, 0, 0, 0, fmt.Errorf("can't read def height: %w", err)
	}

	err = binread.ReadUint32(defFile, &blocks)
	if err != nil {
		return 0, 0, 0, 0, fmt.Errorf("can't read def blocks: %w", err)
	}

	return defType, width, height, blocks, nil
}

func readDefPalette(defFile *os.File) (*color.Palette, error) {
	palette := make(color.Palette, 256)
	for i := 0; i < 256; i++ {
		var r uint8
		err := binread.ReadUint8(defFile, &r)
		if err != nil {
			return nil, fmt.Errorf("can't read [%d][r]: %w", i+1, err)
		}
		var g uint8
		err = binread.ReadUint8(defFile, &g)
		if err != nil {
			return nil, fmt.Errorf("can't read [%d][g]: %w", i+1, err)
		}
		var b uint8
		err = binread.ReadUint8(defFile, &b)
		if err != nil {
			return nil, fmt.Errorf("can't read [%d][b]: %w", i+1, err)
		}

		palette[i] = color.RGBA{R: r, G: g, B: b, A: 255}
	}

	return &palette, nil
}

func readDefBlocksMeta(defFile *os.File, defBlocks uint32) (*[]DefBlockMeta, error) {
	_, err := defFile.Seek(784, io.SeekStart)
	if err != nil {
		return nil, fmt.Errorf("can't seek to def blocks: %w", err)
	}

	blocks := make([]DefBlockMeta, 0 ,defBlocks)
	for i := 0; i < int(defBlocks); i++ {
		var blockId uint32
		err := binread.ReadUint32(defFile, &blockId)
		if err != nil {
			return nil, fmt.Errorf("can't read block id: %w", err)
		}

		var defFilesCount uint32
		err = binread.ReadUint32(defFile, &defFilesCount)
		if err != nil {
			return nil, fmt.Errorf("can't read block entries: %w", err)
		}

		_, err = defFile.Seek(8, io.SeekCurrent)
		if err != nil {
			return nil, fmt.Errorf("can't skip unknown block meta data: %w", err)
		}

		defImagesMeta := make([]DefImage, 0, defFilesCount)
		for j := 0; j < int(defFilesCount); j++ {
			name, err := binread.ReadAvailableChars(defFile, 13)
			if err != nil {
				return nil, fmt.Errorf("can't read def file name(%d): %w", j+1, err)
			}
			defImagesMeta = append(defImagesMeta, DefImage{
				Name: name,
			})
		}
		for j := 0; j < int(defFilesCount); j++ {
			var offset uint32
			err := binread.ReadUint32(defFile, &offset)
			if err != nil {
				return nil, fmt.Errorf("can't read def file offset(%d): %w", j+1, err)
			}
			defImagesMeta[j].Offset = offset
		}

		blocks = append(blocks, DefBlockMeta{
			Id:        blockId,
			DefImages: defImagesMeta,
		})
	}

	return &blocks, nil
}

func extractBlocksContent(defFile *os.File, blocksMeta *[]DefBlockMeta, palette color.Palette, defOutDir string, defType uint32) error {
	err := resetDefOutDir(defOutDir)
	if err != nil {
		return err
	}

	ofm := OutFilesMeta{
		DefType: defType,
		BlocksMeta: *blocksMeta,
		Format: 99999,
	}

	for _, bm := range *blocksMeta {
		defBlockOutDir := filepath.Join(defOutDir, strconv.Itoa(int(bm.Id)))
		err := os.Mkdir(defBlockOutDir, 0700)
		if err != nil {
			return fmt.Errorf("can't create def dir(%s): %w", defBlockOutDir, err)
		}

		var firstFullWidth, firstFullHeight uint32 = 99999, 99999
		for _, di := range bm.DefImages {
			_, err := defFile.Seek(int64(di.Offset), io.SeekStart)
			if err != nil {
				return fmt.Errorf("can't seek to image(%s) offset(%d): %w", di.Name, di.Offset, err)
			}

			imgMeta, err := readImageMeta(defFile)
			if err != nil {
				return fmt.Errorf("can't read image(%s) meta: %w", di.Name, err)
			}

			// SGTWMTA.def and SGTWMTB.def fail here
			if imgMeta.LeftMargin > int32(imgMeta.FullWight) || imgMeta.TopMargin > int32(imgMeta.FullHeight) {
				errMsg := fmt.Sprintf(
					"margins(%dx%d) are higher than dimensions(%dx%d) in %s",
					imgMeta.LeftMargin, imgMeta.TopMargin, imgMeta.FullWight, imgMeta.FullHeight, di.Name,
				)
				return errors.New(errMsg)
			}

			if firstFullWidth == 99999 && firstFullHeight == 99999 {
				firstFullWidth  = imgMeta.FullWight
				firstFullHeight = imgMeta.FullHeight
			} else {
				if firstFullWidth > imgMeta.FullWight {
					imgMeta.FullWight = firstFullWidth // enlarge image width
				}
				if firstFullHeight > imgMeta.FullHeight {
					imgMeta.FullHeight = firstFullHeight // enlarge image height
				}
				if imgMeta.FullWight > firstFullWidth {
					return errors.New(fmt.Sprintf("%s width is greater than in first image", di.Name))
				}
				if imgMeta.FullHeight > firstFullHeight {
					return errors.New(fmt.Sprintf("%s height is greater than in first image", di.Name))
				}
			}

			if ofm.Format == 99999 {
				ofm.Format = imgMeta.Format
			} else if ofm.Format != imgMeta.Format {
				return errors.New(fmt.Sprintf("%s got different format than first image", di.Name))
			}

			var imgRGBA *image.RGBA
			if imgMeta.Width != 0 && imgMeta.Height != 0 {
				pixels, err := readPixels(defFile, di, imgMeta)
				if err != nil {
					return fmt.Errorf("cant read pixels of image %s: %w", di.Name, err)
				}
				imgRGBA = decodePixels(pixels, palette, imgMeta)
			} else {
				imgRGBA = image.NewRGBA(image.Rect(0, 0, 0, 0))
			}

			srcImgName := filepath.Base(strings.TrimSuffix(di.Name, filepath.Ext(di.Name)))
			imageDstPath := filepath.Join(defBlockOutDir, srcImgName + ".png")
			file, err := os.Create(imageDstPath)
			if err != nil {
				return fmt.Errorf("can't create png file(%s): %w", imageDstPath, err)
			}
			defer file.Close()

			err = png.Encode(file, imgRGBA)
			if err != nil {
				return fmt.Errorf("can't encode png(%s): %w", imageDstPath, err)
			}
		}
	}

	ofmFile, err := os.Create(filepath.Join(defOutDir, "meta.json"))
	if err != nil {
		fmt.Println(fmt.Sprint("can't create mata.json: %w", err))
	}
	jsonEncoder := json.NewEncoder(ofmFile)
	jsonEncoder.SetIndent("", "    ")
	err = jsonEncoder.Encode(ofm)
	if err != nil {
		fmt.Println(fmt.Sprint("can't write to mata.json: %w", err))
	}
	return  nil
}

func readImageMeta(defFile *os.File) (*ImageMeta, error) {
	var imageMeta ImageMeta

	err := binread.ReadUint32(defFile, &imageMeta.Size)
	if err != nil {
		return nil, fmt.Errorf("can't read Size: %w", err)
	}
	err = binread.ReadUint32(defFile, &imageMeta.Format)
	if err != nil {
		return nil, fmt.Errorf("can't read Format: %w", err)
	}
	err = binread.ReadUint32(defFile, &imageMeta.FullWight)
	if err != nil {
		return nil, fmt.Errorf("can't read FullWight: %w", err)
	}
	err = binread.ReadUint32(defFile, &imageMeta.FullHeight)
	if err != nil {
		return nil, fmt.Errorf("can't read FullHeight: %w", err)
	}
	err = binread.ReadUint32(defFile, &imageMeta.Width)
	if err != nil {
		return nil, fmt.Errorf("can't read Width: %w", err)
	}
	err = binread.ReadUint32(defFile, &imageMeta.Height)
	if err != nil {
		return nil, fmt.Errorf("can't read Height: %w", err)
	}
	err = binread.ReadInt32(defFile, &imageMeta.LeftMargin)
	if err != nil {
		return nil, fmt.Errorf("can't read LeftMargin: %w", err)
	}
	err = binread.ReadInt32(defFile, &imageMeta.TopMargin)
	if err != nil {
		return nil, fmt.Errorf("can't read TopMargin: %w", err)
	}

	return &imageMeta, nil
}

func resetDefOutDir(defOutDir string) error {
	if err := os.Mkdir(defOutDir, 0700); err != nil {
		if os.IsExist(err) {
			err := os.RemoveAll(defOutDir)
			if err != nil {
				return fmt.Errorf("can't remove def dir: %w", err)
			}
			err = os.Mkdir(defOutDir, 0700)
			if err != nil {
				return fmt.Errorf("can't create def dir: %w", err)
			}
		} else {
			return fmt.Errorf("can't create def dir: %w", err)
		}
	}
	return nil
}