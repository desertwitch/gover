package validation

import (
	"path/filepath"
	"testing"

	"github.com/desertwitch/gover/internal/schema"
	"github.com/stretchr/testify/assert"
)

// TestValidateBasicAttributes_Success simulates successful validation of basic
// attributes.
func TestValidateBasicAttributes_Success(t *testing.T) {
	t.Parallel()

	src := schema.NewMockStorage(t)
	src.On("GetFSPath").Return("/mnt/source")

	dst := schema.NewMockStorage(t)
	dst.On("GetFSPath").Return("/mnt/dest")

	share := schema.NewMockShare(t)

	valid := &schema.Moveable{
		Share:      share,
		Metadata:   &schema.Metadata{},
		Source:     src,
		SourcePath: filepath.Join(src.GetFSPath(), "share/file"),
		Dest:       dst,
		DestPath:   filepath.Join(dst.GetFSPath(), "share/file"),
	}

	t.Run("Success_ValidMoveable", func(t *testing.T) {
		assert.NoError(t, validateBasicAttributes(valid))
	})

	src.AssertExpectations(t)
	dst.AssertExpectations(t)
	share.AssertExpectations(t)
}

// TestValidateBasicAttributes_Fail_Errors simulates a row of failures of basic
// attribute validation.
func TestValidateBasicAttributes_Fail_Errors(t *testing.T) {
	t.Parallel()

	src := schema.NewMockStorage(t)
	src.On("GetFSPath").Return("/mnt/source")

	dst := schema.NewMockStorage(t)
	dst.On("GetFSPath").Return("/mnt/dest")

	share := schema.NewMockShare(t)

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
		{"Fail_ErrNoShareInfo", func(m *schema.Moveable) { m.Share = nil }, ErrNoShareInfo},
		{"Fail_ErrNoMetadata", func(m *schema.Moveable) { m.Metadata = nil }, ErrNoMetadata},
		{"Fail_ErrNoSource", func(m *schema.Moveable) { m.Source = nil; m.SourcePath = "" }, ErrNoSource},
		{"Fail_ErrSourcePathRelative", func(m *schema.Moveable) { m.SourcePath = "relative/path" }, ErrSourcePathRelative},
		{"Fail_ErrSourceMismatch", func(m *schema.Moveable) { m.SourcePath = "/wrong/source" }, ErrSourceMismatch},
		{"Fail_ErrNoDestination", func(m *schema.Moveable) { m.Dest = nil; m.DestPath = "" }, ErrNoDestination},
		{"Fail_ErrDestPathRelative", func(m *schema.Moveable) { m.DestPath = "rel/path" }, ErrDestPathRelative},
		{"Fail_ErrDestMismatch", func(m *schema.Moveable) { m.DestPath = "/wrong/dest" }, ErrDestMismatch},
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
			name:     "Success_NotALink",
			moveable: &schema.Moveable{},
			expected: nil,
		},
		{
			name:     "Fail_ErrNoHardlinkTarget",
			moveable: &schema.Moveable{IsHardlink: true},
			expected: ErrNoHardlinkTarget,
		},
		{
			name: "Fail_ErrHardlinkHasSublinks",
			moveable: &schema.Moveable{
				IsHardlink: true,
				HardlinkTo: &schema.Moveable{},
				Hardlinks:  []*schema.Moveable{{}},
			},
			expected: ErrHardlinkHasSublinks,
		},
		{
			name:     "Fail_ErrHardlinkSetTarget",
			moveable: &schema.Moveable{HardlinkTo: &schema.Moveable{}},
			expected: ErrHardlinkSetTarget,
		},
		{
			name:     "Fail_ErrNoSymlinkTarget",
			moveable: &schema.Moveable{IsSymlink: true},
			expected: ErrNoSymlinkTarget,
		},
		{
			name: "Fail_ErrSymlinkHasSublinks",
			moveable: &schema.Moveable{
				IsSymlink: true,
				SymlinkTo: &schema.Moveable{},
				Symlinks:  []*schema.Moveable{{}},
			},
			expected: ErrSymlinkHasSublinks,
		},
		{
			name:     "Fail_ErrSymlinkSetTarget",
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
