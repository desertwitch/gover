package main

import (
	"fmt"
)

func main() {
	system, err := establishSystem()
	if err != nil {
		fmt.Printf("Error: %+v", err)
	}

	shares := system.Shares
	disks := system.Array.Disks

	moveables := []*Moveable{}

	// Primary to Secondary
	for _, share := range shares {
		if share.UseCache != "yes" || share.CachePool == nil {
			continue
		}
		files, err := getMoveables(share.CachePool, share)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}
		moveables = append(moveables, files...)
	}

	// Secondary to Primary
	for _, share := range shares {
		if share.UseCache != "prefer" || share.CachePool == nil {
			continue
		}
		if share.CachePool2 == nil {
			// Array to Cache
			for _, disk := range disks {
				files, err := getMoveables(disk, share)
				if err != nil {
					fmt.Printf("Error: %v\n", err)
					continue
				}
				moveables = append(moveables, files...)
			}
		} else {
			// Cache2 to Cache
			files, err := getMoveables(share.CachePool2, share)
			if err != nil {
				fmt.Printf("Error: %v", err)
				continue
			}
			moveables = append(moveables, files...)
		}
	}

	for _, moveable := range moveables {
		fmt.Println("-------------------------------------------")
		fmt.Printf("[C] %s (I:%d/S:%d) [%v] [%v]\n", moveable.Path, moveable.Metadata.Inode, moveable.Metadata.Size, moveable.Metadata, moveable.ParentDirs)

		if moveable.Share.SplitLevel >= 0 {
			d, err := allocateDisksBySplitLevel(moveable, moveable.Share.SplitLevel)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
			}
			fmt.Printf("%v", d)
		}
	}
}
