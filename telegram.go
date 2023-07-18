package main

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rs/zerolog/log"
)

func BotInitialization(config Config) *tgbotapi.BotAPI {
	bot, err := tgbotapi.NewBotAPI(config.BotAPI)
	if err != nil {
		log.Panic().Err(err).Msg(" Bot API problem")
	}
	log.Info().Msg("Bot API works correctly")
	return bot
}

func SendMessage(bot *tgbotapi.BotAPI, chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = tgbotapi.ModeHTML
	if _, err := bot.Send(msg); err != nil {
		log.Panic().Err(err).Msg(" Bot can`t send Message")
	}
	log.Info().Msg("Bot send Message")
}

func ReplyMarkup(bot *tgbotapi.BotAPI, chatID int64, text string, button ...[]tgbotapi.KeyboardButton) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = tgbotapi.ModeHTML
	msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(button...)
	if _, err := bot.Send(msg); err != nil {
		log.Panic().Err(err).Msg(" Bot can`t send ReplyMarkup")
	}
	log.Info().Msg("Bot send ReplyMarkup")
}

func RemoveKeyboard(bot *tgbotapi.BotAPI, chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = tgbotapi.ModeHTML
	msg.ReplyMarkup = tgbotapi.ReplyKeyboardRemove{
		RemoveKeyboard: true,
	}
	if _, err := bot.Send(msg); err != nil {
		log.Panic().Err(err).Msg(" Bot can`t send RemoveKeyboard")
	}
	log.Info().Msg("Bot send RemoveKeyboard")

}

func ReplyKeyboardMarkup(bot *tgbotapi.BotAPI, chatID int64, markup tgbotapi.ReplyKeyboardMarkup, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = markup
	if _, err := bot.Send(msg); err != nil {
		log.Panic().Err(err).Msg(" Bot can`t send ReplyKeyboardMarkup")
	}
	log.Info().Msg("Bot send ReplyKeyboardMarkup")

}
