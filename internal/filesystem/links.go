package filesystem

func establishSymlinks(moveables []*Moveable, dst StorageType) {
	realFiles := make(map[string]*Moveable)

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

func establishHardlinks(moveables []*Moveable, dst StorageType) {
	inodes := make(map[uint64]*Moveable)
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

func removeInternalLinks(moveables []*Moveable) []*Moveable {
	var filtered []*Moveable

	for _, m := range moveables {
		if !m.IsSymlink && !m.IsHardlink {
			filtered = append(filtered, m)
		}
	}

	return filtered
}
