package util

import (
	"log"
	"os"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	config "github.com/yigitoo/cli-bank/config"
)

func GenerateUserID() string {
	id := uuid.New()
	return id.String()
}

type CurrencyQuantities struct {
	TRY float64
	USD float64
	EUR float64
}

type User struct {
	Name     string
	Password string
	Balance  CurrencyQuantities
	ID       string
}

func CurrencyQuantitiesToList(currencies CurrencyQuantities) []float64 {
	list := make([]float64, 3)
	list = append(list, currencies.TRY)
	list = append(list, currencies.USD)
	list = append(list, currencies.EUR)
	return list
}

var (
	DatabaseURL string = getDatabaseUrl()
)

func getDatabaseUrl() string {
	if err := godotenv.Load(config.ProjectRootPath + "/.env"); err != nil {
		log.Fatal(err)
	}
	return os.Getenv("DB_LINK")
}
