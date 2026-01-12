package database

import (
	"database/sql"
	"testing"
)

func TestGetCountryByCode(t *testing.T) {
	db, connError := OpenDBForTesting("sqlite")

	if connError != nil {
		t.Error(connError)
		return
	}

	db.BeginTx()

	defer db.Rollback()
	defer db.CloseAndDelete()

	prepError := PrepareTestTables(db)
	if prepError != nil {
		t.Error(prepError)
	}

	gbr, err1 := db.GetCountryByCode("GBR")
	if err1 != nil {
		t.Error(err1)
		return
	}

	if gbr.Code != "GBR" {
		t.Errorf("expected GBR, got %s", gbr.Code)
	}

	if gbr.Name != "United Kingdom" {
		t.Errorf("expected \"United Kingdom\", got %s", gbr.Name)
	}

	_, err2 := db.GetCountryByCode("junk")
	if err2 == nil {
		t.Error("expected an error")
		return
	}
	if err2 != sql.ErrNoRows {
		t.Errorf("expected an sql.ErrNoRows, got %s", err2)
		return
	}
}
