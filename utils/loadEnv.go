package utils

import (
	"github.com/joho/godotenv"
)

func LoadEnv(path string) error {
	err := godotenv.Load(path)
	if err != nil {
		return err
	}

	return nil
}
