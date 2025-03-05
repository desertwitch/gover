package io

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"

	"github.com/desertwitch/gover/internal/filesystem"
	"github.com/zeebo/blake3"
)

//nolint:containedctx
type contextReader struct {
	ctx    context.Context
	reader io.Reader
}

func (cr *contextReader) Read(p []byte) (int, error) {
	select {
	case <-cr.ctx.Done():
		return 0, context.Canceled
	default:
		return cr.reader.Read(p)
	}
}

func (i *Handler) moveFile(ctx context.Context, m *filesystem.Moveable) error {
	var transferComplete bool

	srcFile, err := i.OSOps.Open(m.SourcePath)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer srcFile.Close()

	tmpPath := m.DestPath + ".gover"
	defer func() {
		if !transferComplete {
			i.OSOps.Remove(tmpPath) //nolint:errcheck
		}
	}()

	dstFile, err := i.OSOps.OpenFile(tmpPath, os.O_CREATE|os.O_WRONLY|os.O_EXCL, os.FileMode(m.Metadata.Perms))
	if err != nil {
		return fmt.Errorf("failed to open destination file %s: %w", tmpPath, err)
	}
	defer dstFile.Close()

	srcHasher := blake3.New()
	dstHasher := blake3.New()

	ctxReader := &contextReader{
		ctx:    ctx,
		reader: io.TeeReader(srcFile, srcHasher),
	}
	multiWriter := io.MultiWriter(dstFile, dstHasher)

	if _, err := io.Copy(multiWriter, ctxReader); err != nil {
		if errors.Is(err, context.Canceled) {
			return fmt.Errorf("transfer canceled: %w", err)
		}

		return fmt.Errorf("failed to copy file: %w", err)
	}

	if err := dstFile.Sync(); err != nil {
		return fmt.Errorf("failed to sync destination fs: %w", err)
	}

	srcChecksum := hex.EncodeToString(srcHasher.Sum(nil))
	dstChecksum := hex.EncodeToString(dstHasher.Sum(nil))

	if srcChecksum != dstChecksum {
		return fmt.Errorf("%w: %s (src) != %s (dst)", ErrHashMismatch, srcChecksum, dstChecksum)
	}

	if _, err := i.OSOps.Stat(m.DestPath); err == nil {
		return ErrRenameExists
	} else if !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("failed to check rename destination existence: %w", err)
	}

	if err := i.OSOps.Rename(tmpPath, m.DestPath); err != nil {
		return fmt.Errorf("failed to rename temporary file to destination file: %w", err)
	}

	transferComplete = true

	return nil
}
