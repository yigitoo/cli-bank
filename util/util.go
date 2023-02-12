package util

import (
	"context"
	"fmt" // for clearing screen
	"log"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	config "github.com/yigitoo/cli-bank/config"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	DatabaseURL string = getDatabaseUrl()
)

func GenerateUserID() string {
	id := uuid.New()
	return id.String()
}

type CurrencyQuantities struct {
	TRY float64 `bson:"TRY"`
	USD float64 `bson:"USD"`
	EUR float64 `bson:"EUR"`
}

type User struct {
	ID       primitive.ObjectID `bson:"_id"`
	Name     string
	Password string
	Balance  CurrencyQuantities
}

func getDatabaseUrl() string {
	if err := godotenv.Load(config.ProjectRootPath + "/.env"); err != nil {
		log.Fatal(err)
	}
	return os.Getenv("DB_LINK")
}

func CloseDB(client *mongo.Client) {
	err := client.Disconnect(context.TODO())
	if err != nil {
		log.Fatal(err)
	}
	println("Bizi tercih ettiginiz icin tesekkurler!")
	time.Sleep(2 * time.Second)
	ClearScreen()
}

func ClearScreen() {
	fmt.Printf("\x1bc")
}
