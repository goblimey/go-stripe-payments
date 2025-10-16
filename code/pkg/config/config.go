package config

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// Config holds the configuration.
type Config struct {
	// Tehese config values are taken from the given config file.
	OrganisationName         string  `json:"organisation_name"`           // The name of the organisation for display
	HTTP                     bool    `json:"http"`                        // true if the server should run as HTTP not HTTPS (usually for testing).
	RunUser                  string  `json:"run_user"`                    // The name of the non-root user that will run the server.
	TLSCertificateFile       string  `json:"tls_certificate_file"`        // The TLS certificate file.
	TLSCertificateKeyFile    string  `json:"tls_certificate_key_file"`    // the secret TLS key file.
	EnableOtherMemberTypes   bool    `json:"enable_other_member_types"`   // Enable associate members, friends etc.
	EnableGiftaid            bool    `json:"enable_giftaid"`              // Enable Giftaid.
	EmailAddressForQuestions string  `json:"email_address_for_questions"` // Email address for questions.
	EmailAddressForFailures  string  `json:"email_address_for_failures"`  // Email address for payment failure messages.
	OrdinaryMemberFee        float64 `json:"ordinary_member_fee"`         // Ordinary membership fee.
	AssocMemberFee           float64 `json:"associate_member_fee"`        // Associate membership system.
	FriendFee                float64 `json:"friend_fee"`                  // Friend of the museum fee.
	LogDir                   string  `json:"log_dir"`                     // The directory in which the daily log is created.
	LogLeader                string  `json:"log_leader"`                  // The first part of the log file name.

	// Secrets are taken from the environment.
	StripeSecretKey string
	Hostname        string
	Port            string
	DBType          string
	DBHostname      string
	DBPort          string
	DBDatabase      string
	DBUser          string
	DBPassword      string
	Address         string
}

// GetConfig gets the config from the given file.
func GetConfig(configFile string) (*Config, error) {
	file, err := os.Open(configFile)
	if err != nil {
		em := fmt.Sprintf("[-] Cannot open config file: %s\n", err.Error())
		fmt.Println(em)
		os.Exit(1)
	}

	config, errParse := getConfigFromReader(file)

	if errParse != nil {
		return nil, errParse
	}

	return config, nil
}

// GetConfigFrom Reader gets the config from the given reader.
func getConfigFromReader(configReader io.Reader) (*Config, error) {

	data := make([]byte, 4096)
	n, errRead := configReader.Read(data)
	if errRead != nil {
		em := fmt.Sprintf("[-] Error reading config file: %s\n", errRead.Error())
		fmt.Println(em)
		return nil, errRead
	}

	config, parseError := parseConfigFromBytes(data[:n])
	if parseError != nil {
		em := fmt.Sprintf("[-] Not a valid config file: %s\n", parseError.Error())
		fmt.Println(em)
		return nil, parseError
	}

	return config, nil
}

func parseConfigFromBytes(data []byte) (*Config, error) {
	var config Config
	err := json.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}
	// Get the secets from the environment.

	// The stripe secret key.
	config.StripeSecretKey = os.Getenv("StripeSecretKey")
	// The database hostname.
	config.Hostname = os.Getenv("hostname")
	// The port that this web server will run one.
	config.Port = os.Getenv("port")
	// The database type - "postgres" in production.
	config.DBType = os.Getenv("DBType")
	// The hostname that thedatabase server is running on.
	config.DBHostname = os.Getenv("DBHost")
	// The database host.
	config.DBPort = os.Getenv("DBPort")
	// The database (schema).
	config.DBDatabase = os.Getenv("DBDatabase")
	// The database user.
	config.DBUser = os.Getenv("DBUser")
	// The database password.
	config.DBPassword = os.Getenv("DBPassword")

	// The address of this web server is "hostname:port".
	config.Address = config.Hostname + ":" + config.Port // Accept requests to this name.

	return &config, nil
}
