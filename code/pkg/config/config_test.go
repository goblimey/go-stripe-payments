package config

import (
	"os"
	"testing"

	"github.com/goblimey/go-tools/testsupport"
)

func TestParseConfig(t *testing.T) {

	json := []byte(`
		{
			
			"run_user": "simon",
			"log_dir": ".",
			"logfile_group": "peter",
			"logdir_permissions": "0777",
			"logfile_permissions": "0666",
    		"log_leader": "lead.",
			"log_trailer": ".trail",
			"tls_certificate_file": "pem",
			"tls_certificate_key_file": "key",
			"organisation_name": "some name",
			"enable_other_member_types": true,
			"enable_giftaid": true,
			"email_address_for_failures": "foo@example.com",
			"email_address_for_questions": "bar@example.com",
			"db_type": "type",
			"db_host": "this",
			"db_port": 1234,
			"db_database": "db",
			"db_user": "me",
			"stripe_secret_key": "foo",
			"ordinary_member_fee": 1.1,
			"associate_member_fee": 2.2,
			"friend_fee": 3.3
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

	if conf.LogFileGroup != "peter" {
		t.Errorf("want peter, got %s", conf.LogFileGroup)
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

	if conf.EmailAddressForQuestions != "bar@example.com" {
		t.Errorf("want bar@example.com, got %s", conf.EmailAddressForQuestions)
	}

	if conf.LogDir != "." {
		t.Errorf("want \".\" got %s", conf.LogDir)
	}

	if conf.LogLeader != "lead." {
		t.Errorf("want \"lead.\" got %s", conf.LogLeader)
	}

	if conf.LogTrailer != ".trail" {
		t.Errorf("want \".trail\" got %s", conf.LogLeader)
	}

	md, ed := conf.LogDirMode()
	if ed != nil {
		t.Error(ed)
	}
	if md != 0777 {
		t.Errorf("want 0777 got 0%o", md)
	}

	mf, ef := conf.LogFileMode()
	if ef != nil {
		t.Error(ef)
	}
	if mf != 0666 {
		t.Errorf("want 0666 got 0%o", mf)
	}

	if conf.OrdinaryMemberFee != 1.1 {
		t.Errorf("want 1.1, got %f", conf.OrdinaryMemberFee)
	}

	if conf.AssocMemberFee != 2.2 {
		t.Errorf("want 2.2, got %f", conf.AssocMemberFee)
	}

	if conf.FriendFee != 3.3 {
		t.Errorf("want 3.3, got %f", conf.FriendFee)
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

	const configFile = "config.json"
	const configContents = `
		{
			"organisation_name": "some org",
			"tls_certificate_file": "pem",
			"tls_certificate_key_file": "key"
		}
	`

	writer, fileCreateError := os.Create(configFile)
	if fileCreateError != nil {
		t.Error(fileCreateError)
		return
	}

	json := []byte(configContents)
	_, writeError := writer.Write([]byte(json))
	if writeError != nil {
		t.Error(writeError)
		return
	}

	jsonS := string(json)
	_ = jsonS

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
