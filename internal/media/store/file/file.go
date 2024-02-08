package file

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
)

const (
	defaultFilePermissions = 0600
	defaultDirPermissions  = 0700
)

// Store is media store that uses file system.
type Store struct {
	inputPathPrefix string
	outPathPrefix   string
}

func NewStore(inPath, outPath string) *Store {
	return &Store{
		inputPathPrefix: inPath,
		outPathPrefix:   outPath,
	}
}

func (s *Store) Get(_ context.Context, name string) (io.ReadCloser, int64, error) {
	fullName := fmt.Sprintf("%s/%s", s.inputPathPrefix, name)
	f, err := os.Open(fullName)
	if err != nil {
		return nil, 0, fmt.Errorf("cannot open file: %w", err)
	}
	stat, errS := f.Stat()
	if errS != nil {
		return nil, 0, fmt.Errorf("cannot get file stats: %w", errS)
	}
	return f, stat.Size(), nil
}

func (s *Store) Put(_ context.Context, name string, r io.Reader, size int64) error {
	fullName := fmt.Sprintf("%s/%s", s.outPathPrefix, name)
	if errD := os.MkdirAll(fullName[:strings.LastIndexByte(fullName, '/')], defaultDirPermissions); errD != nil {
		return fmt.Errorf("cannot create dir: %w", errD)
	}
	f, err := os.OpenFile(fullName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, defaultFilePermissions)
	if err != nil {
		return fmt.Errorf("cannot open file for writing: %w", err)
	}
	defer func() {
		_ = f.Close()
	}()
	n, errW := io.Copy(f, r) // Reader MUST send EOF otherwise there will be deadlock
	if errW != nil {
		return fmt.Errorf("error writing to file: %w", errW)
	}
	if n != size {
		return fmt.Errorf("wrote incorrect amount of bytes: expected: %d, actual: %d", size, n)
	}
	return nil
}
