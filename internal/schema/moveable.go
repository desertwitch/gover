package schema

// Moveable is the principal structure for all candidates to be moved.
// It can either be a filesystem element that is a file, link or empty folder.
// Non-empty folder structures are recorded within the [Moveable] for recreation.
//
// Moveables are meant to be passed by reference (pointer) and are not thread-safe.
//
// By design a [Moveable] and its subelements are entirely autonomous and can be processed
// independently of other [Moveable] elements, their directory structures and other fields.
type Moveable struct {
	// Share is a [Share] the [Moveable] belongs to.
	Share Share

	// Source is a [Storage] the [Moveable] is located on.
	Source Storage

	// SourcePath is an absolute path the [Moveable] is located at.
	SourcePath string

	// Dest is a [Storage] the [Moveable] is to be moved to.
	Dest Storage

	// DestPath is an absolute path the [Moveable] is to be moved to.
	DestPath string

	// Hardlinks is a slice of hardlink-type [Moveable] subelements.
	Hardlinks []*Moveable

	// IsHardlink describes if the [Moveable] is a hard-link.
	IsHardlink bool

	// HardlinkTo describes the parent [Moveable] that the [Moveable] is linked to.
	HardlinkTo *Moveable

	// Hardlinks is a slice of symlink-type [Moveable] subelements.
	Symlinks []*Moveable

	// IsSymlink describes if the [Moveable] is a sym-link.
	IsSymlink bool

	// SymlinkTo describes the parent [Moveable] that the [Moveable] is linked to.
	SymlinkTo *Moveable

	// Metadata is the filesystem [Metadata] for the specific [Moveable].
	Metadata *Metadata

	// RootDir is the shallowest [Directory], almost always a [Share] base folder,
	// at the start of a chain of [Directory] structs representing the full directory
	// structure that is required for later recreation on the destination [Storage].
	//
	// [Directory] elements and their children are always unique to a specific [Moveable].
	// This allows for a [Moveable] and its subelements to be operated on fully autonomous.
	RootDir *Directory
}

// GetMetadata returns [Metadata] for a [Moveable].
func (m *Moveable) GetMetadata() *Metadata {
	return m.Metadata
}

// GetSourcePath returns the absolute source path for a [Moveable].
func (m *Moveable) GetSourcePath() string {
	return m.SourcePath
}

// GetSourcePath returns the absolute destination path for a [Moveable].
func (m *Moveable) GetDestPath() string {
	return m.DestPath
}
