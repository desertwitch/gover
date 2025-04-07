package validation

import (
	"path/filepath"
	"testing"

	"github.com/desertwitch/gover/internal/schema"
	"github.com/desertwitch/gover/internal/schema/mocks"
	"github.com/stretchr/testify/assert"
)

// TestValidateDirectories_Success simulates a successful validation of related
// directories.
func TestValidateDirectories_Success(t *testing.T) {
	t.Parallel()

	src := mocks.NewStorage(t)
	src.On("GetFSPath").Return("/mnt/source")

	dst := mocks.NewStorage(t)
	dst.On("GetFSPath").Return("/mnt/dest")

	share := mocks.NewShare(t)
	share.On("GetName").Return("share")

	tests := []struct {
		name  string
		build func() *schema.Moveable
	}{
		{
			name: "Success_ValidDirChain",
			build: func() *schema.Moveable {
				dir2 := &schema.Directory{
					SourcePath: filepath.Join(src.GetFSPath(), "share/dir2"),
					DestPath:   filepath.Join(dst.GetFSPath(), "share/dir2"),
					Metadata:   &schema.Metadata{IsDir: true},
				}
				dir1 := &schema.Directory{
					SourcePath: filepath.Join(src.GetFSPath(), "share"),
					DestPath:   filepath.Join(dst.GetFSPath(), "share"),
					Metadata:   &schema.Metadata{IsDir: true},
					Child:      dir2,
				}
				dir2.Parent = dir1

				return &schema.Moveable{
					RootDir:    dir1,
					Source:     src,
					Dest:       dst,
					Share:      share,
					SourcePath: filepath.Join(src.GetFSPath(), "share"),
					DestPath:   filepath.Join(dst.GetFSPath(), "share"),
				}
			},
		},
		{
			name: "Success_SourceIsBaseRootIsNil",
			build: func() *schema.Moveable {
				return &schema.Moveable{
					RootDir:    nil,
					Source:     src,
					Dest:       dst,
					Share:      share,
					SourcePath: filepath.Join(src.GetFSPath(), "share"),
					DestPath:   filepath.Join(dst.GetFSPath(), "share"),
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

	src.AssertExpectations(t)
	dst.AssertExpectations(t)
	share.AssertExpectations(t)
}

// TestValidateDirectories_Fail_Errors tests a range of related directory
// validation errors.
func TestValidateDirectories_Fail_Errors(t *testing.T) {
	t.Parallel()

	src := mocks.NewStorage(t)
	src.On("GetFSPath").Return("/mnt/source")

	dst := mocks.NewStorage(t)

	share := mocks.NewShare(t)
	share.On("GetName").Return("share")

	tests := []struct {
		name     string
		modify   func(m *schema.Moveable)
		expected error
	}{
		{
			name: "Fail_ErrSourceNotConnectBase",
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
			name: "Fail_ErrNoRelatedMetadata",
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

	src.AssertExpectations(t)
	dst.AssertExpectations(t)
	share.AssertExpectations(t)
}

// TestValidateDirectory_Success simulates a successful validation of a single
// related directory.
func TestValidateDirectory_Success(t *testing.T) {
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

// TestValidateDirectory_Fail_Errors simulates a series of validation failures
// for a single related directory.
func TestValidateDirectory_Fail_Errors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		dir  *schema.Directory
		want error
	}{
		{"Fail_ErrNoRelatedMetadata", &schema.Directory{Metadata: nil}, ErrNoRelatedMetadata},
		{"Fail_ErrRelatedDirSymlink", &schema.Directory{Metadata: &schema.Metadata{IsSymlink: true, IsDir: false}}, ErrRelatedDirSymlink},
		{"Fail_ErrRelatedDirNotDir", &schema.Directory{Metadata: &schema.Metadata{IsDir: false}}, ErrRelatedDirNotDir},
		{"Fail_ErrNoRelatedSourcePath", &schema.Directory{Metadata: &schema.Metadata{IsDir: true}, SourcePath: ""}, ErrNoRelatedSourcePath},
		{"Fail_ErrRelatedSourceRelative", &schema.Directory{Metadata: &schema.Metadata{IsDir: true}, SourcePath: "foo"}, ErrRelatedSourceRelative},
		{"Fail_ErrNoRelatedDestPath", &schema.Directory{Metadata: &schema.Metadata{IsDir: true}, SourcePath: "/abs"}, ErrNoRelatedDestPath},
		{"Fail_ErrRelatedDestRelative", &schema.Directory{Metadata: &schema.Metadata{IsDir: true}, SourcePath: "/abs", DestPath: "foo"}, ErrRelatedDestRelative},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateDirectory(tt.dir)
			assert.ErrorIs(t, err, tt.want)
		})
	}
}

// TestValidateDirRootConnection_Success simulates a successful validation of a
// share base connection.
func TestValidateDirRootConnection_Success(t *testing.T) {
	t.Parallel()

	src := mocks.NewStorage(t)
	src.On("GetFSPath").Return("/mnt/src")

	dst := mocks.NewStorage(t)
	dst.On("GetFSPath").Return("/mnt/dst")

	share := mocks.NewShare(t)
	share.On("GetName").Return("share")

	tests := []struct {
		name       string
		sourcePath string
		destPath   string
		rootDir    *schema.Directory
	}{
		{
			name:       "Success_SourceIsBase_RootNil",
			sourcePath: filepath.Join(src.GetFSPath(), share.GetName()),
			destPath:   filepath.Join(dst.GetFSPath(), share.GetName()),
			rootDir:    nil,
		},
		{
			name:       "Success_BothDeep_RootBase",
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

	src.AssertExpectations(t)
	dst.AssertExpectations(t)
	share.AssertExpectations(t)
}

// TestValidateDirRootConnection_Fail_Errors simulates a series of failures
// regarding share base connection.
func TestValidateDirRootConnection_Fail_Errors(t *testing.T) {
	t.Parallel()

	src := mocks.NewStorage(t)
	src.On("GetFSPath").Return("/mnt/src")

	dst := mocks.NewStorage(t)
	dst.On("GetFSPath").Return("/mnt/dst")

	share := mocks.NewShare(t)
	share.On("GetName").Return("share")

	tests := []struct {
		name    string
		modify  func(m *schema.Moveable)
		wantErr error
	}{
		{
			name: "Fail_SourceIsDeep_RootNotBase",
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
			name: "Fail_SourceIsDeep_RootNil",
			modify: func(m *schema.Moveable) {
				m.SourcePath = filepath.Join(src.GetFSPath(), share.GetName(), "foo/bar/baz.txt")
				m.RootDir = nil
			},
			wantErr: ErrSourceNotConnectBase,
		},
		{
			name: "Fail_DestIsDeep_RootNotBase",
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

	src.AssertExpectations(t)
	dst.AssertExpectations(t)
	share.AssertExpectations(t)
}
