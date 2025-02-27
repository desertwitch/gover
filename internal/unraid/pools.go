package unraid

import (
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
)

type UnraidPool struct {
	Name           string
	FSPath         string
	CFGFile        string
	ActiveTransfer bool
}

func (p *UnraidPool) GetName() string {
	return p.Name
}

func (p *UnraidPool) GetFSPath() string {
	return p.FSPath
}

func (p *UnraidPool) IsActiveTransfer() bool {
	return p.ActiveTransfer
}

func (p *UnraidPool) SetActiveTransfer(active bool) {
	p.ActiveTransfer = active
}

// establishPools returns a map of pointers to established Unraid pools
// TO-DO: Refactor into establishPool() and establishPools()
func establishPools(osOps osProvider) (map[string]*UnraidPool, error) {
	basePath := ConfigDirPools

	if _, err := osOps.Stat(basePath); errors.Is(err, fs.ErrNotExist) {
		return nil, fmt.Errorf("pool config dir does not exist: %w", err)
	}

	pools := make(map[string]*UnraidPool)

	files, err := osOps.ReadDir(basePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read pool config dir: %w", err)
	}

	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".cfg") {
			cfgPath := filepath.Join(basePath, file.Name())
			nameWithoutExt := strings.TrimSuffix(file.Name(), ".cfg")

			fsPath := filepath.Join("/mnt", nameWithoutExt)
			if _, err := osOps.Stat(fsPath); errors.Is(err, fs.ErrNotExist) {
				return nil, fmt.Errorf("pool mount %s does not exist: %w", fsPath, err)
			}

			pool := &UnraidPool{
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
