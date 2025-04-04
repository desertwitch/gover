package configuration

import (
	"fmt"

	"github.com/joho/godotenv"
)

// GodotenvProvider is an implementation wrapping the Gotdotenv framework.
type GodotenvProvider struct{}

// Read reads generic Unix-type configuration files into a map (map[key]value).
func (*GodotenvProvider) Read(filenames ...string) (map[string]string, error) {
	data, err := godotenv.Read(filenames...)
	if err != nil {
		return data, fmt.Errorf("(config-godotenv) %w", err)
	}

	return data, nil
}
