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

type Store struct {
	pathPrefix string
}

func NewStore(prefix string) *Store {
	return &Store{
		pathPrefix: prefix,
	}
}

func (s *Store) Put(_ context.Context, name string, r io.Reader, size int64) error {
	fullName := s.getFullFileName(name)
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

func (s *Store) getFullFileName(name string) string {
	return fmt.Sprintf("%s/%s", s.pathPrefix, name)
}
