package validation

import (
	"path/filepath"
	"testing"

	"github.com/desertwitch/gover/internal/schema"
	"github.com/desertwitch/gover/internal/schema/mocks"
	"github.com/stretchr/testify/assert"
)

// TestValidateMoveable tests validation of a [schema.Moveable] with and without
// subelements.
func TestValidateMoveable(t *testing.T) {
	t.Parallel()

	src := mocks.NewStorage(t)
	src.On("GetFSPath").Return("/mnt/src")

	dst := mocks.NewStorage(t)
	dst.On("GetFSPath").Return("/mnt/dst")

	share := mocks.NewShare(t)
	share.On("GetName").Return("share")

	basePath := filepath.Join(src.GetFSPath(), share.GetName())
	destPath := filepath.Join(dst.GetFSPath(), share.GetName())

	makeValid := func(path string) *schema.Moveable {
		return &schema.Moveable{
			Share:      share,
			Source:     src,
			SourcePath: path,
			Dest:       dst,
			DestPath:   filepath.Join(dst.GetFSPath(), "share", "file"),
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
						DestPath:   filepath.Join(dst.GetFSPath(), "bad"),
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
						DestPath:   filepath.Join(dst.GetFSPath(), "bad"),
						Metadata:   &schema.Metadata{},
					},
				}
			},
			expected: false,
		},
		{
			name: "valid moveable with valid hardlink and symlink",
			modify: func(m *schema.Moveable) {
				m.Hardlinks = []*schema.Moveable{makeValid(filepath.Join(src.GetFSPath(), "share/hardlink"))}
				m.Symlinks = []*schema.Moveable{makeValid(filepath.Join(src.GetFSPath(), "share/symlink"))}
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
			m := makeValid(filepath.Join(src.GetFSPath(), "share/file"))
			tt.modify(m)
			assert.Equal(t, tt.expected, ValidateMoveable(m))
		})
	}

	src.AssertExpectations(t)
	dst.AssertExpectations(t)
	share.AssertExpectations(t)
}
