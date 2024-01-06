package file

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"go.uber.org/zap"
)

const (
	defaultFilePermissions = 0600
	defaultDirPermissions  = 0700
)

type Store struct {
	logger     *zap.Logger
	files      map[string]*os.File
	pathPrefix string
}

func NewStore(logger *zap.Logger, prefix string) *Store {
	return &Store{
		logger:     logger,
		pathPrefix: prefix,
		files:      make(map[string]*os.File),
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

	s.logger.Debug("file stored",
		zap.String("path", fullName),
		zap.Int64("size", size))
	return nil
}

func (s *Store) getFullFileName(name string) string {
	return fmt.Sprintf("%s/%s", s.pathPrefix, name)
}
