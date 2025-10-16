/*
	     go-stripe-payments runs a web server that takes membership subscription and
		 either creates a new account in the Admidio database or, for a renewal, updates
		 the existing account.  To support https it needs access to TLS certificate
		 files.

		 TLS cert files are updated occasionally.  If they are provided by letsencrypt,
		 the files are called cert.pem, chain.pem, fullchain.pem and privkey.pem.
		 However, those are just links to file with a number in the name - cert1.pem,
		 cert2.pem and so on.  The certificate is renewed a few days before it expires.
		 During this process new files are created with the number incremented and the
		 links adjusted to point to those file.  So if cert.pem is a link to cert1.pem,
		 a few days before the certificate is due to expire, cert2.pem etc are created
		 and the links adjusted to both files.  Both sets of certificate files are
		 valid for a few days and during that time either can be used.

		 To handle this, the application reads the cert.pem etc on startup (ie using the
		 links), uses those data for that day and dies at midnight.  It's assumed that it
		 will be run by a script that runs forever in a loop.  The script should start
		 the program running, wait for it to finish, start it again and so on ad infinitum.
		 On the day that the certificate will be renewed the program will pick up cert
		 version n and the next day it will pick up the new certificate version n+1.

		 In production the certificate files are owned by root and privkey.pem is readable
		 only by that user.  The looping script should run as root so that the program
		 runs as root initially.  It reads the certificate files and uses the contents for
		 the rest of that day.  Running a web server as root is a bad idea so as soon as
		 the program has read the cert files and before it offers its web interface, it
		 switches to an ordinary user, specified in the config.
*/
package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/user"
	"strconv"
	"time"

	"github.com/stripe/stripe-go/v81"

	"github.com/goblimey/go-tools/dailylogger"

	"github.com/goblimey/go-stripe-payments/code/apps/payments/handler"
	"github.com/goblimey/go-stripe-payments/code/pkg/config"
	"github.com/goblimey/go-stripe-payments/code/pkg/database"
	"github.com/goblimey/go-stripe-payments/code/pkg/shutdown"
	"github.com/goblimey/go-stripe-payments/code/pkg/usercontrol"
)

// The config.
var conf *config.Config

