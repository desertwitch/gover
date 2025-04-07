package validation

import (
	"path/filepath"
	"testing"

	"github.com/desertwitch/gover/internal/schema"
	"github.com/stretchr/testify/assert"
)

// TestValidateMoveable tests validation of a [schema.Moveable] with and without
// subelements.
func TestValidateMoveable(t *testing.T) {
	t.Parallel()

	src := schema.NewMockStorage(t)
	src.On("GetFSPath").Return("/mnt/src")

	dst := schema.NewMockStorage(t)
	dst.On("GetFSPath").Return("/mnt/dst")

	share := schema.NewMockShare(t)
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
			name:     "Success_Valid_NoLinks",
			modify:   func(m *schema.Moveable) {},
			expected: true,
		},
		{
			name: "Fail_Elem_RelSourcePath",
			modify: func(m *schema.Moveable) {
				m.SourcePath = "relative/path"
			},
			expected: false,
		},
		{
			name: "Fail_Hardlink_FailsBasic",
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
			name: "Fail_Symlink_FailsBasic",
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
			name: "Success_Valid_With_Valid_Subelems",
			modify: func(m *schema.Moveable) {
				m.Hardlinks = []*schema.Moveable{makeValid(filepath.Join(src.GetFSPath(), "share/hardlink"))}
				m.Symlinks = []*schema.Moveable{makeValid(filepath.Join(src.GetFSPath(), "share/symlink"))}
			},
			expected: true,
		},
		{
			name: "Fail_Fails_LinkValid",
			modify: func(m *schema.Moveable) {
				m.IsHardlink = true
				m.HardlinkTo = nil
			},
			expected: false,
		},
		{
			name: "Fail_Fails_DirValid",
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
