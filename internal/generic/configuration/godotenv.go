package configuration

import (
	"fmt"

	"github.com/joho/godotenv"
)

type GodotenvProvider struct{}

func (*GodotenvProvider) Read(filenames ...string) (map[string]string, error) {
	data, err := godotenv.Read(filenames...)

	return data, fmt.Errorf("(config-godotenv) %w", err)
}
