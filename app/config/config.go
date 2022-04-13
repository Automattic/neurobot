package config

import (
	"errors"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type config struct {
	Debug        bool
	DatabasePath string
}

func New(envPath string) (*config, error) {
	if err := godotenv.Load(envPath); err != nil {
		return nil, err
	}

	debug, _ := strconv.ParseBool(os.Getenv("DEBUG"))

	config := &config{
		Debug:        debug,
		DatabasePath: os.Getenv("DB_FILE"),
	}

	if err := config.validate(); err != nil {
		return nil, err
	}

	return config, nil
}

func (c config) validate() error {
	if c.DatabasePath == "" {
		return errors.New("DB_FILE environment variable must be set and not empty")
	}

	return nil
}
