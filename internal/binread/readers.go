package binread

import (
	"encoding/binary"
	"io"
)

func ReadUint8(r io.Reader, n *uint8) error {
	return binary.Read(r, binary.LittleEndian, n)
}

func ReadUint16(r io.Reader, n *uint16) error {
	return binary.Read(r, binary.LittleEndian, n)
}

func ReadInt16(r io.Reader, n *int16) error {
	return binary.Read(r, binary.LittleEndian, n)
}

func ReadUint32(r io.Reader, n *uint32) error {
	return binary.Read(r, binary.LittleEndian, n)
}

func ReadInt32(r io.Reader, n *int32) error {
	return binary.Read(r, binary.LittleEndian, n)
}

func ReadAvailableChars(r io.Reader, charsCount int) (string, error) {
	nameBuf := make([]byte, charsCount)
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