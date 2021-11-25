package defparse

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/netscrn/homm3utils/internal/binread"
)

func readPixels(defFile *os.File, di DefImage, imgMeta *ImageMeta) ([]uint8, error) {
	switch imgMeta.Format {
	case 0:
		return readFormat0Pixels(defFile, di, imgMeta)
	case 1:
		return readFormat1Pixels(defFile, di, imgMeta)
	case 2:
		return readFormat2Pixels(defFile, di, imgMeta)
	case 3:
		return readFormat3Pixels(defFile, di, imgMeta)
	default:
		return nil, errors.New("unknown format")
	}
}

func readFormat0Pixels(defFile *os.File, di DefImage, imgMeta *ImageMeta) ([]uint8, error) {
	pixels := make([]uint8, imgMeta.Width*imgMeta.Height)
	_, err := defFile.Read(pixels)
	if err != nil {
		return nil, fmt.Errorf("can't read image(%s) format0 pixels: %w", di.Name, err)
	}
	return pixels, nil
}

func readFormat1Pixels(defFile *os.File, di DefImage, imgMeta *ImageMeta) ([]uint8, error) {
	pixels := make([]uint8, 0, imgMeta.Width * imgMeta.Height)

	lineOffs := make([]uint32, imgMeta.Height)
	for i := 0; i < int(imgMeta.Height); i++ {
		err := binread.ReadUint32(defFile, &lineOffs[i])
		if err != nil {
			return nil, errors.New(fmt.Sprintf("can't read image(%s) lineoffset number(%d)", di.Name, i))
		}
	}
	for _, lineOff := range lineOffs {
		_, err := defFile.Seek(int64(di.Offset)+32+int64(lineOff), io.SeekStart)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("can't seek image(%s) lineoffset(%d)", di.Name, lineOff))
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

	return pixels, nil
}

func readFormat2Pixels(defFile *os.File, di DefImage, imgMeta *ImageMeta) ([]uint8, error) {
	pixels := make([]uint8, 0, imgMeta.Width * imgMeta.Height)

	lineOffs := make([]int16, imgMeta.Height)
	for x := 0; x < int(imgMeta.Height); x++ {
		err := binread.ReadInt16(defFile, &lineOffs[x])
		if err != nil {
			return nil, err
		}
	}

	for _, lineOff := range lineOffs {
		_, err := defFile.Seek(int64(int(di.Offset)+32+int(lineOff)), io.SeekStart)
		if err != nil {
			return nil, err
		}

		var totalBlockLength uint32
		for ;totalBlockLength < imgMeta.Width; {
			var segment uint8
			err := binread.ReadUint8(defFile, &segment)
			if err != nil {
				return nil, err
			}
			code := segment >> 5
			length := (segment&0x1f)+1

			if code == 7 { // plain bytes
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
			totalBlockLength += uint32(length)
		}
	}

	return pixels, nil
}

func readFormat3Pixels(defFile *os.File, di DefImage, imgMeta *ImageMeta) ([]uint8, error) {
	pixels := make([]uint8, 0, imgMeta.Width * imgMeta.Height)

	lineOffs := make([][]uint16, imgMeta.Height)
	for x := 0; x < int(imgMeta.Height); x++ {
		for y := 0; y < int(imgMeta.Width / 32); y++ {
			var bb uint16
			err := binread.ReadUint16(defFile, &bb)
			if err != nil {
				return nil, err
			}
			lineOffs[x] = append(lineOffs[x], bb)
		}
	}
	for _, lineOff := range lineOffs {
		for _, i := range lineOff {

			_, err := defFile.Seek(int64(int(di.Offset)+32+int(i)), io.SeekStart)
			if err != nil {
				return nil, err
			}
			var totalBlockLength uint32
			for ;totalBlockLength < 32; {
				var segment uint8
				err := binread.ReadUint8(defFile, &segment)
				if err != nil {
					return nil, err
				}
				code := segment >> 5
				length := (segment&0x1f)+1

				if code == 7 { // plain bytes
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
				totalBlockLength += uint32(length)
			}
		}
	}
	return pixels, nil
}