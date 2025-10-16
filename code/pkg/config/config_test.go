package config

import (
	"os"
	"testing"

	"github.com/goblimey/go-tools/testsupport"
)

func TestParseConfig(t *testing.T) {

	json := []byte(`
		{
			"organisation_name": "some name",
			"HTTP": true,
			"run_user": "simon",
			"stripe_secret_key": "foo",
			"tls_certificate_file": "pem",
			"tls_certificate_key_file": "key",
			"enable_other_member_types": true,
			"enable_giftaid": true,
			"email_address_for_failures": "foo@example.com",
			"db_type": "type",
			"db_host": "this",
			"db_port": 1234,
			"db_database": "db",
			"db_user": "me",
			"log_dir": ".",
    		"log_leader": "payments"
		}
	`)

	// Set up the environment variables that the config                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                          parser picks up.
	os.Setenv("hostname", "lh")
	os.Setenv("port", "1")
	os.Setenv("StripeSecretKey", "foo")
	os.Setenv("DBType", "pg")
	os.Setenv("DBHost", "localhost")
	os.Setenv("DBPort", "2")
	os.Setenv("DBDatabase", "db")
	os.Setenv("DBUser", "me")
	os.Setenv("DBPassword", "pw")

	conf, err := parseConfigFromBytes(json)

	if err != nil {
		t.Error(err)
		return
	}

	if conf.Hostname != "lh" {
		t.Errorf("want l got %s", conf.Hostname)
	}
	if conf.Port != "1" {
		t.Errorf("want 1 got %s", conf.Port)
	}
	if conf.StripeSecretKey != "foo" {
		t.Errorf("want foo got %s", conf.StripeSecretKey)
	}
	if conf.DBType != "pg" {
		t.Errorf("want pg got %s", conf.DBHostname)
	}
	if conf.DBHostname != "localhost" {
		t.Errorf("want localhost got %s", conf.DBHostname)
	}
	if conf.DBPort != "2" {
		t.Errorf("want 2 got %s", conf.DBPort)
	}
	if conf.DBDatabase != "db" {
		t.Errorf("want db got %s", conf.DBDatabase)
	}
	if conf.DBUser != "me" {
		t.Errorf("want me got %s", conf.DBUser)
	}
	if conf.DBPassword != "pw" {
		t.Errorf("want pw got %s", conf.DBPassword)
	}

	if conf.OrganisationName != "some name" {
		t.Errorf("want some name, got %s", conf.OrganisationName)
	}

	if conf.RunUser != "simon" {
		t.Errorf("want simon, got %s", conf.RunUser)
	}

	if conf.StripeSecretKey != "foo" {
		t.Errorf("want foo, got %s", conf.StripeSecretKey)
	}

	if conf.TLSCertificateFile != "pem" {
		t.Errorf("want pem, got %s", conf.TLSCertificateFile)
	}

	if conf.TLSCertificateKeyFile != "key" {
		t.Errorf("want key, got %s", conf.TLSCertificateKeyFile)
	}

	if !conf.EnableOtherMemberTypes {
		t.Error("want EnableOtherMemberTypes to be true")
	}

	if !conf.EnableGiftaid {
		t.Error("want EnableGiftaid to be true")
	}

	if conf.EmailAddressForFailures != "foo@example.com" {
		t.Errorf("want foo@example.com, got %s", conf.EmailAddressForFailures)
	}

	if conf.LogDir != "." {
		t.Errorf("want \".\" got %s", conf.LogDir)
	}

	if conf.LogLeader != "payments" {
		t.Errorf("want \"payments\" got %s", conf.LogLeader)
	}
}

func TestParseConfigWithError(t *testing.T) {

	jsonData := []byte(`{junk: "junk"}`)

	_, err := parseConfigFromBytes(jsonData)

	if err == nil {
		t.Error("expected an error")
	}
}

// TestGetConfig checks that getConfig correctly reads a config file.
func TestGetConfig(t *testing.T) {

	// Create a temporary directory with a file containing the config.
	testDirName, createDirectoryError := testsupport.CreateWorkingDirectory()

	if createDirectoryError != nil {
		t.Error(createDirectoryError)
		return
	}

	// Ensure that the test files are tidied away at the end.
	defer testsupport.RemoveWorkingDirectory(testDirName)

	configFile := "config.json"

	writer, fileCreateError := os.Create(configFile)
	if fileCreateError != nil {
		t.Error(fileCreateError)
		return
	}

	const configStr = `
		{
			"organisation_name": "some org",
			"tls_certificate_file": "pem",
			"tls_certificate_key_file": "key"
			
		}
	`
	json := []byte(configStr)
	_, writeError := writer.Write([]byte(json))
	if writeError != nil {
		t.Error(writeError)
		return
	}

	config, errConfig := GetConfig("./config.json")
	if errConfig != nil {
		t.Error(errConfig)
		return
	}

	if config.OrganisationName != "some org" {
		t.Errorf("want some org, got %s", config.StripeSecretKey)
	}

	if config.TLSCertificateFile != "pem" {
		t.Errorf("want pem, got %s", config.TLSCertificateFile)
	}

	if config.TLSCertificateKeyFile != "key" {
		t.Errorf("want key, got %s", config.TLSCertificateKeyFile)
	}
}
