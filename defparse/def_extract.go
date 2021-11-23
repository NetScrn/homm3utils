package defparse

import (
	"errors"
	"fmt"
	"github.com/netscrn/homm3utils/internal/binread"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type DefBlockMeta struct {
	Id            uint32
	DefImagesMeta []DefImageMeta
}
type DefImageMeta struct {
	Name   string
	Offset uint32
}
type OutFilesMeta struct {
	BlocksMeta []DefBlockMeta
	DefType    uint32
	Format     uint32
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

	fmt.Println("Done")
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

		defImagesMeta := make([]DefImageMeta, 0, defFilesCount)
		for j := 0; j < int(defFilesCount); j++ {
			name, err := binread.ReadAvailableChars(defFile, 13)
			if err != nil {
				return nil, fmt.Errorf("can't read def file name(%d): %w", j+1, err)
			}
			defImagesMeta = append(defImagesMeta, DefImageMeta{
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
			Id: blockId,
			DefImagesMeta: defImagesMeta,
		})
	}

	return &blocks, nil
}

func extractBlocksContent(defFile *os.File, blocksMeta *[]DefBlockMeta, palette color.Palette, defOutDir string, defType uint32) error {
	err := os.Mkdir(defOutDir, 0700)
	if err != nil {
		return fmt.Errorf("can't create def dir(%s): %w", defOutDir, err)
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
		for _, dim := range bm.DefImagesMeta {
			_, err := defFile.Seek(int64(dim.Offset), io.SeekStart)
			if err != nil {
				return fmt.Errorf("can't seek to image(%s) offset(%d): %w", dim.Name, dim.Offset, err)
			}

			im, err := readImageMeta(defFile)
			if err != nil {
				return fmt.Errorf("can't read image(%s) meta: %w", dim.Name, err)
			}

			// SGTWMTA.def and SGTWMTB.def fail here
			if im.LeftMargin > int32(im.FullWight) || im.TopMargin > int32(im.FullHeight) {
				errMsg := fmt.Sprintf(
					"margins(%dx%d) are higher than dimensions(%dx%d) in %s",
					im.LeftMargin, im.TopMargin, im.FullWight, im.FullHeight, dim.Name,
				)
				return errors.New(errMsg)
			}

			if firstFullWidth == 99999 && firstFullHeight == 99999 {
				firstFullWidth  = im.FullWight
				firstFullHeight = im.FullHeight
			} else {
				if firstFullWidth > im.FullWight {
					im.FullWight = firstFullWidth   // enlarge image width
				}
				if firstFullHeight > im.FullHeight {
					im.FullHeight = firstFullHeight // enlarge image height
				}
				if im.FullWight > firstFullWidth {
					return errors.New(fmt.Sprintf("%s width is greater than in first image", dim.Name))
				}
				if im.FullHeight > firstFullHeight {
					return errors.New(fmt.Sprintf("%s height is greater than in first image", dim.Name))
				}
			}

			if ofm.Format == 99999 {
				ofm.Format = im.Format
			} else if ofm.Format != im.Format {
				return errors.New(fmt.Sprintf("%s got different format than first image", dim.Name))
			}

			if im.Width != 0 && im.Height != 0 {
				if im.Format == 0 {
					pixels := make([]uint8, im.Width*im.Height)
					_, err := defFile.Read(pixels)
					if err != nil {
						return errors.New(fmt.Sprintf("can't read image(%s) pixels", dim.Name))
					}

					img := image.NewPaletted(image.Rect(0, 0, int(im.Width), int(im.Height)), palette)
					img.Pix = pixels
				} else if im.Format == 1 {
					var pixels []uint8

					lineoffs := make([]uint32, im.Height)
					for i := 0; i < int(im.Height); i++ {
						err = binread.ReadUint32(defFile, &lineoffs[i])
						if err != nil {
							return errors.New(fmt.Sprintf("can't read image(%s) lineoffset number(%d)", dim.Name, i))
						}
					}

					for _, lineoff := range lineoffs {
						_, err = defFile.Seek(int64(dim.Offset)+32+int64(lineoff), io.SeekStart)
						if err != nil {
							return errors.New(fmt.Sprintf("can't seek image(%s) lineoffset(%d)", dim.Name, lineoff))
						}

						var totalRowLength uint32
						for ;totalRowLength < im.Width; {

							var code uint8
							err = binread.ReadUint8(defFile, &code)
							if err != nil {
								return fmt.Errorf("cant read row code: %w", err)
							}

							var length uint8
							err = binread.ReadUint8(defFile, &length)
							if err != nil {
								return fmt.Errorf("cant read row length: %w", err)
							}

							length++

							if code == 0xff {
								for i := 0; i < int(length); i++ {
									var b uint8
									err = binread.ReadUint8(defFile, &b)
									if err != nil {
										return fmt.Errorf("cant read row code: %w", err)
									}
									pixels = append(pixels, b)
								}
							} else {
								for i := 0; i < int(length); i++ {
									pixels = append(pixels, code)
								}
							}
							totalRowLength += uint32(length)
						}
					}

					img := image.NewPaletted(image.Rect(0, 0, int(im.Width), int(im.Height)), palette)
					img.Pix = pixels


					imgrgba := image.NewRGBA(img.Rect)
					draw.Draw(imgrgba, imgrgba.Rect, img, imgrgba.Rect.Min, draw.Src)

					imageDstPath := filepath.Join(defBlockOutDir, filepath.Base(strings.TrimSuffix(dim.Name, filepath.Ext(dim.Name))) + ".png")
					file, err := os.Create(imageDstPath)
					if err != nil {
						return fmt.Errorf("can't create png file(%s): %w", imageDstPath, err)
					}
					defer file.Close()

					png.Encode(file, imgrgba)
				} else if im.Format == 2 {
					log.Println("FIRED 2")
				} else if im.Format == 3 {
					log.Println("FIRED 3")
				}
			}
		}
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