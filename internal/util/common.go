package util

import (
	"encoding/binary"
	"fmt"
)

type Header struct {
	fileName  string
	fileSize  uint64
	reps      uint64
	chunkSize uint32
	dir       string
}

// Serialize converts the Header struct to a byte slice
func (h *Header) Serialize() []byte {
	headerBuf := []byte{1}

	fl := uint32(len(h.fileName))
	temp := make([]byte, 4)
	binary.BigEndian.PutUint32(temp, fl)
	headerBuf = append(headerBuf, temp...)
	headerBuf = append(headerBuf, []byte(h.fileName)...)

	dl := uint32(len(h.dir))
	binary.BigEndian.PutUint32(temp, dl)
	headerBuf = append(headerBuf, temp...)
	headerBuf = append(headerBuf, []byte(h.dir)...)

	temp = make([]byte, 8)
	binary.BigEndian.PutUint64(temp, h.fileSize)
	headerBuf = append(headerBuf, temp...)

	binary.BigEndian.PutUint64(temp, h.reps)
	headerBuf = append(headerBuf, temp...)

	temp = make([]byte, 4)
	binary.BigEndian.PutUint32(temp, h.chunkSize)
	headerBuf = append(headerBuf, temp...)

	headerBuf = append(headerBuf, 0)

	return headerBuf
}

// Deserialize converts a byte slice back into the Header struct
func (h *Header) Deserialize(buf []byte) error {
	if len(buf) < 1 {
		return fmt.Errorf("buffer too small to contain header")
	}

	size := uint32(0)
	if buf[size] != byte(1) {
		return fmt.Errorf("not a header package, expected 1 as first byte, received %v", buf[size])
	}
	size += 1
	if size+4 > uint32(len(buf)) {
		return fmt.Errorf("buffer too small to contain file name length")
	}
	fl := binary.BigEndian.Uint32(buf[size : size+4])
	size += 4

	if size+fl > uint32(len(buf)) {
		return fmt.Errorf("buffer too small to contain file name")
	}
	h.fileName = string(buf[size : size+fl])
	size += fl

	if size+4 > uint32(len(buf)) {
		return fmt.Errorf("buffer too small to contain directory length")
	}

	dl := binary.BigEndian.Uint32(buf[size : size+4])
	size += 4
	if size+dl > uint32(len(buf)) {
		return fmt.Errorf("buffer too small to contain directory")
	}
	h.dir = string(buf[size : size+dl])
	size += dl

	if size+8 > uint32(len(buf)) {
		return fmt.Errorf("buffer too small to contain file size")
	}
	h.fileSize = binary.BigEndian.Uint64(buf[size : size+8])
	size += 8

	if size+8 > uint32(len(buf)) {
		return fmt.Errorf("buffer too small to contain repetitions")
	}
	h.reps = binary.BigEndian.Uint64(buf[size : size+8])
	size += 8

	if size+4 > uint32(len(buf)) {
		return fmt.Errorf("buffer too small to contain chunk size")
	}
	h.chunkSize = binary.BigEndian.Uint32(buf[size : size+4])
	size += 4

	if size >= uint32(len(buf)) || buf[size] != byte(0) {
		return fmt.Errorf("not a header package, expected 0 as last byte, received %v", buf[size])
	}

	return nil
}
