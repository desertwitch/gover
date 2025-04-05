package filesystem

import "github.com/desertwitch/gover/internal/schema"

// establishSymlinks cross-references a slice of [schema.Moveable] for symbolic
// links pointing from one [schema.Moveable] to another [schema.Moveable],
// linking them with each other.
func establishSymlinks(moveables []*schema.Moveable, dst schema.Storage) {
	realFiles := make(map[string]*schema.Moveable)

	for _, m := range moveables {
		if !m.IsHardlink && !m.Metadata.IsSymlink {
			realFiles[m.SourcePath] = m
		}
	}

	for _, m := range moveables {
		if m.Metadata.IsSymlink {
			if target, exists := realFiles[m.Metadata.SymlinkTo]; exists {
				m.IsSymlink = true
				m.SymlinkTo = target

				m.Dest = dst
				target.Symlinks = append(target.Symlinks, m)
			}
		}
	}
}

// establishHardlinks cross-references a slice of [schema.Moveable] for hard
// links pointing from one [schema.Moveable] to another [schema.Moveable],
// linking them with each other.
func establishHardlinks(moveables []*schema.Moveable, dst schema.Storage) {
	inodes := make(map[uint64]*schema.Moveable)

	for _, m := range moveables {
		if target, exists := inodes[m.Metadata.Inode]; exists {
			m.IsHardlink = true
			m.HardlinkTo = target

			m.Dest = dst
			target.Hardlinks = append(target.Hardlinks, m)
		} else {
			inodes[m.Metadata.Inode] = m
		}
	}
}

// removeInternalLinks cleans up a slice of [schema.Moveable] removing all
// symbolic and hard links, which were previously linked to another
// [schema.Moveable], so that only [schema.Moveable] "parents" remain (with
// their internal link fields set to these links).
//
// This is to ensure that only a group of "parent" [schema.Moveable] elements
// remain, which are internally linked to their respective subelements. This
// allows for processing them on their own, while being able to reconstruct all
// symbolic and hard link structures.
func removeInternalLinks(moveables []*schema.Moveable) []*schema.Moveable {
	var filtered []*schema.Moveable

	for _, m := range moveables {
		if !m.IsSymlink && !m.IsHardlink {
			filtered = append(filtered, m)
		}
	}

	return filtered
}
