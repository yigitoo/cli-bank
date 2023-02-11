package main

import (
	"context"
	"fmt"
	"log"
	"os/exec"
	"runtime"

	definitions "github.com/yigitoo/cli-bank/util"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	clientOptions := options.Client().ApplyURI(definitions.DatabaseURL)
	client, err := mongo.Connect(context.TODO(), clientOptions)
	logerr(err)

	// check connection
	err = client.Ping(context.TODO(), nil)
	logerr(err)
	defer closeDB(client)

	information := client.Database("bank").Collection("information")
	var loggedUser definitions.User
	for {
		var cmd *exec.Cmd
		os := runtime.GOOS
		switch os {
		case "windows":
			cmd = exec.Command("cmd", "/c", "cls")
		case "darwin", "linux":
			cmd = exec.Command("bash", "-c", "\"clear\"")
		default:
			fmt.Printf("%s.\n", os)
		}

		if err := cmd.Run(); err != nil {
			log.Fatal(err)
		}

		isLogged, user := AuthOpt(information)
		loggedUser = user
		if isLogged {
			fmt.Printf("Logged in as: %v", user.Name)
			break
		} else {
			continue
		}
	}

	MoneyOperations(information, loggedUser)
}

func AuthOpt(collection *mongo.Collection) (bool, definitions.User) {
	var dbUser definitions.User

	fmt.Println(`
		
	`)
	err := collection.FindOne(context.TODO(), bson.D{{}}).Decode(&dbUser)
	logerr(err)

	user := definitions.User{}
	return true, user
}

func CreateUser(collection *mongo.Collection, user definitions.User) {
	_, err := collection.InsertOne(context.TODO(), user)
	logerr(err)
}

func LoginUser(name string, password string) {

}

func MoneyOperations(collection *mongo.Collection, user definitions.User) {
	for {
		fmt.Println(`
=== YapiKredi Bank (koc'a gitmek istiyorum da .d) ===
1. Para Yatir 
2. Para Cek
3. Parani Baska Bir Birime Donustur (TRY<=>EUR<=>USD)
4. Paranı Başkasına Gönder
Q. Çıkış!
		`)
		var request string
		fmt.Print("Girdi: ")
		fmt.Scanf("%s", &request)
		if request == "1" {
			incMoney(collection, user)
		} else if request == "2" {

		} else if request == "3" {

		} else if request == "4" {

		} else {
			break
		}
	}
}

func incMoney(collection *mongo.Collection, user definitions.User) {

}

func closeDB(client *mongo.Client) {
	err := client.Disconnect(context.TODO())
	logerr(err)
	fmt.Println("Connection closed MongoDB")
}
func logerr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
