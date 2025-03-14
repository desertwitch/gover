package schema

type Moveable struct {
	Share      Share
	Source     Storage
	SourcePath string
	Dest       Storage
	DestPath   string
	Hardlinks  []*Moveable
	IsHardlink bool
	HardlinkTo *Moveable
	Symlinks   []*Moveable
	IsSymlink  bool
	SymlinkTo  *Moveable
	Metadata   *Metadata
	RootDir    *Directory
}

func (m *Moveable) GetMetadata() *Metadata {
	return m.Metadata
}

func (m *Moveable) GetSourcePath() string {
	return m.SourcePath
}

func (m *Moveable) GetDestPath() string {
	return m.DestPath
}
