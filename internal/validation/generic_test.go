package validation

import (
	"path/filepath"
	"testing"

	"github.com/desertwitch/gover/internal/schema"
	"github.com/stretchr/testify/assert"
)

func TestValidateBasicAttributes(t *testing.T) {
	sourceBase := "/mnt/src"
	destBase := "/mnt/dst"

	valid := &schema.Moveable{
		Share:      &fakeShare{"share"},
		Metadata:   &schema.Metadata{},
		Source:     &fakeStorage{name: "source", path: "/mnt/src"},
		SourcePath: filepath.Join(sourceBase, "share/file"),
		Dest:       &fakeStorage{name: "dest", path: "/mnt/dst"},
		DestPath:   filepath.Join(destBase, "share/file"),
	}

	t.Run("valid moveable", func(t *testing.T) {
		assert.NoError(t, validateBasicAttributes(valid))
	})

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
}

func TestValidateLinks(t *testing.T) {
	t.Run("normal file", func(t *testing.T) {
		assert.NoError(t, validateLinks(&schema.Moveable{}))
	})

	t.Run("hardlink: missing target", func(t *testing.T) {
		m := &schema.Moveable{IsHardlink: true}
		assert.ErrorIs(t, validateLinks(m), ErrNoHardlinkTarget)
	})

	t.Run("hardlink: has sublinks", func(t *testing.T) {
		m := &schema.Moveable{
			IsHardlink: true,
			HardlinkTo: &schema.Moveable{},
			Hardlinks:  []*schema.Moveable{{}},
		}
		assert.ErrorIs(t, validateLinks(m), ErrHardlinkHasSublinks)
	})

	t.Run("hardlink set without flag", func(t *testing.T) {
		m := &schema.Moveable{HardlinkTo: &schema.Moveable{}}
		assert.ErrorIs(t, validateLinks(m), ErrHardlinkSetTarget)
	})

	t.Run("symlink: missing target", func(t *testing.T) {
		m := &schema.Moveable{IsSymlink: true}
		assert.ErrorIs(t, validateLinks(m), ErrNoSymlinkTarget)
	})

	t.Run("symlink: has sublinks", func(t *testing.T) {
		m := &schema.Moveable{
			IsSymlink: true,
			SymlinkTo: &schema.Moveable{},
			Symlinks:  []*schema.Moveable{{}},
		}
		assert.ErrorIs(t, validateLinks(m), ErrSymlinkHasSublinks)
	})

	t.Run("symlink set without flag", func(t *testing.T) {
		m := &schema.Moveable{SymlinkTo: &schema.Moveable{}}
		assert.ErrorIs(t, validateLinks(m), ErrSymlinkSetTarget)
	})
}
