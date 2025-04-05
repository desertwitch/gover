package schema

// Directory is a non-empty directory that is recorded as part of a [Moveable]'s
// directory structure, but is not a [Moveable] itself. It is meant to be passed
// by reference (pointer).
type Directory struct {
	SourcePath string
	DestPath   string
	Metadata   *Metadata
	Parent     *Directory
	Child      *Directory
}

// GetMetadata returns [Metadata] for a [Directory].
func (d *Directory) GetMetadata() *Metadata {
	return d.Metadata
}

// GetSourcePath returns the absolute source path for a [Directory].
func (d *Directory) GetSourcePath() string {
	return d.SourcePath
}

// GetDestPath returns the absolute destination path for a [Directory].
func (d *Directory) GetDestPath() string {
	return d.DestPath
}
