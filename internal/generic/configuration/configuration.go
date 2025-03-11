package configuration

import (
	"strconv"
)

const (
	AllocHighWater = "highwater"
	AllocMostFree  = "mostfree"
	AllocFillUp    = "fillup"
)

type genericConfigProvider interface {
	Read(filenames ...string) (envMap map[string]string, err error)
}

type Handler struct {
	GenericHandler genericConfigProvider
}

func NewHandler(genericHandler genericConfigProvider) *Handler {
	return &Handler{
		GenericHandler: genericHandler,
	}
}

func (c *Handler) ReadGeneric(filenames ...string) (map[string]string, error) {
	return c.GenericHandler.Read(filenames...)
}

func (c *Handler) MapKeyToString(envMap map[string]string, key string) string {
	if value, exists := envMap[key]; exists {
		return value
	}

	return ""
}

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
