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
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/user"
	"strconv"
	"syscall"
	"time"

	"github.com/stripe/stripe-go/v81"

	"github.com/goblimey/dailylogger"
	ps "github.com/goblimey/portablesyscall"

	"github.com/goblimey/go-stripe-payments/code/apps/payments/handler"
	"github.com/goblimey/go-stripe-payments/code/pkg/config"
	"github.com/goblimey/go-stripe-payments/code/pkg/database"
	"github.com/goblimey/go-stripe-payments/code/pkg/shutdown"
)

// The config.
var conf *config.Config

func main() {

	// Get configuration.
	var errConfig error
	conf, errConfig = config.GetConfig("./config.json")
	if errConfig != nil {
		fmt.Println(errConfig.Error())
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

	// Create the daily log file.  In production we are running as root.  We don't
	// want to have be root to read the log, so set it owned as the target user.
	hdlr.Logger = GetDailyLogger(conf)

	http.HandleFunc("/", hdlr.Home)
	http.HandleFunc("/index.html", hdlr.Home)
	http.HandleFunc("/subscribe", hdlr.GetPaymentData)
	http.HandleFunc("/subscribe/", hdlr.GetPaymentData)
	http.HandleFunc("/checkout", hdlr.Checkout)
	http.HandleFunc("/success", hdlr.Success)
	http.HandleFunc("/extradetails", hdlr.ExtraDetails)
	http.HandleFunc("/completion", hdlr.Completion)
	http.HandleFunc("/cancel", hdlr.Cancel)
	http.HandleFunc("/create-checkout-session", hdlr.CreateCheckoutSession)
	// Backward compatibility:
	http.HandleFunc("/displayPaymentForm", hdlr.GetPaymentData)
	http.HandleFunc("/displayPaymentForm/", hdlr.GetPaymentData)

	// Run a test connection to the database.
	db := database.New(&dbConfig)
	connectError := db.Connect()
	if connectError != nil {
		hdlr.Fatal(connectError)
		return
	}
	db.Close()

	if ps.OSName == "windows" {
		// The web server is running under Windows.  No TLS certificate files are supplied
		// so we can off HTTP but not HTTPS.  Running as HTTP allows some useful system
		// testing under Windows.
		hdlr.Logger.Info("starting http server " + conf.Address)
		serverStartErr := http.ListenAndServe(conf.Address, nil)
		if serverStartErr != nil {
			hdlr.Fatal(serverStartErr)
		}

	} else {

		// We are going to run an HTTPS server.  That involves reading the TLS cert files
		// so we must be root at this point.  Running a web server under root creates all
		// sorts of security issues so we want to run as a less privilaged user as soon
		// as possible.  We need to be able to log errors so we create a log file readable
		// by the user only.  We don't want to have to be root to read it so we change the
		// owner to the same less privileged user.
		//
		// We are using a Letsencrypt certificate to support HTTPS.  At some point in the
		// future a new one will be created and a few days after that, the one we are using
		// will become invalid.  They are both valid during the intervening few days. We need
		// to be root again to read the new certificate files.  To handle that we use the
		// same approach as the Apache web server.  The server runs under the control of
		// another process running as root.  If the server shuts down, the controller waits
		// a couple of seconds and restarts it.  Each night just before midnight the server
		// shuts itself down and the control process starts it up again just after midnight.
		// If a new certificate is created that day, the one we read earier still works.  The
		// next morning we will pick up the new certificate.  A few days later the old
		// certificate will become invalid but by then it's redundant.
		//
		// If the server falls over during the day due to some fatal error, it will be
		// restarted after a couple of seconds and pick up the current certificate.

		if len(conf.RunUser) <= 0 {
			hdlr.Fatal(errors.New("https specified but no run user"))
		}

		if len(conf.TLSCertificateFile) == 0 || len(conf.TLSCertificateKeyFile) == 0 {
			hdlr.Fatal(errors.New("cert files not specified"))
		}

		// The config contains the names of the TLS certificate files.  Read them.
		// In production we must do this while we are root because the key file is
		// readable only by that user.)
		certFileBytes, readCertFileError := os.ReadFile(conf.TLSCertificateFile)
		if readCertFileError != nil {
			// One obvious explanation is that we are not running as root.
			if syscall.Getuid() != 0 {
				hdlr.Fatal(errors.New("must be root to read the TLS cert"))
			}
			hdlr.Fatal(readCertFileError)
		}

		keyFileBytes, readKeyFileError := os.ReadFile(conf.TLSCertificateKeyFile)
		if readKeyFileError != nil {
			hdlr.Fatal(readKeyFileError)
		}

		// The config contains the name of the user that this server will run under.
		// Get the uid.
		u, userError := user.Lookup(conf.RunUser)
		if userError != nil {
			hdlr.Fatal(userError)
		}

		userID, idError := strconv.Atoi(u.Uid)
		if idError != nil {
			hdlr.Fatal(idError)
		}

		// In production the program should start off running as root, so that it can read
		// the TLS cerificate, but from now on it can run as an ordinary user, specified in
		// the config.  This is done by calling syscall.Setuid, but that only works on a
		// POSIX target such as Linux, not under Windows.
		//
		// The Windows version of syscall offers a Getuid but no Setuid, so a program that
		// calls syscall.Setuid won't compile if Windows is the target.  The material in the
		// portablesyscall package exists for all targets.  Running on a POSIX system such as
		// UNIX or Linux portablesyscall.setuid changes the user running the program from root
		// to the given user.  If Setuid is called by a program running under Windows, it
		// returns an error, but the program still compiles and will run.  It just can't
		// attempt to change the user that's running it.
		//
		// The upshot is that this server can offer an HTTP service under Windows but not HTTPS,
		// which enables some useful system testing in that environment.
		suError := ps.Setuid(userID)
		if suError != nil {
			hdlr.Fatal(suError)
		}

		// From this point onward the server is running as the non-root user defined by the
		// config.  It may fall over later due to a fatal error and, just before midnight, it
		// will shut itself down.
		//
		// The server should be run by some device such as a shell script which repeatedly
		// starts it, waits for it finish, pauses for a few seconds, then starts it again.  The
		// server will shut itself down just before midnight and the runner will start it up
		// again a few seconds into the next morning.
		//
		// Now we create the daily logger.  If we are starting up after a planned midnight
		// shutdown, it's first thing in the morning so we will create a new log file with a
		// datestamped name (such as "payment.2025-02-14.log") in the configured log directory.
		// If we are restarting because the server fell over earlier in the day due to a fatal
		// error we will pick up the log file that was created earlier.
		//

		message := fmt.Sprintf("Running as an HTTPS server as user %s", u.Name)
		hdlr.Logger.Info(message)

		// Set the server to shut down just before midnight.
		now := time.Now()
		go shutdown.PauseAndShutdown(now)

		// We have a certificate.  Create a key pair in memory.
		cert, certError := tls.X509KeyPair(certFileBytes, keyFileBytes)
		if certError != nil {
			hdlr.Logger.Error(certError.Error())
			return
		}

		// Ceate a TLS config for the HTTPS server.
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

		// Create an HTTPS server.
		server := http.Server{
			Addr:      conf.Address,
			TLSConfig: tlsConfig,
			// Add any other options here.
		}

		hdlr.Logger.Info("starting https server as" + conf.Address)
		serverStartErr := server.ListenAndServeTLS("", "")
		hdlr.Fatal(serverStartErr)
	}

	// The ListenAndServe loops forever or returns an error, which stops the program.
	// Either way we never get to here.
}

// GetDailyLogger gets a daily log file which can be written to as a logger
// (each line decorated with filename, date, time, etc).  The name argument
// is used to form the log file name.
// func GetDailyLogger(logDir, leader string) *slog.Logger {
// 	// Creater a daily log writer.
// 	name := leader + "."
// 	dailyLogWriter := dailylogger.New(logDir, name, ".log")

// 	// Create a structured logger that writes to the dailyLogWriter.
// 	logger := slog.New(slog.NewTextHandler(dailyLogWriter, nil))

// 	return logger
// }

// GetDailyLogger gets a daily log file which can be written to as a SLOG logger
// The form (name, )
// is used to form the log file name.  The file will be owned by the configured
// user, will be in the configured group and will have the configured permissions
func GetDailyLogger(conf *config.Config) *slog.Logger {
	f := "GetDailyLogger"
	md, mde := conf.LogDirMode()
	if mde != nil {
		fmt.Printf("%s: error getting log directory mode - %v\n", f, mde)
		return nil
	}
	mf, mfe := conf.LogFileMode()
	if mfe != nil {
		fmt.Printf("%s: error getting log file mode - %v\n", f, mfe)
		return nil
	}
	// Create a daily log writer from the config.
	dailyLogWriter := dailylogger.New(
		time.Now(),
		conf.LogDir,
		conf.LogLeader,
		conf.LogTrailer,
		conf.RunUser,
		conf.LogFileGroup,
		md,
		mf,
	)

	// Create a structured logger that writes to the dailyLogWriter.
	logger := slog.New(slog.NewTextHandler(dailyLogWriter, nil))
	if logger == nil {
		fmt.Println("failed to create SLOG logger")
		return nil
	}

	// Log the first message
	logger.Info("started up and logging")

	return logger
}
