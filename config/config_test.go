package config

import (
	"os"
	"testing"

	"github.com/goblimey/go-tools/testsupport"
)

func TestParseConfig(t *testing.T) {

	json := []byte(`
		{
			"stripe_secret_key": "foo",
			"ssl_certificate_file": "pem",
			"ssl_certificate_key_file": "key"
		}
	`)

	config, err := parseConfigFromBytes(json)

	if err != nil {
		t.Error(err)
		return
	}

	if config.StripeSecretKey != "foo" {
		t.Errorf("want foo, got %s", config.StripeSecretKey)
	}

	if config.SSLCertificateFile != "pem" {
		t.Errorf("want pem, got %s", config.SSLCertificateFile)
	}

	if config.SSLCertificateKeyFile != "key" {
		t.Errorf("want key, got %s", config.SSLCertificateKeyFile)
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
			"stripe_secret_key": "foo",
			"ssl_certificate_file": "pem",
			"ssl_certificate_key_file": "key"
			
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

	if config.StripeSecretKey != "foo" {
		t.Errorf("want foo, got %s", config.StripeSecretKey)
	}

	if config.SSLCertificateFile != "pem" {
		t.Errorf("want pem, got %s", config.SSLCertificateFile)
	}

	if config.SSLCertificateKeyFile != "key" {
		t.Errorf("want key, got %s", config.SSLCertificateKeyFile)
	}
}
