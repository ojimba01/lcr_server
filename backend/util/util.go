package util

import (
	"fmt"

	"github.com/joho/godotenv"
)

// LoadEnv loads the environment variables from the .env file and returns an ErrorResponse if an error occurs

func LoadEnv() error {
	err := godotenv.Load()
	if err != nil {
		return fmt.Errorf("error loading .env file: %w", err)
	}
	return nil
}
