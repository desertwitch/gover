package io

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"time"

	"github.com/desertwitch/gover/internal/generic/filesystem"
	"github.com/desertwitch/gover/internal/generic/queue"
	"github.com/zeebo/blake3"
)

//nolint:containedctx
type contextReader struct {
	ctx        context.Context
	reader     io.Reader
	stats      *queue.TransferInfo
	bytesRead  uint64
	lastUpdate time.Time
}

func (cr *contextReader) Read(p []byte) (int, error) {
	select {
	case <-cr.ctx.Done():
		return 0, context.Canceled
	default:
		n, err := cr.reader.Read(p)

		if n > 0 {
			cr.bytesRead += uint64(n)

			now := time.Now()
			if now.Sub(cr.lastUpdate) >= 500*time.Millisecond {
				cr.stats.Update(cr.bytesRead)
				cr.lastUpdate = now
			}
		}

		return n, err
	}
}

func (i *Handler) moveFile(ctx context.Context, m *filesystem.Moveable, t *queue.TransferInfo) error {
	var transferComplete bool

	srcFile, err := i.osHandler.Open(m.SourcePath)
	if err != nil {
		return fmt.Errorf("(io-movefile) failed to open src: %w", err)
	}
	defer srcFile.Close()

	tmpPath := m.DestPath + ".gover"
	defer func() {
		if !transferComplete {
			i.osHandler.Remove(tmpPath) //nolint:errcheck
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
		ctx:        ctx,
		reader:     io.TeeReader(srcFile, srcHasher),
		stats:      t,
		lastUpdate: time.Now(),
	}

	multiWriter := io.MultiWriter(dstFile, dstHasher)

	t.Start(m.Metadata.Size)

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

	t.End()
	transferComplete = true

	return nil
}
