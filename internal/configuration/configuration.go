package configuration

import (
	"fmt"
	"strconv"
)

const (
	// AllocHighWater is the configuration key for the high-water allocation method.
	AllocHighWater = "highwater"

	// AllocMostFree is the configuration key for the most-free allocation method.
	AllocMostFree = "mostfree"

	// AllocFillUp is the configuration key for the fill-up allocation method.
	AllocFillUp = "fillup"
)

// genericConfigProvider defines methods for reading generic Unix-
// type configuration files, similar to "." or "source" in Bash shell.
type genericConfigProvider interface {
	Read(filenames ...string) (envMap map[string]string, err error)
}

// Handler is the principal implementation for reading configuration files.
type Handler struct {
	// An implementation of [genericConfigProvider].
	genericHandler genericConfigProvider
}

// NewHandler returns a pointer to a new configuration [Handler].
func NewHandler(genericHandler genericConfigProvider) *Handler {
	return &Handler{
		genericHandler: genericHandler,
	}
}

// ReadGeneric reads generic Unix-type configuration files
// into a map (map[key]value) or returns an error if unsuccessful.
func (c *Handler) ReadGeneric(filenames ...string) (map[string]string, error) {
	data, err := c.genericHandler.Read(filenames...)
	if err != nil {
		return data, fmt.Errorf("(config) %w", err)
	}

	return data, nil
}

// MapKeyToString returns the string representation of a given key in a
// map (map[key]value) of configuration elements (or "" on empty/error).
func (c *Handler) MapKeyToString(envMap map[string]string, key string) string {
	if value, exists := envMap[key]; exists {
		return value
	}

	return ""
}

// MapKeyToInt returns the int representation of a given key in a
// map (map[key]value) of configuration elements (or -1 on empty/error).
func (c *Handler) MapKeyToInt(envMap map[string]string, key string) int {
	value := c.MapKeyToString(envMap, key)
	if value == "" {
		return -1
	}
	intValue, err := strconv.Atoi(value)
	if err != nil {
		return -1
	}

	return intValue
}

// MapKeyToInt64 returns the int64 representation of a given key in a
// map (map[key]value) of configuration elements (or -1 on empty/error).
func (c *Handler) MapKeyToInt64(envMap map[string]string, key string) int64 {
	value := c.MapKeyToString(envMap, key)
	if value == "" {
		return -1
	}
	intValue, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return -1
	}

	return intValue
}

// MapKeyToUInt64 returns the uint64 representation of a given key in a
// map (map[key]value) of configuration elements (or 0 on empty/error).
func (c *Handler) MapKeyToUInt64(envMap map[string]string, key string) uint64 {
	value := c.MapKeyToString(envMap, key)
	if value == "" {
		return 0
	}
	intValue, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		return 0
	}

	return intValue
}
