package validation

import (
	"path/filepath"
	"testing"

	"github.com/desertwitch/gover/internal/schema"
	"github.com/desertwitch/gover/internal/schema/mocks"
	"github.com/stretchr/testify/assert"
)

// TestValidateBasicAttributes_Valid simulates successful validation of basic attributes.
func TestValidateBasicAttributes_Valid(t *testing.T) {
	t.Parallel()

	src := mocks.NewStorage(t)
	src.On("GetFSPath").Return("/mnt/source")

	dst := mocks.NewStorage(t)
	dst.On("GetFSPath").Return("/mnt/dest")

	share := mocks.NewShare(t)

	valid := &schema.Moveable{
		Share:      share,
		Metadata:   &schema.Metadata{},
		Source:     src,
		SourcePath: filepath.Join(src.GetFSPath(), "share/file"),
		Dest:       dst,
		DestPath:   filepath.Join(dst.GetFSPath(), "share/file"),
	}

	t.Run("valid moveable", func(t *testing.T) {
		assert.NoError(t, validateBasicAttributes(valid))
	})

	src.AssertExpectations(t)
	dst.AssertExpectations(t)
	share.AssertExpectations(t)
}

// TestValidateBasicAttributes_Errors simulates a row of failures of basic attribute validation.
func TestValidateBasicAttributes_Errors(t *testing.T) {
	t.Parallel()

	src := mocks.NewStorage(t)
	src.On("GetFSPath").Return("/mnt/source")

	dst := mocks.NewStorage(t)
	dst.On("GetFSPath").Return("/mnt/dest")

	share := mocks.NewShare(t)

	valid := &schema.Moveable{
		Share:      share,
		Metadata:   &schema.Metadata{},
		Source:     src,
		SourcePath: filepath.Join(src.GetFSPath(), "share/file"),
		Dest:       dst,
		DestPath:   filepath.Join(dst.GetFSPath(), "share/file"),
	}

	tests := []struct {
		name string
		mod  func(m *schema.Moveable)
		want error
	}{
		{"missing share", func(m *schema.Moveable) { m.Share = nil }, ErrNoShareInfo},
		{"missing metadata", func(m *schema.Moveable) { m.Metadata = nil }, ErrNoMetadata},
		{"missing source", func(m *schema.Moveable) { m.Source = nil; m.SourcePath = "" }, ErrNoSource},
		{"relative source path", func(m *schema.Moveable) { m.SourcePath = "relative/path" }, ErrSourcePathRelative},
		{"source mismatch", func(m *schema.Moveable) { m.SourcePath = "/wrong/source" }, ErrSourceMismatch},
		{"missing dest", func(m *schema.Moveable) { m.Dest = nil; m.DestPath = "" }, ErrNoDestination},
		{"relative dest path", func(m *schema.Moveable) { m.DestPath = "rel/path" }, ErrDestPathRelative},
		{"dest mismatch", func(m *schema.Moveable) { m.DestPath = "/wrong/dest" }, ErrDestMismatch},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := *valid // shallow copy
			tt.mod(&m)
			err := validateBasicAttributes(&m)
			assert.ErrorIs(t, err, tt.want)
		})
	}

	src.AssertExpectations(t)
	dst.AssertExpectations(t)
	share.AssertExpectations(t)
}

// TestValidateLinks tests the validation of links.
func TestValidateLinks(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		moveable *schema.Moveable
		expected error
	}{
		{
			name:     "not a link",
			moveable: &schema.Moveable{},
			expected: nil,
		},
		{
			name:     "hardlink: missing target",
			moveable: &schema.Moveable{IsHardlink: true},
			expected: ErrNoHardlinkTarget,
		},
		{
			name: "hardlink: has sublinks",
			moveable: &schema.Moveable{
				IsHardlink: true,
				HardlinkTo: &schema.Moveable{},
				Hardlinks:  []*schema.Moveable{{}},
			},
			expected: ErrHardlinkHasSublinks,
		},
		{
			name:     "hardlink set without flag",
			moveable: &schema.Moveable{HardlinkTo: &schema.Moveable{}},
			expected: ErrHardlinkSetTarget,
		},
		{
			name:     "symlink: missing target",
			moveable: &schema.Moveable{IsSymlink: true},
			expected: ErrNoSymlinkTarget,
		},
		{
			name: "symlink: has sublinks",
			moveable: &schema.Moveable{
				IsSymlink: true,
				SymlinkTo: &schema.Moveable{},
				Symlinks:  []*schema.Moveable{{}},
			},
			expected: ErrSymlinkHasSublinks,
		},
		{
			name:     "symlink set without flag",
			moveable: &schema.Moveable{SymlinkTo: &schema.Moveable{}},
			expected: ErrSymlinkSetTarget,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateLinks(tt.moveable)
			assert.ErrorIs(t, err, tt.expected)
		})
	}
}
