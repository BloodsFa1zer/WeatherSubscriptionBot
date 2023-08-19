package main

import (
	"example.com/mod/config"
	"example.com/mod/database"
	"example.com/mod/request"
	"example.com/mod/telegram"
	"fmt"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"time"
)

func main() {
	zerolog.TimeFieldFormat = time.TimeOnly

	cfg := config.LoadENV(".env")
	cfg.ParseENV()

	urlWeatherAPI := []byte(cfg.URL)
	bot := telegram.NewBotAPI(*cfg)

	APIKey := cfg.APIKey

	latitude := 0.0
	longitude := 0.0
	checkTime := false
	numberOfIterations := 0
	number := 0

	client := database.NewConnection(*cfg)

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

		foundUser := client.FindUser("UserID", userID)
		// isUserExist == true so the item with the same data exists
		// false if not

		if foundUser == nil && numberOfIterations == 0 {
			bot.SendMessage(chatID, "<b>You are already subscribed to the bot.</b>")
			bot.ReplySubscriptionMarkup(chatID, "You can wait for the next daily weather info or update your geolocation")
			numberOfIterations++
		}

		switch update.Message.Text {
		case "/start":
			if foundUser == nil {
				bot.ShowLocationKeyboard(chatID, telegram.StartMessage)
				log.Info().Msg("User gets msg")
			}
			break
		case "Close":
			bot.RemoveKeyboard(chatID, "Closing the reply keyboard...\n Thanks for using me! :)")
			numberOfIterations = 0
			break
		case "change your old location to the current one":
			bot.ShowLocationKeyboard(chatID, "Share your new location to update it or close if you don`t want to change it")
			log.Info().Msg("User gets msg")
			break
		}

		if update.Message.Location != nil {
			latitude = update.Message.Location.Latitude
			longitude = update.Message.Location.Longitude
			log.Info().Msg(" Successfully gets user location")
			bot.CreateUnitsKeyboard(chatID)

			log.Info().Msg(" user gets keyboard to choose units")

		}

		if unit, ok := request.Units[update.Message.Text]; ok {
			urlWeatherAPI = fmt.Appendf(urlWeatherAPI, "lat=%f&lon=%f&appid=%s&units=%s", latitude, longitude, APIKey, unit)
			bot.RemoveKeyboard(chatID, "Closing the reply keyboard.")
			bot.SendMessage(chatID, telegram.TimeMessage)
			checkTime = true
			number = 0
		} else if checkTime == true && bot.IsValidMessage(update.Message.Text) == true {
			timeResult, checkTimeFormat := bot.GetTimeFromString(update.Message.Text)
			if checkTimeFormat == true {
				res := timeResult.Format("15:04")
				WeatherResponse := string(urlWeatherAPI)
				weather := request.WeatherAPI{
					WeatherURL: WeatherResponse,
				}

				user := database.User{
					UserID:   userID,
					Link:     WeatherResponse,
					SendTime: res,
				}

				result := ""
				if foundUser != nil {
					client.CreateUser(user)
					bot.SendMessage(chatID, "Your location and info are <b>successfully added</b>")

				} else {
					client.UpdateUser(&foundUser.ID, user)
					bot.SendMessage(chatID, "Your location and info are <b>successfully updated</b>")
				}

				bot.RemoveKeyboard(chatID, "Wait for the next weather update")
				result = weather.RequestResult()
				fmt.Println(WeatherResponse)
				bot.SendMessage(chatID, result)
				go bot.InBackgroundMessage(chatID, weather, client)
				urlWeatherAPI = []byte(cfg.URL)
				select {}
			} else {
				bot.SendMessage(chatID, telegram.TimeMessage)
				continue
			}

		} else if number != 0 && checkTime == true {
			bot.SendMessage(chatID, telegram.TimeMessage)
		}

	}
}
