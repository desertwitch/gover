package filesystem

import (
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"path/filepath"
	"sort"
	"strings"
)

type RelatedDirectory struct {
	SourcePath string
	DestPath   string
	Metadata   *Metadata
	Parent     *RelatedDirectory
	Child      *RelatedDirectory
}

func (d *RelatedDirectory) GetMetadata() *Metadata {
	return d.Metadata
}

func (d *RelatedDirectory) GetSourcePath() string {
	return d.SourcePath
}

func (d *RelatedDirectory) GetDestPath() string {
	return d.DestPath
}

func (f *FileHandler) walkParentDirs(m *Moveable, basePath string) error {
	var prevElement *RelatedDirectory
	path := m.SourcePath

	for path != basePath && path != "/" && path != "." {
		path = filepath.Dir(path)

		if strings.HasPrefix(path, basePath) {
			thisElement := &RelatedDirectory{
				SourcePath: path,
			}

			metadata, err := f.getMetadata(path)
			if err != nil {
				return fmt.Errorf("failed to get metadata: %w", err)
			}
			thisElement.Metadata = metadata

			if prevElement != nil {
				thisElement.Child = prevElement
				prevElement.Parent = thisElement
			} else {
				m.DeepestDir = thisElement
			}

			prevElement = thisElement
		} else {
			break
		}
	}
	m.RootDir = prevElement

	return nil
}

func (f *FileHandler) ensureDirectoryStructure(m *Moveable, job *InternalProgressReport) error {
	dir := m.RootDir

	for dir != nil {
		// TO-DO: Handle generic errors here and otherwere for .Stat or .Lstat
		if _, err := f.OSOps.Stat(dir.DestPath); errors.Is(err, fs.ErrNotExist) {
			if err := f.UnixOps.Mkdir(dir.DestPath, dir.Metadata.Perms); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", dir.DestPath, err)
			}

			if err := f.ensurePermissions(dir.DestPath, dir.Metadata); err != nil {
				return fmt.Errorf("failed to ensure permissions: %w", err)
			}

			job.AnyProcessed = append(job.AnyProcessed, dir)
			job.DirsProcessed = append(job.DirsProcessed, dir)
		}
		dir = dir.Child
	}

	return nil
}

func (f *FileHandler) removeEmptyDirs(batch *InternalProgressReport) error {
	sort.Slice(batch.DirsProcessed, func(i, j int) bool {
		return calculateDirectoryDepth(batch.DirsProcessed[i]) > calculateDirectoryDepth(batch.DirsProcessed[j])
	})

	removed := make(map[string]struct{})

	for _, dir := range batch.DirsProcessed {
		if _, alreadyRemoved := removed[dir.SourcePath]; alreadyRemoved {
			continue
		}
		isEmpty, err := f.IsEmptyFolder(dir.SourcePath)
		if err != nil {
			slog.Warn("Warning (cleanup): failure establishing source directory emptiness (skipped)", "path", dir.SourcePath, "err", err)
			continue
		}
		if isEmpty {
			if err := f.OSOps.Remove(dir.SourcePath); err != nil {
				slog.Warn("Warning (cleanup): failure removing empty source directory (skipped)", "path", dir.SourcePath, "err", err)
				continue
			}
			removed[dir.SourcePath] = struct{}{}
		}
	}

	return nil
}

func calculateDirectoryDepth(dir *RelatedDirectory) int {
	depth := 0
	for dir != nil {
		dir = dir.Parent
		depth++
	}
	return depth
}
