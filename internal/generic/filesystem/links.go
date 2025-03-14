package filesystem

import "github.com/desertwitch/gover/internal/generic/schema"

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

func removeInternalLinks(moveables []*schema.Moveable) []*schema.Moveable {
	var filtered []*schema.Moveable

	for _, m := range moveables {
		if !m.IsSymlink && !m.IsHardlink {
			filtered = append(filtered, m)
		}
	}

	return filtered
}
