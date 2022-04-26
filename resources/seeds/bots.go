package seeds

import (
	"fmt"
	"log"
	configuration "neurobot/app/config"
	"neurobot/model/bot"
	"os"
	"strings"
)

func Bots(repository bot.Repository, config *configuration.Config) {
	seeds := []bot.Bot{
		makePrimaryBot("Primary bot", config), // MUST be the first to be created so that ID = 1
		makeBot("afkbot", "Used by afk_notifier and !afk"),
		makeBot("messengerbot", "Used by a8c_matrix()"),
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

func makePrimaryBot(description string, config *configuration.Config) bot.Bot {
	return bot.Bot{
		Description: description,
		Username:    config.PrimaryBotUsername,
		Password:    config.PrimaryBotPassword,
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
