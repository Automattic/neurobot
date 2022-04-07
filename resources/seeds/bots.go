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
