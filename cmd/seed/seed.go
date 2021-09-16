package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/rs/zerolog/log"
	minervaLog "github.com/sy-software/minerva-go-utils/log"
	"github.com/sy-software/minerva-olive/internal/core/domain"
	"github.com/sy-software/minerva-olive/internal/core/repositories/redis"
	"github.com/sy-software/minerva-olive/internal/core/service"
	"github.com/sy-software/minerva-olive/mocks"
)

const defaultConfigFile = "./config.json"

func main() {
	rand.Seed(time.Now().UnixNano())
	minervaLog.ConfigureLogger(minervaLog.LogLevel(os.Getenv("LOG_LEVEL")), os.Getenv("CONSOLE_OUTPUT") != "")
	config := domain.LoadConfig()
	db, err := redis.GetRedisDB(&config)
	if err != nil {
		log.Error().Stack().Err(err).Msg("Can't initialize Redis DB")
		os.Exit(1)
	}
	repo := redis.NewRedisRepo(&config, db)
	if err != nil {
		log.Error().Stack().Err(err).Msg("Can't initialize Redis DB")
		os.Exit(1)
	}
	configService := service.NewConfigService(&config, repo, repo, &mocks.MockSecrets{})

	for set := 0; set < 10; set++ {
		setName := fmt.Sprintf("sample%d", set)
		configService.CreateSet(setName)
		for key := 0; key < rand.Intn(50)+1; key++ {
			configService.AddItem(*domain.NewConfigItem(
				fmt.Sprintf("key%d", key),
				randValue(),
				domain.Plain,
			), setName)
		}
	}
}

func randValue() interface{} {
	types := []string{
		"number",
		"bool",
		"string",
	}

	t := types[rand.Intn(len(types))]
	switch t {
	case "number":
		return rand.Int()
	case "bool":
		return rand.Intn(100) > 50
	case "string":
		return randomString(rand.Intn(1024))
	default:
		return false
	}
}

func randomString(n int) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	s := make([]rune, n)
	for i := range s {
		s[i] = letters[rand.Intn(len(letters))]
	}
	return string(s)
}
