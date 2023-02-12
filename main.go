package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"reflect"
	"strings"
	"sync"
	"time"

	definitions "github.com/yigitoo/cli-bank/util"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	wg                   sync.WaitGroup
	session_user         primitive.ObjectID
	supported_currencies = []string{"TRY", "EUR", "USD"}
	floatType            = reflect.TypeOf(float64(0))
)

func main() {
	clientOptions := options.Client().ApplyURI(definitions.DatabaseURL)
	client, err := mongo.Connect(context.TODO(), clientOptions)
	logerr(err)

	// check connection
	err = client.Ping(context.TODO(), nil)
	logerr(err)
	defer definitions.CloseDB(client)

	information := client.Database("bank").Collection("information")
	var loggedUser definitions.User

	for {
		definitions.ClearScreen()

		isLogged, user := AuthOpt(information)
		session_user = user.ID
		loggedUser = user
		if isLogged {
			fmt.Printf("%v olarak giris yapildi!", user.Name)
			time.Sleep(2 * time.Second)
			break
		} else {
			time.Sleep(2 * time.Second)
			CreateUser(information)
		}
	}

	definitions.ClearScreen()
	MoneyOperations(information, loggedUser)
}

func AuthOpt(collection *mongo.Collection) (bool, definitions.User) {
	var dbUser definitions.User
	var username, password string

	definitions.ClearScreen()

	fmt.Println(`=== GIRIS YAP ===`)
	fmt.Print("Kullanici adi: ")
	fmt.Scanln(&username)
	fmt.Print("Sifre: ")
	fmt.Scanln(&password)
	wg.Add(1)
	for {
		if password != "" {
			wg.Done()
			break
		}
	}
	err := collection.FindOne(context.TODO(), bson.D{
		{"name", username},
		{"password", password},
	}).Decode(&dbUser)

	if err != nil {
		return false, definitions.User{
			Name:     "temp",
			Password: "temp",
			Balance: definitions.CurrencyQuantities{
				TRY: 0,
				EUR: 0,
				USD: 0,
			},
			ID: primitive.NewObjectID(),
		}
	}

	return true, dbUser
}

func CreateUser(collection *mongo.Collection) {
	definitions.ClearScreen()

	var username, password string

	fmt.Println(`
=== KAYIT OL ===`)

	fmt.Print("Kullanici adin: ")
	fmt.Scanln(&username)
	fmt.Print("Belirledigin sifre: ")
	fmt.Scanln(&password)
	wg.Add(1)
	for {
		if password != "" {
			wg.Done()
			break
		}
	}

	user := definitions.User{
		ID:       primitive.NewObjectID(),
		Name:     username,
		Password: password,
		Balance: definitions.CurrencyQuantities{
			TRY: 0.0,
			EUR: 0.0,
			USD: 0.0,
		},
	}

	_, err := collection.InsertOne(context.TODO(), user)
	logerr(err)
}

func MoneyOperations(collection *mongo.Collection, user definitions.User) {
	for {
		user = fetchUser(collection, session_user)
		definitions.ClearScreen()

		fmt.Println(`
=== YapiKredi Bank (koc'a gitmek istiyorum da .d) ===
1. Para Yatir 
2. Para Cek
3. Parani Baska Bir Birime Donustur (TRY<=>EUR<=>USD)
4. Paranı Başkasına Gönder
Q. Çıkış!`)
		fmt.Printf(`
Bakiye: %f TRY, %f USD, %f EUR`, user.Balance.TRY, user.Balance.USD, user.Balance.EUR)
		var request string
		fmt.Print("Girdi: ")
		fmt.Scanln(&request)

		if request == "1" {
			changeMoney(collection, user, "+")
		} else if request == "2" {
			changeMoney(collection, user, "-")
		} else if request == "3" {
			convertMoney(collection, user)
		} else if request == "4" {
			SendForwardById(collection, user)
		} else if strings.ToUpper(request) == "Q" {
			break
		} else {
			continue
		}
	}
}

func changeMoney(collection *mongo.Collection, user definitions.User, pos_or_neg string) {
	var quantity float64
	var currency string

	var run bool = true
	for {
		fmt.Print("(TRY, EUR, USD) Hangi birimde yatıracaksın?: ")
		fmt.Scanln(&currency)
		for _, iter := range supported_currencies {
			if strings.ToUpper(currency) == iter {
				run = false
			}
		}

		if !run {
			break
		}
	}

	fmt.Println("Tamam simdi miktarini gir.")

	if pos_or_neg == "+" {
		fmt.Print("\nNe kadar yatırmak istiyorsun?: ")
		fmt.Scan(&quantity)
	} else {
		fmt.Print("Ne kadar cekmek istiyorsun?: ")
		fmt.Scan(&quantity)
		quantity = -quantity
	}
	time.Sleep(500 * time.Millisecond)
	var newBalance definitions.CurrencyQuantities
	if strings.ToUpper(currency) == "TRY" {
		newBalance = definitions.CurrencyQuantities{
			TRY: user.Balance.TRY + quantity,
			EUR: user.Balance.EUR,
			USD: user.Balance.USD,
		}
		fmt.Println(newBalance)
	}
	if strings.ToUpper(currency) == "USD" {
		newBalance = definitions.CurrencyQuantities{
			TRY: user.Balance.TRY,
			EUR: user.Balance.EUR,
			USD: user.Balance.USD + quantity,
		}
		fmt.Println(newBalance)
	}
	if strings.ToUpper(currency) == "EUR" {
		newBalance = definitions.CurrencyQuantities{
			TRY: user.Balance.TRY,
			EUR: user.Balance.EUR + quantity,
			USD: user.Balance.USD,
		}
		fmt.Println(newBalance)
	}

	filter := bson.D{{"_id", session_user}}
	update := bson.D{{"$set", bson.M{
		"balance": bson.M{
			"TRY": newBalance.TRY,
			"USD": newBalance.USD,
			"EUR": newBalance.EUR,
		},
	}}}

	_, err := collection.UpdateOne(context.TODO(), filter, update)
	logerr(err)

	fmt.Print("Ok!👌")
	for i := 0; i < 3; i++ {
		fmt.Print(".")
		time.Sleep(500 * time.Millisecond)
	}
	time.Sleep(500 * time.Millisecond)
}

