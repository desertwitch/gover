package configuration

import (
	"strconv"
)

type genericConfigProvider interface {
	Read(filenames ...string) (envMap map[string]string, err error)
}

type ConfigHandler struct {
	GenericConfigHandler genericConfigProvider
}

func NewConfigHandler(genericHandler genericConfigProvider) *ConfigHandler {
	return &ConfigHandler{
		GenericConfigHandler: genericHandler,
	}
}

func (c *ConfigHandler) ReadGeneric(filenames ...string) (envMap map[string]string, err error) {
	return c.GenericConfigHandler.Read(filenames...)
}

func (c *ConfigHandler) MapKeyToString(envMap map[string]string, key string) string {
	if value, exists := envMap[key]; exists {
		return value
	}
	return ""
}

func (c *ConfigHandler) MapKeyToInt(envMap map[string]string, key string) int {
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

func (c *ConfigHandler) MapKeyToInt64(envMap map[string]string, key string) int64 {
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
