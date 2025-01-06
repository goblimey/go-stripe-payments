package handler

import (
	"fmt"
	"testing"
)

// TestGetTickBox checks the getTextBox function.
func TestGetTickBox(t *testing.T) {
	var testData = []struct {
		input string
		want  bool
	}{
		{"on", true},
		{"checked", true},
		{"off", false},
		{"unchecked", false},
		{"foo", false},
		{"", false},
	}

	for _, td := range testData {
		got := getTickBox(td.input)

		if td.want != got {
			t.Errorf("%s: want %v got %v", td.input, td.want, got)
		}
	}
}

// TestValidation1 checks the first stage of validation.
func TestSimpleValidation1(t *testing.T) {

	var testData = []struct {
		description string
		form        PaymentFormData
		wantValid   bool
		want        PaymentFormData
	}{
		{
			"valid - all",
			PaymentFormData{ // This also checks the Trimspace calls.
				Valid:       false,
				PaymentYear: 2024, OrdinaryMemberFeeStr: "1.2", AssociateMemberFeeStr: "3.4", FriendFeeStr: "5.6",
				FirstName: " a\t", LastName: " b ", Email: " a@b.com ", FriendStr: "on",
				AssocFirstName: " f ", AssocLastName: " l ", AssocEmail: "  a@l.com  ", AssocFriendStr: "on",
				DonationToSocietyStr: " 7.8\t", DonationToMuseumStr: " 8.9 ",
				DonationToSociety: 0.0, DonationToMuseum: 0.0, Friend: true, AssocFriend: true, UserID: 0, AssocUserID: 0,
				GeneralErrorMessage: "", FirstNameErrorMessage: "", LastNameErrorMessage: "", EmailErrorMessage: "",
				AssocFirstNameErrorMessage: "", AssocLastNameErrorMessage: "", DonationToSocietyErrorMessage: "", DonationToMuseumErrorMessage: "",
			},
			true,
			PaymentFormData{
				true, 2024, "1.2", "3.4", "5.6",
				"a", "b", "a@b.com", "on",
				"f", "l", "a@l.com", "on", "7.8", "8.9",
				true, true,
				1.2, 3.4, 5.6, 7.8, 8.9, 0, 0, "", "", "", "", "", "", "", "",
			},
		},
		{
			"valid - no friends",
			PaymentFormData{
				false, 2024, "24.0", "6.0", "5.0", "a", "b", "a@b.com", "off", "f", "l", "a@l.com", "off",
				"1.5", "2.5", false, false,
				0.0, 0.0, 0.0, 0.0, 0.0, 0, 0, "", "", "", "", "", "", "", "",
			},
			true,
			PaymentFormData{
				true, 2024, "24.0", "6.0", "5.0", "a", "b", "a@b.com", "off", "f", "l", "a@l.com", "off",
				"1.5", "2.5", false, false,
				24.0, 6.0, 5, 1.5, 2.5, 0, 0, "", "", "", "", "", "", "", "",
			},
		},
		{
			"valid - no associate",
			PaymentFormData{
				false, 2024, "24.0", "6.0", "5.0", "a", "b", "a@b.com", "on", "", "", "", "",
				"1.5", "2.5", false, false,
				0.0, 0.0, 0.0, 0.0, 0.0, 0, 0, "", "", "", "", "", "", "", "",
			},
			true,
			PaymentFormData{
				true, 2024, "24.0", "6.0", "5.0", "a", "b", "a@b.com", "on", "", "", "", "off",
				"1.5", "2.5", true, false,
				24.0, 6.0, 5, 1.5, 2.5, 0, 0, "", "", "", "", "", "", "", "",
			},
		},
		{
			"valid - ordinary member is not friend",
			PaymentFormData{
				false, 2024, "24.0", "6.0", "5.0", "a", "b", "a@b.com", "", "f", "l", "a@l.com", "on",
				"1.5", "2.5", false, false,
				0.0, 0.0, 0.0, 0.0, 0.0, 0, 0, "", "", "", "", "", "", "", "",
			},
			true,
			PaymentFormData{
				true, 2024, "24.0", "6.0", "5.0", "a", "b", "a@b.com", "off", "f", "l", "a@l.com", "on",
				"1.5", "2.5", false, true,
				24.0, 6.0, 5, 1.5, 2.5, 0, 0, "", "", "", "", "", "", "", "",
			},
		},
		{
			"valid - associate is not friend",
			PaymentFormData{
				false, 2024, "24.0", "6.0", "5.0", "a", "b", "a@b.com", "on", "f", "l", "a@l.com", "off",
				"1.5", "2.5", false, false,
				0.0, 0.0, 0.0, 0.0, 0.0, 0, 0, "", "", "", "", "", "", "", "",
			},
			true,
			PaymentFormData{
				true, 2024, "24.0", "6.0", "5.0", "a", "b", "a@b.com", "on", "f", "l", "a@l.com", "off",
				"1.5", "2.5", true, false,
				24.0, 6.0, 5, 1.5, 2.5, 0, 0, "", "", "", "", "", "", "", "",
			},
		},
		{
			"ordinary member first name missing",
			PaymentFormData{
				false, 2024, "24.0", "6.0", "5.0", "", "b", "a@b.com", "on", "f", "l", "a@l.com", "on",
				"1.5", "2.5", false, false,
				0.0, 0.0, 0.0, 0.0, 0.0, 0, 0, "", "", "", "", "", "", "", "",
			},
			false,
			PaymentFormData{
				false, 2024, "24.0", "6.0", "5.0", "", "b", "a@b.com", "on", "f", "l", "a@l.com", "on",
				"1.5", "2.5", true, true,
				24.0, 6.0, 5, 1.5, 2.5, 0, 0, "", firstNameErrorMessage, "", "", "", "", "", "",
			},
		},
		{
			"ordinary member last name missing",
			PaymentFormData{
				false, 2024, "24.0", "6.0", "5.0", "a", "", "a@b.com", "on", "f", "l", "a@l.com", "on",
				"1.5", "2.5", false, false,
				0.0, 0.0, 0.0, 0.0, 0.0, 0, 0, "", "", "", "", "", "", "", "",
			},
			false,
			PaymentFormData{
				false, 2024, "24.0", "6.0", "5.0", "a", "", "a@b.com", "on", "f", "l", "a@l.com", "on",
				"1.5", "2.5", true, true,
				24.0, 6.0, 5, 1.5, 2.5, 0, 0, "", "", lastNameErrorMessage, "", "", "", "", "",
			},
		},
		{
			"ordinary member email missing",
			PaymentFormData{
				false, 2024, "24.0", "6.0", "5.0", "a", "b", "", "on", "f", "l", "b@l.com", "on",
				"1.5", "2.5", false, false,
				0.0, 0.0, 0.0, 0.0, 0.0, 0, 0, "", "", "", "", "", "", "", "",
			},
			false,
			PaymentFormData{
				false, 2024, "24.0", "6.0", "5.0", "a", "b", "", "on", "f", "l", "b@l.com", "on",
				"1.5", "2.5", true, true,
				24.0, 6.0, 5, 1.5, 2.5, 0, 0, "", "", "", emailErrorMessage, "", "", "", "",
			},
		},
		{
			"associate member first name missing",
			PaymentFormData{
				false, 2024, "24.0", "6.0", "5.0", "a", "b", "a@l.com", "on", "", "l", "b@l.com", "on",
				"1.5", "2.5", false, false,
				0.0, 0.0, 0.0, 0.0, 0.0, 0, 0, "", "", "", "", "", "", "", "",
			},
			false,
			PaymentFormData{
				false, 2024, "24.0", "6.0", "5.0", "a", "b", "a@l.com", "on", "", "l", "b@l.com", "on",
				"1.5", "2.5", true, true,
				24.0, 6.0, 5, 1.5, 2.5, 0, 0, "", "", "", "", assocFirstNameErrorMessage, "", "", "",
			},
		},
		{
			"associate member last name missing",
			PaymentFormData{
				false, 2024, "24.0", "6.0", "5.0", "a", "b", "a@l.com", "", "f", "", "b@l.com", "",
				"1.5", "2.5", false, false,
				0.0, 0.0, 0.0, 0.0, 0.0, 0, 0, "", "", "", "", "", "", "", "",
			},
			false,
			PaymentFormData{
				false, 2024, "24.0", "6.0", "5.0", "a", "b", "a@l.com", "off", "f", "", "b@l.com", "off",
				"1.5", "2.5", false, false,
				24.0, 6.0, 5, 1.5, 2.5, 0, 0, "", "", "", "", "", assocLastNameErrorMessage, "", "",
			},
		},
		{
			"associate member but no ordinary member",
			PaymentFormData{
				false, 2024, "24.0", "6.0", "5.0", "", "", "", "", "f", "l", "a@l.com", "",
				"1.5", "2.5", false, false,
				0.0, 0.0, 0.0, 0.0, 0.0, 0, 0, "", "", "", "", "", "", "", "",
			},
			false,
			PaymentFormData{
				false, 2024, "24.0", "6.0", "5.0", "", "", "", "off", "f", "l", "a@l.com", "off",
				"1.5", "2.5", false, false,
				24.0, 6.0, 5, 1.5, 2.5, 0, 0, "", firstNameErrorMessage, lastNameErrorMessage, emailErrorMessage, "", "", "", "",
			},
		},
		{
			"associate email address but associate member's name missing",
			PaymentFormData{
				false, 2024, "24.0", "6.0", "5.0", "a", "b", "a@b.com", "on", "", "", "a@l.com", "on",
				"1.5", "2.5", false, false,
				0.0, 0.0, 0.0, 0.0, 0.0, 0, 0, "", "", "", "", "", "", "", "",
			},
			false,
			PaymentFormData{
				false, 2024, "24.0", "6.0", "5.0", "a", "b", "a@b.com", "on", "", "", "a@l.com", "on",
				"1.5", "2.5", true, true,
				24.0, 6.0, 5, 1.5, 2.5, 0, 0, "", "", "", "", assocFirstNameErrorMessage, assocLastNameErrorMessage, "", "",
			},
		},
		{
			"associate friend tick box but associate member's name missing",
			PaymentFormData{
				false, 2024, "24.0", "6.0", "5.0", "a", "b", "a@b.com", "on", "", "", "", "on",
				"1.5", "2.5", false, false,
				0.0, 0.0, 0.0, 0.0, 0.0, 0, 0, "", "", "", "", "", "", "", "",
			},
			false,
			PaymentFormData{
				false, 2024, "24.0", "6.0", "5.0", "a", "b", "a@b.com", "on", "", "", "", "on",
				"1.5", "2.5", true, true,
				24.0, 6.0, 5, 1.5, 2.5, 0, 0, "", "", "", "", assocFirstNameErrorMessage, assocLastNameErrorMessage, "", "",
			},
		},
		{
			"donation to society invalid number",
			PaymentFormData{
				false, 2024, "24.0", "6.0", "5.0", "a", "b", "a@b.com", "off", "f", "l", "a@l.com", "off",
				"junk", "2.5", false, false,
				0.0, 0.0, 0.0, 0.0, 0.0, 0, 0, "", "", "", "", "", "", "", "",
			}, false,
			PaymentFormData{
				false, 2024, "24.0", "6.0", "5.0", "a", "b", "a@b.com", "off", "f", "l", "a@l.com", "off",
				"junk", "2.5", false, false,
				24.0, 6.0, 5.0, 0.0, 2.5, 0, 0, "", "", "", "", "", "", invalidNumber, "",
			},
		},
		{
			"donation to museum invalid number",
			PaymentFormData{
				false, 2024, "24.0", "6.0", "5.0", "a", "b", "a@b.com", "on", "f", "l", "a@l.com", "on",
				"1.5", "junk", false, false,
				0.0, 0.0, 0.0, 0.0, 0.0, 0, 0, "", "", "", "", "", "", "", "",
			},
			false,
			PaymentFormData{
				false, 2024, "24.0", "6.0", "5.0", "a", "b", "a@b.com", "on", "f", "l", "a@l.com", "on",
				"1.5", "junk", true, true,
				24.0, 6.0, 5, 1.5, 0.0, 0, 0, "", "", "", "", "", "", "", invalidNumber,
			},
		},
	}

	for _, td := range testData {
		valid := simpleValidate(&td.form)
		if td.wantValid != valid {
			t.Errorf("%s want %v got %v", td.description, td.wantValid, valid)
		}

		if td.want != td.form {
			fmt.Printf("%s:\nwant %v\n got %v\n", td.description, td.want, td.form)
		}

	}
}
