package configuration

import "github.com/joho/godotenv"

type GodotenvProvider struct{}

func (*GodotenvProvider) Read(filenames ...string) (map[string]string, error) {
	return godotenv.Read(filenames...)
}
