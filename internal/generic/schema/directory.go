package schema

type Directory struct {
	SourcePath string
	DestPath   string
	Metadata   *Metadata
	Parent     *Directory
	Child      *Directory
}

func (d *Directory) GetMetadata() *Metadata {
	return d.Metadata
}

func (d *Directory) GetSourcePath() string {
	return d.SourcePath
}

func (d *Directory) GetDestPath() string {
	return d.DestPath
}
