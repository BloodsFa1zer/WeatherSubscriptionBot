package main

import (
	"fmt"
	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"time"
)

var timeFormat = "15:45"

func main() {
	//	fmt.Println(time.Now().Format("15:04"))
	zerolog.TimeFieldFormat = time.TimeOnly

	cfg := LoadENV(".env")
	cfg.ParseENV()

	urlWeatherAPI := []byte(cfg.URL)
	bot := NewBotAPI(*cfg)

	APIKey := cfg.APIKey

	latitude := 0.0
	longitude := 0.0
	checkTime := false

	client := NewMongoDBConnection(*cfg)

	numberOfIterations := 0
	//existence := false
	number := 0

	var lastUpdateID int

	for update := range bot.GetUpdates() {

		if update.Message == nil {
			log.Info().Msg("there are no commands from user")
			continue
		}
		if update.UpdateID > lastUpdateID {
			lastUpdateID = update.UpdateID
		}

		number = 1
		userID := update.Message.From.ID
		chatID := update.Message.Chat.ID

		existence, ID := client.MongoDBFind("UserID", userID)
		// existence == true so the item with the same data exists
		// false if not

		if existence == true && numberOfIterations == 0 {
			bot.SendMessage(chatID, "<b>You are already subscribed to the bot.</b>")
			bot.ReplyKeyboardMarkup(chatID, subscriptionKeyboard, "You can wait for the next daily weather info or update your geolocation")
			numberOfIterations++
		}

		switch update.Message.Text {
		case "/start":
			if existence != true {
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

		}

		_, ok := units[update.Message.Text]
		if ok == true {
			urlWeatherAPI = fmt.Appendf(urlWeatherAPI, "lat=%f&lon=%f&appid=%s&units=%s", latitude, longitude, APIKey, units[update.Message.Text])
			bot.RemoveKeyboard(chatID, "Closing the reply keyboard.")
			bot.SendMessage(chatID, timeMessage)
			checkTime = true
			number = 0
		}

		if checkTime == true && bot.isValidMessage(update.Message.Text) == true {
			timeResult, checkTimeFormat := bot.isValidTime(chatID, update.Message.Text)
			if checkTimeFormat == true {
				res := timeResult.Format("15:04")
				WeatherResponse := string(urlWeatherAPI)
				weather := WeatherAPI{
					WeatherURL: WeatherResponse,
				}

				user := User{
					UserID:   userID,
					Link:     WeatherResponse,
					SendTime: res,
				}

				result := ""
				if existence == false {
					client.MongoDBWrite(user)
					bot.SendMessage(chatID, "Your location and info are <b>successfully added</b>")

				} else {
					client.MongoDBUpdate(&ID, user)
					bot.SendMessage(chatID, "Your location and info are <b>successfully updated</b>")
				}

				bot.RemoveKeyboard(chatID, "Wait for the next weather update")
				result = weather.RequestResult()
				fmt.Println(WeatherResponse)
				bot.SendMessage(chatID, result)
				go bot.inBackgroundMessage(chatID, weather, client)
				urlWeatherAPI = []byte(cfg.URL)
				select {}
			} else {
				bot.SendMessage(chatID, timeMessage)
				continue
			}

		} else if number != 0 && checkTime == true {
			bot.SendMessage(chatID, timeMessage)
		}

	}
}
