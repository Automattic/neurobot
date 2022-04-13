package config

import (
	"encoding/json"
	"errors"
	"os"
	"strconv"

	"github.com/apex/log"
	"github.com/joho/godotenv"
)

type config struct {
	Debug               bool
	WebhookListenerPort int
	DatabasePath        string
	HomeserverName      string
	PrimaryBotUsername  string
	PrimaryBotPassword  string
	WorkflowsTOMLPath   string
}

func LoadFromEnvFile(envPath string) *config {
	logger := log.WithFields(log.Fields{
		"envPath": envPath,
	})

	config, err := newConfig(envPath)
	if err != nil {
		logger.WithError(err).Fatal("Failed to load .env file")
	}

	configAsMap := config.Map()
	configAsMap["EnvPath"] = envPath
	configAsMap["PrimaryBotPassword"] = "******"
	logger.WithFields(log.Fields(configAsMap)).Info("Configuration loaded")

	return config
}

func newConfig(envPath string) (*config, error) {
	if err := godotenv.Load(envPath); err != nil {
		return nil, err
	}

	debug, _ := strconv.ParseBool(os.Getenv("DEBUG"))

	webhookListenerPort, err := strconv.Atoi(os.Getenv("WEBHOOK_LISTENER_PORT"))
	if err != nil {
		webhookListenerPort = 8080
	}

	config := &config{
		Debug:               debug,
		WebhookListenerPort: webhookListenerPort,
		DatabasePath:        os.Getenv("DB_FILE"),
		HomeserverName:      os.Getenv("MATRIX_SERVER_NAME"),
		PrimaryBotUsername:  os.Getenv("MATRIX_USERNAME"),
		PrimaryBotPassword:  os.Getenv("MATRIX_PASSWORD"),
		WorkflowsTOMLPath:   os.Getenv("WORKFLOWS_DEF_TOML_FILE"),
	}

	if err := config.validate(); err != nil {
		return nil, err
	}

	return config, nil
}

// Map returns the config as a map.
func (c config) Map() (values map[string]interface{}) {
	serialized, _ := json.Marshal(c)
	_ = json.Unmarshal(serialized, &values)
	return
}

func (c config) validate() error {
	if c.DatabasePath == "" {
		return errors.New("DB_FILE environment variable must be set and not empty")
	}

	if c.HomeserverName == "" {
		return errors.New("MATRIX_SERVER_NAME environment variable must be set and not empty")
	}

	if c.PrimaryBotUsername == "" {
		return errors.New("MATRIX_USERNAME environment variable must be set and not empty")
	}

	if c.PrimaryBotPassword == "" {
		return errors.New("MATRIX_PASSWORD environment variable must be set and not empty")
	}

	if c.WorkflowsTOMLPath == "" {
		return errors.New("WORKFLOWS_DEF_TOML_FILE environment variable must be set and not empty")
	}

	return nil
}
