package validation

import (
	"path/filepath"
	"testing"

	"github.com/desertwitch/gover/internal/schema"
	"github.com/stretchr/testify/assert"
)

// TestValidateDirectories_Valid simulates a successful validation of related
// directories.
func TestValidateDirectories_Valid(t *testing.T) {
	t.Parallel()

	src := &fakeStorage{name: "source", path: "/mnt/source"}
	dst := &fakeStorage{name: "dest", path: "/mnt/dest"}
	share := &fakeShare{name: "share"}

	tests := []struct {
		name  string
		build func() *schema.Moveable
	}{
		{
			name: "valid directory chain",
			build: func() *schema.Moveable {
				dir2 := &schema.Directory{
					SourcePath: filepath.Join(src.path, "share/dir2"),
					DestPath:   filepath.Join(dst.path, "share/dir2"),
					Metadata:   &schema.Metadata{IsDir: true},
				}
				dir1 := &schema.Directory{
					SourcePath: filepath.Join(src.path, "share"),
					DestPath:   filepath.Join(dst.path, "share"),
					Metadata:   &schema.Metadata{IsDir: true},
					Child:      dir2,
				}
				dir2.Parent = dir1

				return &schema.Moveable{
					RootDir:    dir1,
					Source:     src,
					Dest:       dst,
					Share:      share,
					SourcePath: filepath.Join(src.path, "share"),
					DestPath:   filepath.Join(dst.path, "share"),
				}
			},
		},
		{
			name: "valid with nil RootDir (SourcePath = base of Share)",
			build: func() *schema.Moveable {
				return &schema.Moveable{
					RootDir:    nil,
					Source:     src,
					Dest:       dst,
					Share:      share,
					SourcePath: filepath.Join(src.path, "share"),
					DestPath:   filepath.Join(dst.path, "share"),
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			moveable := tt.build()
			err := validateDirectories(moveable)
			assert.NoError(t, err)
		})
	}
}

// TestValidateDirectories_Errors tests a range of related directory validation
// errors.
func TestValidateDirectories_Errors(t *testing.T) {
	t.Parallel()

	src := &fakeStorage{name: "source", path: "/mnt/source"}
	dst := &fakeStorage{name: "dest", path: "/mnt/dest"}
	share := &fakeShare{name: "share"}

	tests := []struct {
		name     string
		modify   func(m *schema.Moveable)
		expected error
	}{
		{
			name: "invalid root connection (src)",
			modify: func(m *schema.Moveable) {
				m.RootDir = &schema.Directory{
					SourcePath: "/mnt/source/share/foo",
					DestPath:   "/mnt/dest/share",
					Metadata:   &schema.Metadata{IsDir: true},
				}
				m.SourcePath = "/mnt/source/share/foo/bar/baz.txt"
				m.DestPath = "/mnt/dest/share/foo/bar/baz.txt"
			},
			expected: ErrSourceNotConnectBase,
		},
		{
			name: "invalid directory metadata (nil)",
			modify: func(m *schema.Moveable) {
				m.RootDir = &schema.Directory{
					Metadata: nil,
				}
				m.SourcePath = "/mnt/source/share"
				m.DestPath = "/mnt/dest/share"
			},
			expected: ErrNoRelatedMetadata,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &schema.Moveable{
				Source: src,
				Dest:   dst,
				Share:  share,
			}
			tt.modify(m)

			err := validateDirectories(m)
			assert.ErrorIs(t, err, tt.expected)
		})
	}
}

