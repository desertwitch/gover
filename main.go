package main

import (
	"fmt"
	"strings"
)

func diskNames(disks []*UnraidDisk) string {
	var names []string
	for _, disk := range disks {
		names = append(names, disk.Name)
	}
	return strings.Join(names, ", ")
}

func main() {
	disks, err := establishDisks()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	for _, disk := range disks {
		fmt.Printf("Discovered Disk: Name=%s, Path=%s\n", disk.Name, disk.FSPath)
	}

	pools, err := establishPools()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	for _, pool := range pools {
		fmt.Printf("Discovered Pool: Name=%s, FSPath=%s, CFGFile=%s\n", pool.Name, pool.FSPath, pool.CFGFile)
	}

	// Read shares
	shares, err := establishShares(disks, pools)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Print results
	for _, share := range shares {
		fmt.Printf("Share: %s\n", share.Name)
		fmt.Printf("  Use Cache: %s\n", share.UseCache)
		fmt.Printf("  Cache Pool: %v\n", share.CachePool)
		fmt.Printf("  Cache Pool 2: %v\n", share.CachePool2)
		fmt.Printf("  Allocator: %s\n", share.Allocator)
		fmt.Printf("  Split Level: %d\n", share.SplitLevel)
		fmt.Printf("  Space Floor: %d\n", share.SpaceFloor)
		fmt.Printf("  Disable COW: %v\n", share.DisableCOW)
		fmt.Printf("  Included Disks: %v\n", share.IncludedDisks)
		fmt.Printf("  Excluded Disks: %v\n", share.ExcludedDisks)
		fmt.Println()
	}
}
