package defparse

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/netscrn/homm3utils/internal/binread"
)

func readFormat1Pixels(defFile *os.File, dim DefImage, imgMeta *ImageMeta) (*[]uint8, error) {
	var pixels []uint8

	lineOffs := make([]uint32, imgMeta.Height)
	for i := 0; i < int(imgMeta.Height); i++ {
		err := binread.ReadUint32(defFile, &lineOffs[i])
		if err != nil {
			return nil, errors.New(fmt.Sprintf("can't read image(%s) lineoffset number(%d)", dim.Name, i))
		}
	}
	for _, lineOff := range lineOffs {
		_, err := defFile.Seek(int64(dim.Offset)+32+int64(lineOff), io.SeekStart)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("can't seek image(%s) lineoffset(%d)", dim.Name, lineOff))
		}

		var totalRowLength uint32
		for ;totalRowLength < imgMeta.Width; {
			var code uint8
			err = binread.ReadUint8(defFile, &code)
			if err != nil {
				return nil, fmt.Errorf("cant read row code: %w", err)
			}
			var length uint8
			err = binread.ReadUint8(defFile, &length)
			if err != nil {
				return nil, fmt.Errorf("cant read row length: %w", err)
			}
			length++

			if code == 0xff { // plain bytes
				b := make([]byte, length)
				_, err = defFile.Read(b)
				if err != nil {
					return nil, fmt.Errorf("cant read row code: %w", err)
				}
				pixels = append(pixels, b...)
			} else { // RLE
				for i := 0; i < int(length); i++ {
					pixels = append(pixels, code)
				}
			}
			totalRowLength += uint32(length)
		}
	}

	return &pixels, nil
}