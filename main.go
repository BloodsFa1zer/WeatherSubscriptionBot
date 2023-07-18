package main

import (
	"fmt"
	"github.com/caarlos0/env/v9"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type Config struct {
	URL    string `env:"URL_WEATHER"`
	APIKey string `env:"API_KEY_WEATHER"`
	URI_BD string `env:"URI_MongoDB"`
	BotAPI string `env:"BOT_API_KEY"`
}

var startMessage = "Hello! If you want to proceed, share <u>your location</u> with bot so it can provide you with" +
	" <b>weather data according to your region.</b>"

var unitsMessage = "Choose in which units you need to see weather info:"

var subscriptionKeyboard = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("change your old location to the current one"),
	),
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("Close"),
	),
)

var locationButton = tgbotapi.KeyboardButton{
	Text:            "Click to share your location",
	RequestLocation: true,
}

var closeButton = tgbotapi.KeyboardButton{
	Text: "Close",
}

func htmlFormat(chatID int64, text string) tgbotapi.MessageConfig {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = tgbotapi.ModeHTML
	return msg
}

var units = map[string]string{
	"Fahrenheit":      "imperial",
	"Celsius":         "metric",
	"Kelvin(default)": "standard",
}

type User struct {
	IDs    primitive.ObjectID `bson:"_id,omitempty"`
	UserID int64              `bson:"UserID,omitempty"`
	Link   string             `bson:"link,omitempty"`
}

func main() {
	zerolog.TimeFieldFormat = time.TimeOnly

	unitsKeyboard := make([]tgbotapi.KeyboardButton, len(units)+1)
	index := 0
	for key := range units {
		unitsKeyboard[index] = tgbotapi.NewKeyboardButton(key)
		index++
	}
	unitsKeyboard[index] = tgbotapi.NewKeyboardButton("Close")

	err := godotenv.Load(".env")
	if err != nil {
		log.Panic().Err(err).Msg(" does not load .env")
	}
	log.Info().Msg("successfully load .env")

	cfg := Config{}

	err = env.Parse(&cfg)
	if err != nil {
		log.Panic().Err(err).Msg(" unable to parse environment variables")
	}
	log.Info().Msg("successfully parsed .env")

	urlWeatherAPI := []byte(cfg.URL)
	bot, err := tgbotapi.NewBotAPI(cfg.BotAPI)
	if err != nil {
		log.Panic().Err(err).Msg(" Bot API problem")
	}
	APIKey := cfg.APIKey

	bot.Debug = true
	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 60

	latitude := 0.0
	longitude := 0.0

	usersCollection, err := MongoDBConnection(cfg)

	updates := bot.GetUpdatesChan(updateConfig)
	numberOfIterations := 0
	for update := range updates {
		if update.Message == nil {
			log.Info().Msg("there are no commands from user")
			continue
		}

		existence := false
		ID := primitive.ObjectID{}

		userID := update.Message.From.ID
		chatID := update.Message.Chat.ID

		existence, err, ID = MongoDBFind(usersCollection, "UserID", userID)

		// existence == true so the item with the same data exists
		// false if not

		msg := htmlFormat(chatID, "")

		if existence == true && numberOfIterations == 0 {
			msg = htmlFormat(chatID, "<b>You are already subscribed to the bot.</b>")
			bot.Send(msg)
			msg = htmlFormat(chatID, "You can wait for the next daily weather info or update your geolocation")
			msg.ReplyMarkup = subscriptionKeyboard
			numberOfIterations++
			if _, err := bot.Send(msg); err != nil {
				log.Panic().Err(err).Msg(" Bot`s keyboard problem")
			}
		}

		switch update.Message.Text {
		case "/start":
			msg := htmlFormat(chatID, startMessage)
			if existence == false {
				msg.ReplyMarkup = tgbotapi.NewReplyKeyboard([]tgbotapi.KeyboardButton{locationButton})
				if _, err := bot.Send(msg); err != nil {
					log.Panic().Err(err).Msg(" Bot`s keyboard problem")
				}
				log.Info().Msg("User gets msg")
			}
			break
		case "Close":
			msg = htmlFormat(chatID, "Closing the reply keyboard...\n Thanks for using me! :)")
			msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
			if _, err = bot.Send(msg); err != nil {
				log.Panic().Err(err)
			}
			break
		case "change your old location to the current one":
			msg := htmlFormat(chatID, "Share your new location to update it or close if you don`t want to change it")
			msg.ReplyMarkup = tgbotapi.NewReplyKeyboard([]tgbotapi.KeyboardButton{locationButton}, []tgbotapi.KeyboardButton{closeButton})
			if _, err := bot.Send(msg); err != nil {
				log.Panic().Err(err).Msg(" Bot`s keyboard problem")
			}
			log.Info().Msg("User gets msg")
			break

		}

		if update.Message.Location != nil {
			latitude = update.Message.Location.Latitude
			longitude = update.Message.Location.Longitude
			log.Info().Msg(" Successfully gets user location")
			msg = htmlFormat(chatID, unitsMessage)
			msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(unitsKeyboard)
			if _, err := bot.Send(msg); err != nil {
				log.Panic().Err(err).Msg(" Bot`s keyboard problem")
			}
			log.Info().Msg(" user gets keyboard to choose units")

			urlWeatherAPI = fmt.Appendf(urlWeatherAPI, "lat=%f&lon=%f&appid=%s&units=%s", latitude, longitude, APIKey, units[update.Message.Text])

			user := User{
				UserID: userID,
				Link:   string(urlWeatherAPI),
			}

			if existence == false {
				MongoDBWrite(usersCollection, user)
			} else {
				err = MongoDBUpdate(usersCollection, ID, user)
				if err != nil {
					panic(err)
				}
				msg := htmlFormat(chatID, "Your location is <b>successfully updated</b>")
				msg.ReplyMarkup = tgbotapi.ReplyKeyboardRemove{
					RemoveKeyboard: true,
				}
				if _, err := bot.Send(msg); err != nil {
					log.Panic().Err(err).Msg(" Bot`s keyboard problem")
				}
				log.Info().Msg("User gets msg")

			}
			result := ""
			result = GetWeatherData(string(urlWeatherAPI))
			fmt.Println(string(urlWeatherAPI))
			bot.Send(htmlFormat(chatID, result))
			urlWeatherAPI = []byte(cfg.URL)
		}
		//ticker := time.NewTicker(24 * time.Hour)

	}
	// select {}
}
