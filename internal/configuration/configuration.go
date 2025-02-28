package configuration

import (
	"strconv"
)

type genericConfigProvider interface {
	Read(filenames ...string) (envMap map[string]string, err error)
}

type ConfigProviderImpl struct {
	GenericConfigReader genericConfigProvider
}

func (c *ConfigProviderImpl) ReadGeneric(filenames ...string) (envMap map[string]string, err error) {
	return c.GenericConfigReader.Read(filenames...)
}

func (c *ConfigProviderImpl) MapKeyToString(envMap map[string]string, key string) string {
	if value, exists := envMap[key]; exists {
		return value
	}
	return ""
}

func (c *ConfigProviderImpl) MapKeyToInt(envMap map[string]string, key string) int {
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

func (c *ConfigProviderImpl) MapKeyToInt64(envMap map[string]string, key string) int64 {
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
