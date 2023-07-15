package main

import (
	"context"
	"fmt"
	"github.com/caarlos0/env/v9"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
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
var locationMessage = "Choose option that is more suitable in your certain case."

var locationKeyboard = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("that is your constant location\n(can change it later)"),
	),
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("temporary location\n(will need to change it to the constant one)"),
	),
)

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

func main() {

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

	client, err := mongo.NewClient(options.Client().ApplyURI(cfg.URI_BD))
	if err != nil {
		log.Fatal().Err(err).Msg("can`t add localhost")
	}
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal().Err(err)
	}
	defer client.Disconnect(ctx)

	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		log.Fatal().Err(err)
	}

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

		msg := htmlFormat(update.Message.Chat.ID, "")

		switch update.Message.Text {
		case "/start":
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, startMessage)
			msg.ParseMode = tgbotapi.ModeHTML
			msg.ReplyMarkup = tgbotapi.NewReplyKeyboard([]tgbotapi.KeyboardButton{locationButton})
			if _, err := bot.Send(msg); err != nil {
				log.Panic().Err(err).Msg(" Bot`s keyboard problem")
			}
			log.Info().Msg("User gets msg")
			break
		case "Close":
			msg = tgbotapi.NewMessage(update.Message.Chat.ID, "Closing the reply keyboard...\n Thanks for using me! :)")
			msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
			if _, err = bot.Send(msg); err != nil {
				log.Panic().Err(err)
			}
			break
		}

		if update.Message.Location != nil {
			msg = tgbotapi.NewMessage(update.Message.Chat.ID, locationMessage)
			msg.ReplyMarkup = locationKeyboard
			if _, err := bot.Send(msg); err != nil {
				log.Panic().Err(err).Msg(" Bot`s keyboard problem")
			}
			log.Info().Msg(" user gets keyboard to choose units")
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
			bot.Send(htmlFormat(update.Message.Chat.ID, result))
			urlWeatherAPI = []byte(cfg.URL)
		}

	}
}
