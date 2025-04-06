package unraid

import (
	"fmt"
	"path/filepath"
	"strings"
)

// Pool is an Unraid pool, as part of an Unraid [System].
type Pool struct {
	Name   string
	FSPath string
}

// IsPool is a type identifier.
func (p *Pool) IsPool() bool {
	return true
}

// GetName returns the pool's name.
func (p *Pool) GetName() string {
	return p.Name
}

// GetFSPath returns an absolute filesystem path to the pool's mountpoint.
func (p *Pool) GetFSPath() string {
	return p.FSPath
}

// establishPools returns a map (map[poolName]*Pool) to all Unraid [Pool]. It is
// the principal method for reading all pool information from the system.
func (u *Handler) establishPools() (map[string]*Pool, error) {
	basePath := ConfigDirPools

	if exists, err := u.fsHandler.Exists(basePath); !exists {
		return nil, fmt.Errorf("(unraid-pools) config dir does not exist (%s): %w", basePath, err)
	}

	pools := make(map[string]*Pool)

	files, err := u.osHandler.ReadDir(basePath)
	if err != nil {
		return nil, fmt.Errorf("(unraid-pools) failed to readdir: %w", err)
	}

	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".cfg") {
			nameWithoutExt := strings.TrimSuffix(file.Name(), ".cfg")

			fsPath := filepath.Join("/mnt", nameWithoutExt)
			if exists, _ := u.fsHandler.Exists(fsPath); !exists {
				return nil, fmt.Errorf("(unraid-pools) mountpoint does not exist (%s): %w", fsPath, err)
			}

			pool := &Pool{
				Name:   nameWithoutExt,
				FSPath: fsPath,
			}

			pools[pool.Name] = pool
		}
	}

	return pools, nil
}
