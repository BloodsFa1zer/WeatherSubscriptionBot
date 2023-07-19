package main

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rs/zerolog/log"
)

type botAPI struct {
	Key *tgbotapi.BotAPI
}

func BotInitialization(config Config) *tgbotapi.BotAPI {

	key, err := tgbotapi.NewBotAPI(config.BotAPI)
	if err != nil {
		log.Panic().Err(err).Msg(" Bot API problem")
	}
	bot := botAPI{Key: key}

	log.Info().Msg("Bot API works correctly")
	return bot.Key
}

func (bot *botAPI) SendMessage(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = tgbotapi.ModeHTML
	if _, err := bot.Key.Send(msg); err != nil {
		log.Panic().Err(err).Msg(" Bot can`t send Message")
	}
	log.Info().Msg("Bot send Message")
}

func (bot *botAPI) ReplyMarkup(chatID int64, text string, button ...[]tgbotapi.KeyboardButton) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = tgbotapi.ModeHTML
	msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(button...)
	if _, err := bot.Key.Send(msg); err != nil {
		log.Panic().Err(err).Msg(" Bot can`t send ReplyMarkup")
	}
	log.Info().Msg("Bot send ReplyMarkup")
}

func (bot *botAPI) RemoveKeyboard(chatID int64, text string) {
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

func (bot *botAPI) ReplyKeyboardMarkup(chatID int64, markup tgbotapi.ReplyKeyboardMarkup, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = markup
	if _, err := bot.Key.Send(msg); err != nil {
		log.Panic().Err(err).Msg(" Bot can`t send ReplyKeyboardMarkup")
	}
	log.Info().Msg("Bot send ReplyKeyboardMarkup")

}
