// Package file provides helpers that are used during video file upload.
package file

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/minio/sha256-simd"
)

type Part struct {
	Checksum string `json:"checksum"`
	Num      uint   `json:"num"`
	Size     uint   `json:"size"`
}

// MakePartsFromFile splits file to parts. It returns slice of parts with fixed size
// (but last part most probably will have less size) and calculated sha256 checksums.
//
// []Part actually just represent points in file, and later actual bytes are
// anyway read from source file using offsets.
//
// Main purpose of []Part is that it goes directly into Video object (which is sent to videoapi).
func MakePartsFromFile(filePath string, partSize, size uint64) ([]Part, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("unable to open file: %w", err)
	}
	defer func() { _ = f.Close() }()

	partCount := size / partSize
	if size%partSize != 0 {
		partCount++
	}
	var parts = make([]Part, 0, partCount)
	for i := uint64(0); i < partCount; i++ {
		h := sha256.New()
		n, errCp := io.CopyN(h, f, int64(partSize))
		if errCp != nil {
			if !errors.Is(errCp, io.EOF) || i != partCount-1 {
				return nil, fmt.Errorf("unable to calculate sha256 sum: %w", errCp)
			}
		}
		parts = append(parts, Part{
			Num:      uint(i),
			Size:     uint(n),
			Checksum: base64.StdEncoding.EncodeToString(h.Sum(nil)),
		})
	}
	return parts, nil
}

// GetSize opens file and returns its size.
func GetSize(filePath string) (uint64, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return 0, fmt.Errorf("unable to open file: %w", err)
	}
	defer func() { _ = f.Close() }()
	stat, err := f.Stat()
	if err != nil {
		return 0, fmt.Errorf("unable to get file stats: %w", err)
	}
	return uint64(stat.Size()), nil
}
