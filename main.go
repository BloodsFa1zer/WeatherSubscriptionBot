package main

import (
	"encoding/json"
	"fmt"
	"github.com/caarlos0/env/v9"
	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
	"io"
	"net/http"
)

type Config struct {
	URL    string `env:"URL_WEATHER"`
	APIKey string `env:"API_KEY_WEATHER"`
	BotAPI string `env:"BOT_API_KEY"`
}

type WeatherData struct {
	Temperature struct {
		CurrentTemperature float64 `json:"temp"`
	} `json:"main"`
	WeatherDescription []struct {
		OverallDescription string `json:"description"`
	} `json:"weather"`
	CityName string `json:"name"`
	WindData struct {
		WindSpeed float64 `json:"speed"`
	} `json:"wind"`
}

var startMessage = "Hello! If you want to proceed, share <u>your location<u> with bot so it can provide you with" +
	" <b>weather data according to your region.<b>"
var unitsMessage = "Choose in which units you need to see weather info:"

func GetWeatherData(URL string, SpeedPerTime string) string {
	var WeatherInfo *WeatherData
	response, err := http.Get(URL)
	if err != nil {
		log.Fatal().Err(err).Msg(" Can`t read a response")
	}
	log.Info().Msg("Successfully read and return response")

	body, err := io.ReadAll(response.Body)
	if err != nil {
		log.Panic().Err(err).Msg(" can`t read data")
	}
	log.Info().Msg("successfully read data")

	err = json.Unmarshal(body, &WeatherInfo)
	if err != nil {
		log.Fatal().Err(err).Msg(" Can`t unmarshal data")
	}
	fmt.Println("weather info:", WeatherInfo)
	log.Info().Msg("Successfully unmarshal data and return it")
	var result []byte
	result = fmt.Appendf(result, "The weather in *%s* is *%.2f* and can be described as: *%s*. \n The <b>wind speed</b> is %.2f %s",
		WeatherInfo.CityName, WeatherInfo.Temperature.CurrentTemperature, WeatherInfo.WeatherDescription[0].OverallDescription, WeatherInfo.WindData.WindSpeed, SpeedPerTime)
	return string(result)
}

var units = map[string]string{
	"Fahrenheit":      "imperial",
	"Celsius":         "metric",
	"Kelvin(default)": "standard",
}

func main() {

	unitsKeyboard := make([]tgbotapi.KeyboardButton, len(units)+1)
	index := 0
	for key := range units {
		unitsKeyboard[index] = tgbotapi.NewKeyboardButton(key)
		index++
	}
	unitsKeyboard[index] = tgbotapi.NewKeyboardButton("Close")

	err := godotenv.Load("variables.env")
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

	locationButton := tgbotapi.KeyboardButton{
		Text:            "Click to share your location",
		RequestLocation: true,
	}

	latitude := 0.0
	longitude := 0.0

	updates := bot.GetUpdatesChan(updateConfig)
	for update := range updates {
		if update.Message == nil {
			log.Info().Msg("there are no commands from user")
			continue
		}

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
		//msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Hello, <b>this is bold text</b>, <i>this is italic text</i>, <a href=\"https://example.com\">this is a link</a>.")
		//
		//msg.ParseMode = tgbotapi.ModeHTML
		////msg.DisableWebPagePreview = true
		//
		//_, err = bot.Send(msg)
		//if err != nil {
		//	log.Fatal().Err(err)
		//}
		switch update.Message.Text {
		case "/start":
			//msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Hello! If you want to proceed, share <u>your location<u> with bot so it can provide you with\" +\n <b>weather data according to your region.<b>")
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, startMessage)
			msg.ReplyMarkup = tgbotapi.NewReplyKeyboard([]tgbotapi.KeyboardButton{locationButton})
			if _, err := bot.Send(msg); err != nil {
				log.Panic().Err(err).Msg(" Bot`s keyboard problem")
			}
			log.Info().Msg("User gets msg")
			break
		case "Close":
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Closing the reply keyboard...\n Thanks for using me! :)")
			msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
			if _, err = bot.Send(msg); err != nil {
				log.Panic().Err(err)
			}
			break
		}

		if update.Message.Location != nil {
			latitude = update.Message.Location.Latitude
			longitude = update.Message.Location.Longitude
			log.Info().Msg(" Successfully gets user location")
			msg = tgbotapi.NewMessage(update.Message.Chat.ID, unitsMessage)
			msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(unitsKeyboard)
			if _, err := bot.Send(msg); err != nil {
				log.Panic().Err(err).Msg(" Bot`s keyboard problem")
			}
			log.Info().Msg(" user gets keyboard to choose units")
		}
		if _, ok := units[update.Message.Text]; ok == true {
			urlWeatherAPI = fmt.Appendf(urlWeatherAPI, "lat=%f&lon=%f&appid=%s&units=%s", latitude, longitude, APIKey, units[update.Message.Text])
			result := ""
			if update.Message.Text == "Fahrenheit" {
				result = GetWeatherData(string(urlWeatherAPI), "miles/hour")
			} else {
				result = GetWeatherData(string(urlWeatherAPI), "meter/sec")
			}
			fmt.Println(string(urlWeatherAPI))
			bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, result))
			urlWeatherAPI = []byte(cfg.URL)
		}
	}
}
