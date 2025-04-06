// Package unraid implements structures and routines for defining and querying
// of an Unraid-type storage system.
package unraid

import (
	"os"
)

// osProvider defines the needed operating system methods.
type osProvider interface {
	ReadDir(name string) ([]os.DirEntry, error)
}

// fsProvider defines the needed filesystem-related methods.
type fsProvider interface {
	Exists(path string) (bool, error)
}

// configProvider defines the needed configuration-related methods.
type configProvider interface {
	ReadGeneric(filenames ...string) (envMap map[string]string, err error)
	MapKeyToString(envMap map[string]string, key string) string
	MapKeyToInt(envMap map[string]string, key string) int
	MapKeyToInt64(envMap map[string]string, key string) int64
	MapKeyToUInt64(envMap map[string]string, key string) uint64
}

// Handler is the principal implementation for the Unraid services.
type Handler struct {
	fsHandler     fsProvider
	configHandler configProvider
	osHandler     osProvider
}

// NewHandler returns a pointer to a new Unraid [Handler].
func NewHandler(fsHandler fsProvider, configHandler configProvider, osHandler osProvider) *Handler {
	return &Handler{
		fsHandler:     fsHandler,
		configHandler: configHandler,
		osHandler:     osHandler,
	}
}