func convertMoney(collection *mongo.Collection, user definitions.User) {
	var toBoConverted, willBeConverted string
	var quantity float64

	fmt.Print("Cevirmek istedigin birimi gir: ")
	fmt.Scanln(&toBoConverted)

	fmt.Print("Bunu neye cevirmek istersin?: ")
	fmt.Scanln(&willBeConverted)

	fmt.Print("Ne kadarını cevireceksin?: ")
	fmt.Scanln(&quantity)

	exchange_rate := check_exchange_rates(strings.ToUpper(toBoConverted), strings.ToUpper(willBeConverted))

	newBalanceMap := map[string]float64{
		"TRY": user.Balance.TRY,
		"USD": user.Balance.USD,
		"EUR": user.Balance.EUR,
	}

	exchange_rate_converted, err := interface_to_float(exchange_rate)
	logerr(err)

	newBalanceMap[strings.ToUpper(toBoConverted)] -= quantity
	newBalanceMap[strings.ToUpper(willBeConverted)] += quantity * exchange_rate_converted

	filter := bson.D{{"_id", session_user}}
	update := bson.D{{"$set", bson.M{
		"balance": bson.M{
			"TRY": newBalanceMap["TRY"],
			"USD": newBalanceMap["USD"],
			"EUR": newBalanceMap["EUR"],
		},
	}}}

	_, err = collection.UpdateOne(context.TODO(), filter, update)
	logerr(err)

}

func check_exchange_rates(base string, target string) interface{} {
	var req_url string = "https://api.exchangerate.host/convert?from=" + base + "&to=" + target
	req, err := http.Get(req_url)
	logerr(err)
	defer req.Body.Close()
	body, _ := io.ReadAll(req.Body)

	var res map[string]interface{}
	err = json.Unmarshal(body, &res)
	logerr(err)

	return res["result"]
}

func SendForwardById(collection *mongo.Collection, user definitions.User) {
	var forwardedIDString, currency string
	var quantity float64
	fmt.Print("\n=== Send money via IBAN === (iban at yegen .d)\n")
	fmt.Print("IBAN'ını gir: ")
	fmt.Scanln(&forwardedIDString)

	forwardedID, err := primitive.ObjectIDFromHex(forwardedIDString)
	logerr(err)

	var forwarededUser definitions.User
	err = collection.FindOne(context.TODO(), bson.D{{"_id", forwardedID}}).Decode(&forwarededUser)
	if err != nil {
		println("Kullanıcı bulunamadı.")
		time.Sleep(2 * time.Second)
	}

	for {
		fmt.Print("Para birmini gir: ")
		fmt.Scanln(&currency)

		currency = strings.ToUpper(currency)

		if currency == "TRY" || currency == "USD" || currency == "EUR" {
			break
		}
	}

	fmt.Print("Miktar gir: ")
	fmt.Scanln(&quantity)

	newBalanceMap1 := map[string]float64{
		"TRY": user.Balance.TRY,
		"USD": user.Balance.USD,
		"EUR": user.Balance.EUR,
	}
	newBalanceMap1[currency] -= quantity

	filter := bson.D{{"_id", session_user}}
	update := bson.D{{"$set", bson.M{
		"balance": bson.M{
			"TRY": newBalanceMap1["TRY"],
			"USD": newBalanceMap1["USD"],
			"EUR": newBalanceMap1["EUR"],
		},
	}}}

	_, err = collection.UpdateOne(context.TODO(), filter, update)
	logerr(err)

	newBalanceMap2 := map[string]float64{
		"TRY": forwarededUser.Balance.TRY,
		"USD": forwarededUser.Balance.USD,
		"EUR": forwarededUser.Balance.EUR,
	}
	newBalanceMap2[currency] += quantity

	filter = bson.D{{"_id", forwardedID}}
	update = bson.D{{"$set", bson.M{
		"balance": bson.M{
			"TRY": newBalanceMap2["TRY"],
			"USD": newBalanceMap2["USD"],
			"EUR": newBalanceMap2["EUR"],
		},
	}}}

	_, err = collection.UpdateOne(context.TODO(), filter, update)
	logerr(err)

}

func fetchUser(collection *mongo.Collection, session_user primitive.ObjectID) definitions.User {
	var dbUser definitions.User
	err := collection.FindOne(context.TODO(), bson.D{{"_id", session_user}}).Decode(&dbUser)
	logerr(err)

	return dbUser
}

func interface_to_float(unk interface{}) (float64, error) {
	v := reflect.ValueOf(unk)
	v = reflect.Indirect(v)
	if !v.Type().ConvertibleTo(floatType) {
		return 0, fmt.Errorf("cannot convert %v to float64", v.Type())
	}
	fv := v.Convert(floatType)
	return fv.Float(), nil
}

func logerr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
