package validation

import (
	"path/filepath"
	"testing"

	"github.com/desertwitch/gover/internal/schema"
	"github.com/stretchr/testify/assert"
)

func TestValidateDirectories_Valid(t *testing.T) {
	src := &fakeStorage{"source", "/mnt/source"}
	dst := &fakeStorage{"dest", "/mnt/dest"}
	share := &fakeShare{"share"}

	dir2 := &schema.Directory{
		SourcePath: "/mnt/source/share/dir2",
		DestPath:   "/mnt/dest/share/dir2",
		Metadata:   &schema.Metadata{IsDir: true},
	}
	dir1 := &schema.Directory{
		SourcePath: "/mnt/source/share",
		DestPath:   "/mnt/dest/share",
		Metadata:   &schema.Metadata{IsDir: true},
		Child:      dir2,
	}
	dir2.Parent = dir1

	moveable := &schema.Moveable{
		RootDir:    dir1,
		Source:     src,
		Dest:       dst,
		Share:      share,
		SourcePath: "/mnt/source/share",
		DestPath:   "/mnt/dest/share",
	}

	err := validateDirectories(moveable)
	assert.NoError(t, err)
}

func TestValidateDirectories_Valid_NilRoot(t *testing.T) {
	src := &fakeStorage{"source", "/mnt/source"}
	dst := &fakeStorage{"dest", "/mnt/dest"}
	share := &fakeShare{"share"}

	moveable := &schema.Moveable{
		RootDir:    nil,
		Source:     src,
		Dest:       dst,
		Share:      share,
		SourcePath: "/mnt/source/share",
		DestPath:   "/mnt/dest/share",
	}

	err := validateDirectories(moveable)
	assert.NoError(t, err)
}

func TestValidateDirectory_Valid(t *testing.T) {
	dir := &schema.Directory{
		SourcePath: "/mnt/data/share/folder",
		DestPath:   "/mnt/dest/share/folder",
		Metadata: &schema.Metadata{
			IsDir:     true,
			IsSymlink: false,
		},
	}
	err := validateDirectory(dir)
	assert.NoError(t, err)
}

func TestValidateDirectory_Errors(t *testing.T) {
	tests := []struct {
		name string
		dir  *schema.Directory
		want error
	}{
		{"nil metadata", &schema.Directory{Metadata: nil}, ErrNoRelatedMetadata},
		{"symlink", &schema.Directory{Metadata: &schema.Metadata{IsSymlink: true, IsDir: false}}, ErrRelatedDirSymlink},
		{"not a dir", &schema.Directory{Metadata: &schema.Metadata{IsDir: false}}, ErrRelatedDirNotDir},
		{"empty source path", &schema.Directory{Metadata: &schema.Metadata{IsDir: true}, SourcePath: ""}, ErrNoRelatedSourcePath},
		{"relative source path", &schema.Directory{Metadata: &schema.Metadata{IsDir: true}, SourcePath: "foo"}, ErrRelatedSourceRelative},
		{"empty dest path", &schema.Directory{Metadata: &schema.Metadata{IsDir: true}, SourcePath: "/abs"}, ErrNoRelatedDestPath},
		{"relative dest path", &schema.Directory{Metadata: &schema.Metadata{IsDir: true}, SourcePath: "/abs", DestPath: "foo"}, ErrRelatedDestRelative},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateDirectory(tt.dir)
			assert.ErrorIs(t, err, tt.want)
		})
	}
}

func TestValidateDirRootConnection_Valid(t *testing.T) {
	src := &fakeStorage{name: "src", path: "/mnt/src"}
	dst := &fakeStorage{name: "dst", path: "/mnt/dst"}
	share := &fakeShare{name: "share"}

	tests := []struct {
		name       string
		sourcePath string
	}{
		{
			name:       "Root dir matches base exactly",
			sourcePath: filepath.Join(src.GetFSPath(), share.GetName()),
		},
		{
			name:       "Root dir is an ancestor of deeper path",
			sourcePath: filepath.Join(src.GetFSPath(), share.GetName(), "foo/bar"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &schema.Moveable{
				Share:      share,
				Source:     src,
				Dest:       dst,
				SourcePath: tt.sourcePath,
				RootDir: &schema.Directory{
					SourcePath: filepath.Join(src.GetFSPath(), share.GetName()),
					DestPath:   filepath.Join(dst.GetFSPath(), share.GetName()),
					Metadata:   &schema.Metadata{IsDir: true},
				},
			}

			err := validateDirRootConnection(m)
			assert.NoError(t, err)
		})
	}
}

func TestValidateDirRootConnection_Errors(t *testing.T) {
	src := &fakeStorage{name: "src", path: "/mnt/src"}
	dst := &fakeStorage{name: "dst", path: "/mnt/dst"}
	share := &fakeShare{name: "share"}

	tests := []struct {
		name    string
		modify  func(m *schema.Moveable)
		wantErr error
	}{
		{
			name: "RootDir present, but not base of SourcePath",
			modify: func(m *schema.Moveable) {
				m.SourcePath = filepath.Join(src.GetFSPath(), share.GetName(), "foo/bar/baz.txt")
				m.RootDir = &schema.Directory{
					SourcePath: filepath.Join(src.GetFSPath(), "foo"),
					DestPath:   filepath.Join(dst.GetFSPath(), "foo"),
					Metadata:   &schema.Metadata{IsDir: true},
				}
			},
			wantErr: ErrSourceNotConnectBase,
		},
		{
			name: "RootDir is nil",
			modify: func(m *schema.Moveable) {
				m.SourcePath = filepath.Join(src.GetFSPath(), share.GetName(), "foo/bar/baz.txt")
				m.RootDir = nil
			},
			wantErr: ErrSourceNotConnectBase,
		},
		{
			name: "RootDir present, but not base of DestPath",
			modify: func(m *schema.Moveable) {
				m.SourcePath = filepath.Join(src.GetFSPath(), share.GetName(), "foo/bar/baz.txt")
				m.RootDir = &schema.Directory{
					SourcePath: filepath.Join(src.GetFSPath(), share.GetName()),
					DestPath:   filepath.Join(dst.GetFSPath(), "foo"),
					Metadata:   &schema.Metadata{IsDir: true},
				}
			},
			wantErr: ErrDestNotConnectBase,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &schema.Moveable{
				Share:  share,
				Source: src,
				Dest:   dst,
			}

			tt.modify(m)

			err := validateDirRootConnection(m)
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}
