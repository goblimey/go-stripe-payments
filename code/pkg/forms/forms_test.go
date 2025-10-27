package forms

import (
	"testing"
)

// TestGetCountryFromCode checks the CountryFromCode function.
func TestGetCountryFromCode(t *testing.T) {

	form := ExtraDetailForm{CountryCode: "junk"}

	got1 := form.SetCountryFromCode()

	if got1 != "" {
		t.Errorf("want empty string got %s", got1)
	}

	form.CountryCode = "GBR"

	got2 := form.SetCountryFromCode()

	if got2 != "United Kingdom of Great Britain and Northern Ireland" {
		t.Errorf("want United Kingdom of Great Britain and Northern Ireland got %s", got2)
	}

}
