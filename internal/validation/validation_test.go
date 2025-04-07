package validation

import (
	"path/filepath"
	"testing"

	"github.com/desertwitch/gover/internal/schema"
	"github.com/stretchr/testify/assert"
)

type fakeStorage struct {
	name string
	path string
}

func (fs *fakeStorage) GetName() string   { return fs.name }
func (fs *fakeStorage) GetFSPath() string { return fs.path }

type fakeShare struct {
	name string
}

func (s *fakeShare) GetName() string                          { return s.name }
func (s *fakeShare) GetUseCache() string                      { return "" }
func (s *fakeShare) GetCachePool() schema.Pool                { return nil }
func (s *fakeShare) GetCachePool2() schema.Pool               { return nil }
func (s *fakeShare) GetAllocator() string                     { return "" }
func (s *fakeShare) GetSplitLevel() int                       { return 0 }
func (s *fakeShare) GetSpaceFloor() uint64                    { return 0 }
func (s *fakeShare) GetDisableCOW() bool                      { return false }
func (s *fakeShare) GetIncludedDisks() map[string]schema.Disk { return nil }

// TestValidateMoveable tests validation of a [schema.Moveable] with and without
// subelements.
func TestValidateMoveable(t *testing.T) {
	t.Parallel()

	src := &fakeStorage{name: "src", path: "/mnt/src"}
	dst := &fakeStorage{name: "dst", path: "/mnt/dst"}
	share := &fakeShare{name: "share"}

	basePath := filepath.Join(src.path, share.name)
	destPath := filepath.Join(dst.path, share.name)

	makeValid := func(path string) *schema.Moveable {
		return &schema.Moveable{
			Share:      share,
			Source:     src,
			SourcePath: path,
			Dest:       dst,
			DestPath:   filepath.Join(dst.path, "share", "file"),
			Metadata:   &schema.Metadata{},
			RootDir: &schema.Directory{
				SourcePath: basePath,
				DestPath:   destPath,
				Metadata:   &schema.Metadata{IsDir: true},
			},
		}
	}

	tests := []struct {
		name     string
		modify   func(m *schema.Moveable)
		expected bool
	}{
		{
			name:     "valid moveable with no links",
			modify:   func(m *schema.Moveable) {},
			expected: true,
		},
		{
			name: "invalid root moveable (relative source path)",
			modify: func(m *schema.Moveable) {
				m.SourcePath = "relative/path"
			},
			expected: false,
		},
		{
			name: "invalid hardlink fails moveable validation",
			modify: func(m *schema.Moveable) {
				m.Hardlinks = []*schema.Moveable{
					{
						Share:      share,
						Source:     src,
						SourcePath: "invalid/hardlink",
						Dest:       dst,
						DestPath:   filepath.Join(dst.path, "bad"),
						Metadata:   &schema.Metadata{},
					},
				}
			},
			expected: false,
		},
		{
			name: "invalid symlink fails moveable validation",
			modify: func(m *schema.Moveable) {
				m.Symlinks = []*schema.Moveable{
					{
						Share:      share,
						Source:     src,
						SourcePath: "invalid/symlink",
						Dest:       dst,
						DestPath:   filepath.Join(dst.path, "bad"),
						Metadata:   &schema.Metadata{},
					},
				}
			},
			expected: false,
		},
		{
			name: "valid moveable with valid hardlink and symlink",
			modify: func(m *schema.Moveable) {
				m.Hardlinks = []*schema.Moveable{makeValid(filepath.Join(src.path, "share/hardlink"))}
				m.Symlinks = []*schema.Moveable{makeValid(filepath.Join(src.path, "share/symlink"))}
			},
			expected: true,
		},
		{
			name: "link validation fails",
			modify: func(m *schema.Moveable) {
				m.IsHardlink = true
				m.HardlinkTo = nil
			},
			expected: false,
		},
		{
			name: "directory validation fails",
			modify: func(m *schema.Moveable) {
				m.RootDir = &schema.Directory{
					SourcePath: "not/abs",
					DestPath:   "/mnt/dst/share",
					Metadata:   &schema.Metadata{IsDir: true},
				}
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := makeValid(filepath.Join(src.path, "share/file"))
			tt.modify(m)
			assert.Equal(t, tt.expected, ValidateMoveable(m))
		})
	}
}
