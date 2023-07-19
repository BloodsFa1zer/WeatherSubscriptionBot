package main

import (
	"fmt"
	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"time"
)

func main() {
	zerolog.TimeFieldFormat = time.TimeOnly

	cfg := LoadENV(".env")
	cfg.ParseENV()

	urlWeatherAPI := []byte(cfg.URL)
	bot := NewBotAPI(*cfg)

	APIKey := cfg.APIKey
	var url WeatherURL

	bot.Key.Debug = true
	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 60

	latitude := 0.0
	longitude := 0.0

	client := NewMongoDBConnection(*cfg)

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
			bot.CreateKeyboard(chatID, len(units)+1, units, unitsMessage)

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

				url = WeatherURL(urlWeatherAPI)
				result := url.RequestResult()
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
