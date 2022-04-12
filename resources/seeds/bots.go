package seeds

import (
	"fmt"
	"log"
	"neurobot/model/bot"
	"os"
	"strings"
)

func Bots(repository bot.Repository) {
	seeds := []bot.Bot{
		makePrimaryBot("Primary bot"), // MUST be the first to be created so that ID = 1
		makeBot("afkbot", "Used by afk_notifier and !afk"),
	}

	for _, seed := range seeds {
		existing, _ := repository.FindByUsername(seed.Username)
		if existing.ID > 0 {
			// Bot already exists, we'll update it.
			seed.ID = existing.ID
		}

		err := repository.Save(&seed)
		if err != nil {
			log.Fatalf("Failed to seed bots: %s", err)
		}
	}
}

func makePrimaryBot(description string) bot.Bot {
	username, exists := os.LookupEnv("MATRIX_USERNAME")
	if !exists || username == "" {
		log.Fatalf("environment variable MATRIX_USERNAME not set")
	}

	password, exists := os.LookupEnv("MATRIX_PASSWORD")
	if !exists || password == "" {
		log.Fatalf("environment variable MATRIX_PASSWORD not set")
	}

	return bot.Bot{
		Description: description,
		Username:    username,
		Password:    password,
		Active:      true,
	}
}

func makeBot(username string, description string) bot.Bot {
	// The env variable is called, for example, AFKBOT_PASSWORD.
	passwordEnvName := fmt.Sprintf("%s_PASSWORD", strings.ToUpper(username))

	passwordEnvValue, exists := os.LookupEnv(passwordEnvName)
	if !exists || passwordEnvValue == "" {
		log.Fatalf("environment variable %s is not set", passwordEnvName)
	}

	return bot.Bot{
		Description: description,
		Username:    username,
		Password:    passwordEnvValue,
		Active:      true,
	}
}
