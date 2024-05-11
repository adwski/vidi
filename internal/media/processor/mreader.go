package processor

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"go.uber.org/zap"
)

const (
	maxOpenReaders = 10
)

// MediaReader is an uploaded media file parts reader.
// It abstracts multiple io.ReadSeekClosers as one,
// so it can be used together with lazy read mode in mp4ff.
//
// It assumes that each underlying reader represents continuous
// piece of data in order that is indicated by reader number.
//
// It knows full size of object it abstracts and keeps track of
// read/seek position.
//
// During each Read/Seek call it calculates reader number by knowing
// current position, amount of data to read and total size of object
// and routes Read/Seek call to reader with a particular number.
//
// If desired amount of data exceeds current reader's boundary
// MediaReader switches to next reader in order until data is fully read.
//
// To save memory and reduce number of open connections MediaReader
// only keeps last 'maxOpenReaders' readers open. If amount of readers
// is more than maxOpenReaders, several least recently used readers
// are closed in order to satisfy condition. Closed reader can be
// reopened just the same as if it was never used before.
//
// If error is occurred in any of the stages, Read/Seek will always
// return last error, and MediaReader must be closed.
//
// Current implementation is not thread-safe and should be used by only one goroutine.
type MediaReader struct {
	traceLogger *zap.Logger
	store       MediaStore
	readers     map[uint64]io.ReadSeekCloser
	readersTS   map[uint64]int64
	err         error
	s3ath       string
	parts       uint
	partSize    uint64
	totalSize   uint64
	pos         uint64
}

func NewMediaReader(ms MediaStore, path string, parts uint, totalSize, partSize uint64) *MediaReader {
	return &MediaReader{
		store:     ms,
		parts:     parts,
		s3ath:     path,
		totalSize: totalSize,
		partSize:  partSize,
		readers:   make(map[uint64]io.ReadSeekCloser),
		readersTS: make(map[uint64]int64),
	}
}

func (mr *MediaReader) Read(b []byte) (int, error) {
	if mr.traceLogger != nil {
		mr.traceLogger.Debug("read call", zap.Int("len", len(b)))
	}
	if mr.err != nil {
		// error occurred previously. At this point we cannot guarantee successful reads
		// TODO: As possible future improvements we can try to reopen connections on error
		return 0, fmt.Errorf("previous error: %w", mr.err)
	}
	var (
		partNum      = mr.pos / mr.partSize
		mustReadSize = uint64(len(b))
		dstStart     uint64
		start        = mr.pos
	)

	// iterate over parts until we've read all we need
	for {
		if mr.traceLogger != nil {
			mr.traceLogger.Debug("reading at", zap.Uint64("pos", mr.pos))
		}
		if partNum >= uint64(mr.parts) {
			return int(mr.pos - start), fmt.Errorf("requested partNum out of bounds: %d, parts: %d", partNum, mr.parts)
		}
		if err := mr.ensureReaderIsOpen(partNum); err != nil {
			mr.err = err
			return int(mr.pos - start), err
		}

		partBorder := min(mr.partSize*(partNum+1), mr.totalSize) // totalSize in case of last part
		if mr.traceLogger != nil {
			mr.traceLogger.Debug("border",
				zap.Uint64("partBorder", partBorder),
				zap.Uint64("mustReadSize", mustReadSize))
		}
		if partBorder >= mr.pos+mustReadSize {
			// staying withing same part
			if dstStart == mustReadSize {
				if err := mr.recycleReader(partNum); err != nil {
					return int(mr.pos - start), err
				}
				return int(mr.pos - start), io.EOF
			}
			n, err := mr.readers[partNum].Read(b[dstStart : dstStart+mustReadSize])
			if err != nil {
				if !errors.Is(err, io.EOF) {
					err = fmt.Errorf("error reading bytes from pos %d [%d:%d] within same part %d: %w",
						mr.pos, dstStart, dstStart+mustReadSize, partNum, err)
					mr.err = err
				}
			}
			mr.pos += uint64(n)
			return int(mr.pos - start), err
		} else if mr.pos == partBorder && mr.totalSize == partBorder {
			if err := mr.recycleReader(partNum); err != nil {
				return int(mr.pos - start), err
			}
			return int(mr.pos - start), io.EOF
		}

		// Will have to switch readers
		// First, read up to the border of current reader
		amount := partBorder - mr.pos
		n, err := mr.readers[partNum].Read(b[dstStart:amount])
		if err != nil {
			if !(errors.Is(err, io.EOF) && uint64(n) == amount-dstStart) {
				err = fmt.Errorf("error reading full bytes from %d [%d:%d] from part %d: %w",
					mr.pos, dstStart, amount, partNum, err)
				mr.err = err
				mr.pos += uint64(n)
				return int(mr.pos - start), err
			}
			// in case of success, we'll always end up here because we're reading part till the end
		} else {
			// err should not be nil, classic case of "error: no error"
			mr.pos += uint64(n)
			err = fmt.Errorf("did not reach EOF while fully reading part: %d, read bytes: %d", partNum, n)
			return int(mr.pos - start), err
		}
		// advance input buffer position
		dstStart += amount
		// advance internal read position
		mr.pos += uint64(n)

		// Switch to next part
		partNum++
		// decrease amount of must-read bytes by what we've already red
		mustReadSize -= amount
	}
}

