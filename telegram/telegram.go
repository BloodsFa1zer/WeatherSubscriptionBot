package telegram

import (
	"example.com/mod/config"
	"example.com/mod/database"
	"example.com/mod/request"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rs/zerolog/log"
	"regexp"
	"time"
)

var StartMessage = "Hello! If you want to proceed, share <u>your location</u> with bot so it can provide you with" +
	" <b>weather data according to your region.</b>"

var UnitsMessage = "Choose in which units you need to see weather info:"

var TimeMessage = "You need to enter time when the daily weather update will be sent, that is correspond with given example: 15:45"

var subscriptionKeyboard = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("change your old location to the current one"),
	),
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("Close"),
	),
)

var closeButton = tgbotapi.KeyboardButton{
	Text: "Close",
}

var locationButton = tgbotapi.KeyboardButton{
	Text:            "Click to share your location",
	RequestLocation: true,
}

type BotAPI struct {
	Key *tgbotapi.BotAPI
}

func NewBotAPI(config config.Config) *BotAPI {
	key, err := tgbotapi.NewBotAPI(config.BotAPI)
	if err != nil {
		log.Panic().Err(err).Msg(" Bot API problem")
	}
	bot := BotAPI{
		Key: key,
	}
	return &bot
}

func (bot *BotAPI) SendMessage(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = tgbotapi.ModeHTML
	if _, err := bot.Key.Send(msg); err != nil {
		log.Panic().Err(err).Msg(" Bot can`t send Message")
	}
	log.Info().Msg("Bot send Message")
}

func (bot *BotAPI) ReplyMarkup(chatID int64, text string, button ...[]tgbotapi.KeyboardButton) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = tgbotapi.ModeHTML
	msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(button...)
	if _, err := bot.Key.Send(msg); err != nil {
		log.Panic().Err(err).Msg(" Bot can`t send ReplyMarkup")
	}
	log.Info().Msg("Bot send ReplyMarkup")
}

func (bot *BotAPI) RemoveKeyboard(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = tgbotapi.ModeHTML
	msg.ReplyMarkup = tgbotapi.ReplyKeyboardRemove{
		RemoveKeyboard: true,
	}
	if _, err := bot.Key.Send(msg); err != nil {
		log.Panic().Err(err).Msg(" Bot can`t send RemoveKeyboard")
	}
	log.Info().Msg("Bot send RemoveKeyboard")

}

func (bot *BotAPI) ReplyKeyboardMarkup(chatID int64, markup tgbotapi.ReplyKeyboardMarkup, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = markup
	if _, err := bot.Key.Send(msg); err != nil {
		log.Panic().Err(err).Msg(" Bot can`t send ReplyKeyboardMarkup")
	}
	log.Info().Msg("Bot send ReplyKeyboardMarkup")

}

func (bot *BotAPI) createKeyboard(chatID int64, length int, Map map[string]string, text string) {
	unitsKeyboard := make([]tgbotapi.KeyboardButton, length)
	index := 0
	for key := range Map {
		unitsKeyboard[index] = tgbotapi.NewKeyboardButton(key)
		index++
	}
	unitsKeyboard[index] = tgbotapi.NewKeyboardButton("Close")

	bot.ReplyMarkup(chatID, text, unitsKeyboard)

}

func (bot *BotAPI) ReplySubscriptionMarkup(chatID int64, text string) {
	bot.ReplyKeyboardMarkup(chatID, subscriptionKeyboard, text)
}

func (bot *BotAPI) CreateUnitsKeyboard(chatID int64) {
	bot.createKeyboard(chatID, len(request.Units)+1, request.Units, UnitsMessage)
}

func (bot *BotAPI) GetUpdates() <-chan tgbotapi.Update {
	bot.Key.Debug = true
	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 60

	updates := bot.Key.GetUpdatesChan(updateConfig)

	return updates
}

func (bot *BotAPI) ShowLocationKeyboard(chatID int64, text string) {
	bot.ReplyMarkup(chatID, text, []tgbotapi.KeyboardButton{locationButton, closeButton})
}

func (bot *BotAPI) InBackgroundMessage(chatID int64, result request.WeatherAPI, client *database.ClientConnection) {
	ticker := time.NewTicker(1 * time.Minute)

	go func() {
		for {
			select {
			case <-ticker.C:
				foundUser := client.FindUser("time", time.Now().Format("15:04"))
				if foundUser != nil {
					res := result.RequestResult()
					bot.SendMessage(chatID, res)
				}
			}
		}
	}()
}

func (bot *BotAPI) IsValidMessage(message string) bool {
	timePattern := regexp.MustCompile(`^\d{2}:\d{2}$`)
	return timePattern.MatchString(message)

}

func (bot *BotAPI) GetTimeFromString(message string) (time.Time, bool) {

	timeResult, err := time.Parse("15:04", message)
	if err != nil {
		return time.Time{}, false
	}
	return timeResult, true
}
