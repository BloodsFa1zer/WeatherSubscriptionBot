package main

import (
	"encoding/json"
	"fmt"
	_ "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rs/zerolog/log"
	"io"
	"net/http"
)

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
	result = fmt.Appendf(result, "The weather in <b>%s</b> is <b>%.2f</b> and can be described as: <u>%s.</u> \n The wind speed is <b>%.2f %s</b>",
		WeatherInfo.CityName, WeatherInfo.Temperature.CurrentTemperature, WeatherInfo.WeatherDescription[0].OverallDescription, WeatherInfo.WindData.WindSpeed, SpeedPerTime)
	return string(result)
}