func (mr *MediaReader) Seek(offset int64, whence int) (int64, error) {
	if mr.traceLogger != nil {
		mr.traceLogger.Debug("seek call",
			zap.Uint64("pos", mr.pos),
			zap.Int("whence", whence),
			zap.Int64("offset", offset))
	}
	if mr.err != nil {
		return 0, fmt.Errorf("previous error: %w", mr.err)
	}
	var (
		partNum uint64
		pOffset int64
	)
	switch whence {
	case io.SeekStart:
		if offset < 0 {
			// offset should be positive for io.SeekStart
			return 0, errors.New("invalid offset for whence 0")
		}
		partNum = uint64(offset) / mr.partSize
		pOffset = offset % int64(mr.partSize)
	case io.SeekCurrent:
		if offset < 0 {
			// Offset should be positive for io.SeekCurrent.
			// In general case I think negative offset with SeekCurrent makes sense
			// (as long as it stays within range), but s3 client doesn't allow this
			// and with response body readers it might as well be impossible.
			// So if we really want to seek backwards, we may cache responses
			// or reopen reader and convert to SeekStart.
			// Anyway, this is too advance use cases for now.
			return 0, errors.New("invalid offset for whence 1")
		}
		partNum = (mr.pos + uint64(offset)) / mr.partSize
		pOffset = int64((mr.pos + uint64(offset)) % mr.partSize)
	case io.SeekEnd:
		if offset > 0 {
			// offset should be negative for io.SeekEnd
			return 0, errors.New("invalid offset for whence 2")
		}
		offsetFromStart := int64(mr.totalSize) + offset
		partNum = uint64(offsetFromStart) / mr.partSize
		pOffset = offsetFromStart % int64(mr.partSize)
	default:
		return 0, errors.New("invalid whence")
	}
	if err := mr.ensureReaderIsOpen(partNum); err != nil {
		mr.err = err
		return 0, err
	}
	// seek from the start with calculated offset
	n, err := mr.readers[partNum].Seek(pOffset, io.SeekStart)
	if err != nil {
		err = fmt.Errorf("cannot seek to %d(%d) with whence %d for part %d: %w", offset, pOffset, whence, partNum, err)
		mr.err = err
		return n, err
	}
	if whence == io.SeekStart {
		mr.pos = uint64(offset)
	} else {
		mr.pos += uint64(offset)
	}
	return int64(mr.pos), nil
}

func (mr *MediaReader) Close() error {
	var err error
	for _, rc := range mr.readers {
		if rErr := rc.Close(); rErr != nil {
			err = errors.Join(err, rErr)
		}
	}
	mr.readers = nil
	mr.readersTS = nil
	return err
}

func (mr *MediaReader) ensureReaderIsOpen(partNum uint64) error {
	if _, ok := mr.readers[partNum]; !ok {
		// spawn reader if it not exists
		var artifactName = fmt.Sprintf("%s/%d", mr.s3ath, partNum)
		rc, _, err := mr.store.Get(context.TODO(), artifactName)
		if err != nil {
			return fmt.Errorf("failed to get s3 reader object %s: %w", artifactName, err)
		}
		mr.readers[partNum] = rc
	}
	mr.readersTS[partNum] = time.Now().UnixMilli() // update time mark for readers gc
	mr.gc()                                        // recycle least recently used readers
	return nil
}

func (mr *MediaReader) recycleReader(partNum uint64) error {
	err := mr.readers[partNum].Close()
	delete(mr.readers, partNum)
	delete(mr.readersTS, partNum)
	if err != nil {
		return fmt.Errorf("cannot close reader %d: %w", partNum, err)
	}
	return nil
}

func (mr *MediaReader) gc() {
	for len(mr.readers) > maxOpenReaders {
		// sweep
		var (
			minTS  = time.Now().UnixMilli()
			minNum uint64
		)
		for num, ts := range mr.readersTS {
			if ts < minTS {
				minTS = ts
				minNum = num
			}
		}
		err := mr.readers[minNum].Close()
		if err != nil {
			mr.err = fmt.Errorf("failed to close reader %d during gc: %w", minNum, err)
			return
		}
		delete(mr.readers, minNum)
		delete(mr.readersTS, minNum)
	}
}
