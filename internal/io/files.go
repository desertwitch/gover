package io

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"

	"github.com/desertwitch/gover/internal/schema"
	"github.com/zeebo/blake3"
)

// contextReader is an implementation of [io.Reader] that is Context-aware for
// receiving mid-transfer cancellation.
type contextReader struct {
	ctx    context.Context //nolint:containedctx
	reader io.Reader
}

// Read wraps the [io.Reader] reading function while being aware of and handling
// any mid-transfer Context cancellations.
func (cr *contextReader) Read(p []byte) (int, error) {
	select {
	case <-cr.ctx.Done():
		return 0, context.Canceled
	default:
		return cr.reader.Read(p) //nolint:wrapcheck
	}
}

// moveFile is the principal method for moving a file-type [schema.Moveable].
func (i *Handler) moveFile(ctx context.Context, m *schema.Moveable) error {
	var transferComplete bool

	srcFile, err := i.osHandler.Open(m.SourcePath)
	if err != nil {
		return fmt.Errorf("(io-movefile) failed to open src: %w", err)
	}
	defer srcFile.Close()

	tmpPath := m.DestPath + ".gover"
	defer func() {
		if !transferComplete {
			_ = i.osHandler.Remove(tmpPath)
		}
	}()

	dstFile, err := i.osHandler.OpenFile(tmpPath, os.O_CREATE|os.O_WRONLY|os.O_EXCL, os.FileMode(m.Metadata.Perms))
	if err != nil {
		return fmt.Errorf("(io-movefile) failed to open dst: %w", err)
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
			return fmt.Errorf("(io-movefile) canceled: %w", err)
		}

		return fmt.Errorf("(io-movefile) failed to copy: %w", err)
	}

	if err := dstFile.Sync(); err != nil {
		return fmt.Errorf("(io-movefile) failed to sync dst: %w", err)
	}

	srcChecksum := hex.EncodeToString(srcHasher.Sum(nil))
	dstChecksum := hex.EncodeToString(dstHasher.Sum(nil))

	if srcChecksum != dstChecksum {
		return fmt.Errorf("(io-movefile) %w: %s (src) != %s (dst)", ErrHashMismatch, srcChecksum, dstChecksum)
	}

	if _, err := i.osHandler.Stat(m.DestPath); err == nil {
		return fmt.Errorf("(io-movefile) %w", ErrRenameExists)
	} else if !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("(io-movefile) failed to stat (pre rename existence): %w", err)
	}

	if err := i.osHandler.Rename(tmpPath, m.DestPath); err != nil {
		return fmt.Errorf("(io-movefile) failed to rename tmp file to dst file: %w", err)
	}

	transferComplete = true

	return nil
}
