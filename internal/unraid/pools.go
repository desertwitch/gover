package unraid

import (
	"fmt"
	"path/filepath"
	"strings"
)

type Pool struct {
	Name           string
	FSPath         string
	CFGFile        string
	ActiveTransfer bool
}

func (p *Pool) GetName() string {
	return p.Name
}

func (p *Pool) GetFSPath() string {
	return p.FSPath
}

func (p *Pool) IsActiveTransfer() bool {
	return p.ActiveTransfer
}

func (p *Pool) SetActiveTransfer(active bool) {
	p.ActiveTransfer = active
}

// TO-DO: Refactor into establishPool() and establishPools().
func (u *Handler) EstablishPools() (map[string]*Pool, error) {
	basePath := ConfigDirPools

	if exists, err := u.FSOps.Exists(basePath); !exists {
		return nil, fmt.Errorf("pool config dir does not exist: %w", err)
	}

	pools := make(map[string]*Pool)

	files, err := u.FSOps.ReadDir(basePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read pool config dir: %w", err)
	}

	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".cfg") {
			cfgPath := filepath.Join(basePath, file.Name())
			nameWithoutExt := strings.TrimSuffix(file.Name(), ".cfg")

			fsPath := filepath.Join("/mnt", nameWithoutExt)
			if exists, err := u.FSOps.Exists(fsPath); !exists {
				return nil, fmt.Errorf("pool mount %s does not exist: %w", fsPath, err)
			}

			pool := &Pool{
				Name:           nameWithoutExt,
				FSPath:         fsPath,
				CFGFile:        cfgPath,
				ActiveTransfer: false,
			}

			pools[pool.Name] = pool
		}
	}

	return pools, nil
}
