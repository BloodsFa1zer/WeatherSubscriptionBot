package main

import (
	"context"
	"fmt"
	"github.com/caarlos0/env/v9"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/bson"
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

var subscriptionKeyboard = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("continue with your location"),
	),
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("change your old location to the current one"),
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

type Book struct {
	Title  string
	Author string
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

	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(cfg.URI_BD))
	if err != nil {
		panic(err)
	}

	if err := client.Ping(context.TODO(), readpref.Primary()); err != nil {
		panic(err)
	}

	usersCollection := client.Database("telegram").Collection("usersID")

	//user := bson.D{{"fullName", "User 1"}, {"age", 30}}
	//
	//result, err := usersCollection.InsertOne(context.TODO(), user)
	//if err != nil {
	//	panic(err)
	//}
	//
	//users := []interface{}{
	//	bson.D{{"fullName", "User 2"}, {"age", 25}},
	//	bson.D{{"fullName", "User 3"}, {"age", 20}},
	//	bson.D{{"fullName", "User 4"}, {"age", 28}},
	//}
	//// insert the bson object slice using InsertMany()
	//results, err := usersCollection.InsertMany(context.TODO(), users)
	//// check for errors in the insertion
	//if err != nil {
	//	panic(err)
	//}
	//
	//// display the ids of the newly inserted objects
	//fmt.Println(result.InsertedID)
	//fmt.Println(results.InsertedIDs)

	//filter := bson.D{
	//	{"$and",
	//		bson.A{
	//			bson.D{
	//				{"age", bson.D{{"$gt", 25}}},
	//			},
	//		},
	//	},
	//}

	//cursor, err := usersCollection.Find(context.TODO(), bson.D{})
	//if err != nil {
	//	panic(err)
	//}

	//var results []bson.M
	//if err = cursor.All(context.TODO(), &results); err != nil {
	//	log.Fatal().Err(err)
	//}
	//for _, result := range results {
	//	fmt.Println(result)
	//}

	check := false

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
	closeButton := tgbotapi.KeyboardButton{
		Text: "Close",
	}

	latitude := 0.0
	longitude := 0.0

	updates := bot.GetUpdatesChan(updateConfig)
	for update := range updates {
		if update.Message == nil {
			log.Info().Msg("there are no commands from user")
			continue
		}

		userID := update.Message.From.ID
		chatID := update.Message.Chat.ID
		opts := options.FindOne().SetSort(bson.D{{}})
		ctx := usersCollection.FindOne(context.TODO(), bson.D{{"$eq", userID}}, opts)
		if ctx != nil {
			check = true // if check = true user with the same userID is already stored in the DB
		}

		msg := htmlFormat(chatID, "")

		switch update.Message.Text {
		case "/start":
			msg := htmlFormat(chatID, startMessage)
			msg.ReplyMarkup = tgbotapi.NewReplyKeyboard([]tgbotapi.KeyboardButton{locationButton})
			if _, err := bot.Send(msg); err != nil {
				log.Panic().Err(err).Msg(" Bot`s keyboard problem")
			}
			log.Info().Msg("User gets msg")
			break
		case "Close":
			msg = tgbotapi.NewMessage(chatID, "Closing the reply keyboard...\n Thanks for using me! :)")
			msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
			if _, err = bot.Send(msg); err != nil {
				log.Panic().Err(err)
			}
			break
		case "change your old location to the current one":
			msg := htmlFormat(chatID, "If you need to update your location, choose suitable option or close keyboard")
			usersCollection.DeleteOne(context.TODO(), bson.D{{"userID", userID}})
			msg.ReplyMarkup = tgbotapi.NewReplyKeyboard([]tgbotapi.KeyboardButton{locationButton}, []tgbotapi.KeyboardButton{closeButton})
			if _, err := bot.Send(msg); err != nil {
				log.Panic().Err(err).Msg(" Bot`s keyboard problem")
			}
			log.Info().Msg("User gets msg")
			break
		}

		if update.Message.Location != nil {
			msg = tgbotapi.NewMessage(chatID, locationMessage)
			msg.ReplyMarkup = locationKeyboard
			if _, err := bot.Send(msg); err != nil {
				log.Panic().Err(err).Msg(" Bot`s keyboard problem")
			}
			log.Info().Msg(" user gets keyboard to choose units")
			latitude = update.Message.Location.Latitude
			longitude = update.Message.Location.Longitude
			log.Info().Msg(" Successfully gets user location")
			msg = tgbotapi.NewMessage(chatID, unitsMessage)
			msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(unitsKeyboard)
			if _, err := bot.Send(msg); err != nil {
				log.Panic().Err(err).Msg(" Bot`s keyboard problem")
			}
			log.Info().Msg(" user gets keyboard to choose units")
		}

		if _, ok := units[update.Message.Text]; ok == true {
			urlWeatherAPI = fmt.Appendf(urlWeatherAPI, "lat=%f&lon=%f&appid=%s&units=%s", latitude, longitude, APIKey, units[update.Message.Text])
			if check != true {
				user := bson.D{{"UserID", userID}, {"link", string(urlWeatherAPI)}}
				_, err := usersCollection.InsertOne(context.TODO(), user)
				if err != nil {
					log.Panic().Err(err).Msg(" can`t insert user`s data into database")
				}
				log.Info().Msg("successfully insert user`s data")
			} else {
				msg = htmlFormat(chatID, "<b>You are already subscribed to the bot.</b>")
				msg.ReplyMarkup = subscriptionKeyboard
				if _, err := bot.Send(msg); err != nil {
					log.Panic().Err(err).Msg(" Bot`s keyboard problem")
				}
				log.Info().Msg(" user gets keyboard to choose")
			}
			result := ""

			if update.Message.Text == "Fahrenheit" {
				result = GetWeatherData(string(urlWeatherAPI), "miles/hour")
			} else {
				result = GetWeatherData(string(urlWeatherAPI), "meter/sec")
			}
			fmt.Println(string(urlWeatherAPI))
			bot.Send(htmlFormat(chatID, result))
			urlWeatherAPI = []byte(cfg.URL)
		}

	}
}