func main() {

	// Get configuration.
	var errConfig error
	conf, errConfig = config.GetConfig("./config.json")
	if errConfig != nil {
		fmt.Println(errConfig.Error)
		os.Exit(-1)
	}

	// The stripe secret key.
	stripe.Key = conf.StripeSecretKey

	dbConfig := database.DBConfig{
		Type: conf.DBType,
		Host: conf.DBHostname,
		Port: conf.DBPort,
		Name: conf.DBDatabase,
		User: conf.DBUser,
		Pass: conf.DBPassword,
	}

	hdlr := handler.New(conf)
	// http.Handle("/", http.FileServer(http.Dir("public")))
	http.HandleFunc("/", hdlr.Home)
	http.HandleFunc("/index.html", hdlr.Home)
	http.HandleFunc("/displayPaymentForm", hdlr.GetPaymentData)
	http.HandleFunc("/checkout", hdlr.Checkout)
	http.HandleFunc("/success", hdlr.Success)
	http.HandleFunc("/cancel", hdlr.Cancel)
	http.HandleFunc("/create-checkout-session", hdlr.CreateCheckoutSession)

	// Run a test connection to the database.
	db := database.New(&dbConfig)
	connectError := db.Connect()
	if connectError != nil {
		fmt.Println(connectError.Error())
		return
	}
	db.Close()

	if conf.HTTP {
		// No TLS cert files are supplied so we offer an HTTP service.  There
		// is no need to start off running as root and no need to shut down
		// at midnight.
		hdlr.Logger = GetDailyLogger(conf.LogDir, conf.LogLeader)
		hdlr.Logger.Info("starting http server " + conf.Address)
		serverStartErr := http.ListenAndServe(conf.Address, nil)
		if serverStartErr != nil {
			hdlr.Fatal(serverStartErr)
		}

	} else {

		// We are going to run an HTTPS server.  That involves reading the TLS cert files
		// so we must be root at this point.  We don't want the daily log to be owned by root
		// so don't create it yet.  Report any problems on stdout.
		//
		// Each night just before midnight the app shuts down.  It's restarted just after
		// midnight.  Initially it runs as root so that it can read the cert files.  We shut
		// the server down at the end of each day to support the certificate renewal scheme
		// used by agencies such as LetsEnrcrypt.  A few days before the current cert becomes
		// invalid a new cerificate is created.  Both the old cert and the new cert are valid
		// at first so the cert data that was read earlier that day will still work even
		// though it's not now the current version.  When the app restarts the next morning it
		// will pick up the new certificate.  The old cert is then redundant and will become
		// invalid in a few days.
		//
		// To organise the daily restart, the app is run by a shell script that watches for it
		// shutting down, pauses for a few seconds (to move into the next day) and starts it
		// up again.  If the server falls over earlier in the day due to some fatal error, it
		// will be restarted after a couple of seconds and pick up the current certificate.

		if len(conf.RunUser) <= 0 {
			log.Fatal("https specified but no run user")
		}

		if len(conf.TLSCertificateFile) == 0 || len(conf.TLSCertificateKeyFile) == 0 {
			log.Fatal("cert files not specified")
		}

		// The config contains the names of the TLS certificate files.  Read them.
		// In production we must do this while we are root because the key file is
		// readable only by that user.)
		certFileBytes, readCertFileError := os.ReadFile(conf.TLSCertificateFile)
		if readCertFileError != nil {
			// One obvious explanation is that we are not running as root.
			if usercontrol.Getuid() != 0 {
				log.Fatal("must be root to read the TLS cert")
			}
			log.Fatal(readCertFileError.Error())
		}

		keyFileBytes, readKeyFileError := os.ReadFile(conf.TLSCertificateKeyFile)
		if readKeyFileError != nil {
			log.Fatal(readKeyFileError.Error())
		}

		// The config contains the name of the user that this server will run under.
		// Get the uid.
		u, userError := user.Lookup(conf.RunUser)
		if userError != nil {
			log.Fatal(userError.Error())
		}

		userID, idError := strconv.Atoi(u.Uid)
		if idError != nil {
			log.Fatal(idError.Error())
		}

		// In production the program should start off running as root, so that it can read
		// the TLS cerificate, but from now on it can run as an ordinary user, specified in
		// the config.  The Setuid below switches users but it only works under UNIX or
		// Linux.  Under Windows the function exists but it returns an error, so the program
		// compiles but it can't run the code here.  The program can run under Windows but
		// only in HTTP mode, allowing some system testing to be done in that environment.
		suError := usercontrol.Setuid(userID)
		if suError != nil {
			log.Fatal(suError.Error())
		}

		// From this point onward the server is running as the non-root user defined by the
		// config.  It may fall over later due to a fatal error and, just before midnight, it
		// will shut itself down.
		//
		// The server should be run by some runner device such as a shell script which repeatedly
		// starts it, waits for it finish, pauses for a few seconds, then starts it again, and
		// so on.  The server will shut itself down just before midnight and the runner will
		// start it up again a few seconds into the next morning.
		//
		// Now we create the daily logger.  If we are starting up after a planned midnight
		// shutdown, it's first thing in the morning so we will create a new log file with a
		// datestamped name (such as "payment.2025-02-14.log") in the configured log directory.
		// If we are restarting because the server fell over earlier in the day due to a fatal
		// error we will pick up the log file that was created earlier.
		//
		// We don't want the daily log file to be owned by root so we wait until now to create
		// it.  The log directory must be writeable by the user that we just switched to.
		hdlr.Logger = GetDailyLogger(conf.LogDir, conf.LogLeader)

		message := fmt.Sprintf("Running as an HTTPS server as user %s, uid %d", u.Name, userID)
		hdlr.Logger.Info(message)

		// Set the server to shut down just before midnight.
		now := time.Now()
		go shutdown.PauseAndShutdown(now)

		// We have a certificate - start an https server.
		cert, certError := tls.X509KeyPair(certFileBytes, keyFileBytes)
		if certError != nil {
			hdlr.Logger.Error(certError.Error())
			return
		}

		var tlsConfig *tls.Config
		if len(conf.DBHostname) == 0 {
			//A hostname is supplied.  Include it in the config.
			tlsConfig = &tls.Config{
				Certificates: []tls.Certificate{cert},
				ServerName:   conf.DBHostname,
			}
		} else {
			tlsConfig = &tls.Config{
				Certificates: []tls.Certificate{cert},
			}
		}

		// Build a server.
		server := http.Server{
			Addr:      conf.Address,
			TLSConfig: tlsConfig,
			// Add any other options here.
		}

		hdlr.Logger.Info("starting https server " + conf.Address)
		serverStartErr := server.ListenAndServeTLS("", "")
		hdlr.Fatal(serverStartErr)
	}

	// The ListenAndServe loops forever or returns an error, which stops the program.
	// Either way we never get to here.
}

// GetDailyLogger gets a daily log file which can be written to as a logger
// (each line decorated with filename, date, time, etc).  The name argument
// is used to form the log file name.
func GetDailyLogger(logDir, leader string) *slog.Logger {
	// Creater a daily log writer.
	name := leader + "."
	dailyLogWriter := dailylogger.New(logDir, name, ".log")

	// Create a structured logger that writes to the dailyLogWriter.
	logger := slog.New(slog.NewTextHandler(dailyLogWriter, nil))

	return logger

	// Create a logger which writes to the writer.
	// logFlags := log.LstdFlags | log.Lshortfile | log.Lmicroseconds
	// return log.New(dailyLog, leader, logFlags)
}
