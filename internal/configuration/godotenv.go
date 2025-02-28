package configuration

import "github.com/joho/godotenv"

type GodotenvProvider struct{}

func (*GodotenvProvider) Read(filenames ...string) (envMap map[string]string, err error) {
	return godotenv.Read(filenames...)
}