// TestValidateDirectory_Valid simulates a successful validation of a single
// related directory.
func TestValidateDirectory_Valid(t *testing.T) {
	t.Parallel()

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

// TestValidateDirectory_Errors simulates a series of validation failures for a
// single related directory.
func TestValidateDirectory_Errors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		dir  *schema.Directory
		want error
	}{
		{"related dir has no metadata", &schema.Directory{Metadata: nil}, ErrNoRelatedMetadata},
		{"related dir is symlink", &schema.Directory{Metadata: &schema.Metadata{IsSymlink: true, IsDir: false}}, ErrRelatedDirSymlink},
		{"related dir not a dir", &schema.Directory{Metadata: &schema.Metadata{IsDir: false}}, ErrRelatedDirNotDir},
		{"related dir empty source path", &schema.Directory{Metadata: &schema.Metadata{IsDir: true}, SourcePath: ""}, ErrNoRelatedSourcePath},
		{"related dir source path relative", &schema.Directory{Metadata: &schema.Metadata{IsDir: true}, SourcePath: "foo"}, ErrRelatedSourceRelative},
		{"related dir empty dest path", &schema.Directory{Metadata: &schema.Metadata{IsDir: true}, SourcePath: "/abs"}, ErrNoRelatedDestPath},
		{"related dir dest path relative", &schema.Directory{Metadata: &schema.Metadata{IsDir: true}, SourcePath: "/abs", DestPath: "foo"}, ErrRelatedDestRelative},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateDirectory(tt.dir)
			assert.ErrorIs(t, err, tt.want)
		})
	}
}

// TestValidateDirRootConnection_Valid simulates a successful validation of a
// share base connection.
func TestValidateDirRootConnection_Valid(t *testing.T) {
	t.Parallel()

	src := &fakeStorage{name: "src", path: "/mnt/src"}
	dst := &fakeStorage{name: "dst", path: "/mnt/dst"}
	share := &fakeShare{name: "share"}

	tests := []struct {
		name       string
		sourcePath string
		destPath   string
		rootDir    *schema.Directory
	}{
		{
			name:       "source path is share base, root dir is nil",
			sourcePath: filepath.Join(src.GetFSPath(), share.GetName()),
			destPath:   filepath.Join(dst.GetFSPath(), share.GetName()),
			rootDir:    nil,
		},
		{
			name:       "source path is deeper path, root dir is share base",
			sourcePath: filepath.Join(src.GetFSPath(), share.GetName(), "foo/bar"),
			destPath:   filepath.Join(dst.GetFSPath(), share.GetName(), "foo/bar"),
			rootDir: &schema.Directory{
				SourcePath: filepath.Join(src.GetFSPath(), share.GetName()),
				DestPath:   filepath.Join(dst.GetFSPath(), share.GetName()),
				Metadata:   &schema.Metadata{IsDir: true},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &schema.Moveable{
				Share:      share,
				Source:     src,
				Dest:       dst,
				SourcePath: tt.sourcePath,
				DestPath:   tt.destPath,
				RootDir:    tt.rootDir,
			}

			err := validateDirRootConnection(m)
			assert.NoError(t, err)
		})
	}
}

// TestValidateDirRootConnection_Errors simulates a series of failures regarding
// share base connection.
func TestValidateDirRootConnection_Errors(t *testing.T) {
	t.Parallel()

	src := &fakeStorage{name: "src", path: "/mnt/src"}
	dst := &fakeStorage{name: "dst", path: "/mnt/dst"}
	share := &fakeShare{name: "share"}

	tests := []struct {
		name    string
		modify  func(m *schema.Moveable)
		wantErr error
	}{
		{
			name: "source path is deeper path, root dir src is not share base",
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
			name: "source path is deeper path, root dir is nil",
			modify: func(m *schema.Moveable) {
				m.SourcePath = filepath.Join(src.GetFSPath(), share.GetName(), "foo/bar/baz.txt")
				m.RootDir = nil
			},
			wantErr: ErrSourceNotConnectBase,
		},
		{
			name: "dest path is deeper path, root dir dest is not share base",
			modify: func(m *schema.Moveable) {
				m.SourcePath = filepath.Join(src.GetFSPath(), share.GetName(), "foo/bar/baz.txt")
				m.DestPath = filepath.Join(dst.GetFSPath(), share.GetName(), "foo/bar/baz.txt")
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
