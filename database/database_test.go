package database

import (
	"testing"
)

func TestConnectPostgres(t *testing.T) {

	db := New(dbConfigForTestingWithPostgres)

	err := db.Connect()
	if err != nil {
		t.Error(err)
	}

	defer db.Close()

	if db == nil {
		t.Error("should return non-nil")
	}
}

func TestConnectSQLite(t *testing.T) {

	db := New(dbConfigForTestingWithSQLite)

	err := db.Connect()
	if err != nil {
		t.Error(err)
	}

	defer db.Close()

	if db == nil {
		t.Error("should return non-nil")
	}
}

// TestPostgresParamsToSQLiteParams checks the postgresParamsToSQLiteParams.
func TestPostgresParamsToSQLiteParams(t *testing.T) {

	const multilineQuery = `abc$21
		def$21`
	const multilineResult = `abc?
		def?`

	var testData = []struct {
		query string
		want  string
	}{
		{"abc$1def$2", "abc?def?"},
		{"$1$2$3", "???"},
		{"noparams", "noparams"},
		{"abc$21def$21", "abc?def?"},
		{multilineQuery, multilineResult},
	}

	for _, td := range testData {
		got := postgresParamsToSQLiteParams(td.query)
		if got != td.want {
			t.Errorf("got %s want %s", got, td.want)
		}
	}
}
