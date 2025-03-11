package unraid

import (
	"fmt"
	"path/filepath"
	"strings"
)

type Pool struct {
	Name   string
	FSPath string
}

func (p *Pool) IsPool() bool {
	return true
}

func (p *Pool) GetName() string {
	return p.Name
}

func (p *Pool) GetFSPath() string {
	return p.FSPath
}

func (u *Handler) establishPools() (map[string]*Pool, error) {
	basePath := ConfigDirPools

	if exists, err := u.FSHandler.Exists(basePath); !exists {
		return nil, fmt.Errorf("(unraid-pools) config dir does not exist (%s): %w", basePath, err)
	}

	pools := make(map[string]*Pool)

	files, err := u.FSHandler.ReadDir(basePath)
	if err != nil {
		return nil, fmt.Errorf("(unraid-pools) failed to readdir: %w", err)
	}

	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".cfg") {
			nameWithoutExt := strings.TrimSuffix(file.Name(), ".cfg")

			fsPath := filepath.Join("/mnt", nameWithoutExt)
			if exists, _ := u.FSHandler.Exists(fsPath); !exists {
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
