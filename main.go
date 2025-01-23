package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/stripe/stripe-go/v81"

	"github.com/goblimey/go-stripe-payments/config"
	"github.com/goblimey/go-stripe-payments/database"
	"github.com/goblimey/go-stripe-payments/handler"
)

// The config.
var conf *config.Config

func main() {

	// Get configuration.
	var errConfig error
	conf, errConfig = config.GetConfig("./config.json")
	if errConfig != nil {
		log.Fatal(errConfig)
		return
	}

	// Set the stripe secret key from the configuration.
	stripe.Key = conf.StripeSecretKey

	hdlr := handler.New(conf)

	//Get the database config and do a test connection.
	dbConfig := database.GetDBConfigFromTheEnvironment()
	db := database.New(dbConfig)
	connectError := db.Connect()
	if connectError != nil {
		fmt.Println(connectError.Error())
		return
	}

	db.Close()

	http.Handle("/", http.FileServer(http.Dir("public")))
	http.HandleFunc("/displayPaymentForm", hdlr.GetPaymentData)
	http.HandleFunc("/checkout", hdlr.Checkout)
	http.HandleFunc("/success", hdlr.Success)
	http.HandleFunc("/cancel", hdlr.Cancel)
	http.HandleFunc("/create-checkout-session", hdlr.CreateCheckoutSession)
	addr := ":4242"
	if len(conf.SSLCertificateFile) > 0 {
		// We have a certificate - accept https
		log.Fatal(
			http.ListenAndServeTLS(
				addr,
				conf.SSLCertificateFile,
				conf.SSLCertificateKeyFile, nil,
			),
		)
	} else {
		// No certificate file - accept http.
		log.Fatal(http.ListenAndServe(addr, nil))
	}

	// http.ListenAndServe loops forever or stops the program so
	// we never get to here.
}
