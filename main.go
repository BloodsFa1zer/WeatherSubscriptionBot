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
	bot := botAPI{
		Key: BotInitialization(cfg),
	}

	APIKey := cfg.APIKey

	bot.Key.Debug = true
	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 60

	latitude := 0.0
	longitude := 0.0

	client := ClientConnection{collection: MongoDBConnection(cfg)}

	updates := bot.Key.GetUpdatesChan(updateConfig)
	numberOfIterations := 0
	existence := false
	for update := range updates {
		if update.Message == nil {
			log.Info().Msg("there are no commands from user")
			continue
		}

		ID := primitive.ObjectID{}

		userID := update.Message.From.ID
		chatID := update.Message.Chat.ID

		existence, ID = client.MongoDBFind("UserID", userID)
		// existence == true so the item with the same data exists
		// false if not

		if existence == true && numberOfIterations == 0 {
			bot.SendMessage(chatID, "<b>You are already subscribed to the bot.</b>")
			bot.ReplyKeyboardMarkup(chatID, subscriptionKeyboard, "You can wait for the next daily weather info or update your geolocation")
			numberOfIterations++
		}

		switch update.Message.Text {
		case "/start":
			if existence == false {
				bot.ReplyMarkup(chatID, startMessage, []tgbotapi.KeyboardButton{locationButton})
				log.Info().Msg("User gets msg")
			}
			break
		case "Close":
			bot.RemoveKeyboard(chatID, "Closing the reply keyboard...\n Thanks for using me! :)")
			numberOfIterations = 0
			break
		case "change your old location to the current one":
			bot.ReplyMarkup(chatID, "Share your new location to update it or close if you don`t want to change it", []tgbotapi.KeyboardButton{locationButton, closeButton})
			log.Info().Msg("User gets msg")
			break
		}

		if update.Message.Location != nil {
			latitude = update.Message.Location.Latitude
			longitude = update.Message.Location.Longitude
			log.Info().Msg(" Successfully gets user location")
			bot.ReplyKeyboardMarkup(chatID, tgbotapi.NewReplyKeyboard(unitsKeyboard), unitsMessage)

			log.Info().Msg(" user gets keyboard to choose units")
			//result := ""
			//result = GetWeatherData(string(urlWeatherAPI))
			//fmt.Println(string(urlWeatherAPI))
			//SendMessage(bot, chatID, result)
			//urlWeatherAPI = []byte(cfg.URL)
		}
		_, ok := units[update.Message.Text]
		if ok == true {
			urlWeatherAPI = fmt.Appendf(urlWeatherAPI, "lat=%f&lon=%f&appid=%s&units=%s", latitude, longitude, APIKey, units[update.Message.Text])
			user := User{
				UserID: userID,
				Link:   string(urlWeatherAPI),
			}
			if existence == false {
				client.MongoDBWrite(user)
				result := RequestResult(string(urlWeatherAPI))
				fmt.Println(string(urlWeatherAPI))
				bot.SendMessage(chatID, result)
				urlWeatherAPI = []byte(cfg.URL)
				bot.SendMessage(chatID, "Your location and info are <b>successfully added</b>")
				bot.RemoveKeyboard(chatID, "Wait for the next weather update")
			} else {
				client.MongoDBUpdate(ID, user)

				bot.SendMessage(chatID, "Your location and info are <b>successfully updated</b>")
				bot.RemoveKeyboard(chatID, "Wait for the next weather update")
			}
		}
	}
}
