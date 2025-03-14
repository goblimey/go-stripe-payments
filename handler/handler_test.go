package handler

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
)

func TestSuccess(t *testing.T) {
	body := strings.NewReader(`{"username": "admin","password":"secret"}`)
	http.NewRequest("POST", "/success", body)
}

// TestCheckDonation checks the checkDonation validation method -
// a donation string value must contain a 0 or positive float.
func TestCheckDonation(t *testing.T) {
	var testData = []struct {
		str              string
		wantErrorMessage string
		wantValue        float64
	}{

		{"1.3", "", 1.3},
		{"0.1", "", 0.1},
		{"junk", invalidNumber, 0.0},
		{"-0.1", negativeNumber, 0.0},
		{"", "", 0.0},
	}

	for _, td := range testData {
		gotErrorMessage, gotValue := checkDonation(td.str)
		if td.wantValue != gotValue {
			t.Errorf("%s: want %f got %f", td.str, td.wantValue, gotValue)
		}
		if td.wantErrorMessage != gotErrorMessage {
			t.Errorf("%s: want %s got %s", td.str, td.wantErrorMessage, gotErrorMessage)
		}
	}
}

// TestGetTickBox checks the getTickBox function.
func TestGetTickBox(t *testing.T) {
	var testData = []struct {
		input      string
		wantStatus bool
		wantOutput string
	}{
		{"on", true, "checked"},
		{"off", false, "unchecked"},
		{"foo", false, "unchecked"},
		{"", false, "unchecked"},
	}

	for _, td := range testData {
		gotStatus, gotOutput := getTickBox(td.input)

		if td.wantStatus != gotStatus {
			t.Errorf("%s: want %v got %v", td.input, td.wantStatus, gotStatus)
		}

		if td.wantOutput != gotOutput {
			t.Errorf("%s: want %v got %v", td.input, td.wantOutput, gotOutput)
		}
	}
}

// TestSimpleValidation checks the first stage of validation.
func TestSimpleValidation(t *testing.T) {

	var testData = []struct {
		description                string
		form                       PaymentFormData
		wantValid                  bool
		wantForm                   PaymentFormData
		wantOrdinaryMembershipFee  string
		wantAssociateMembershipFee string
		wantFriendMembershipFee    string
		wantDonationToSociety      string
		wantDonationToMuseum       string
	}{
		{
			"valid - all",
			PaymentFormData{ // This also checks the Trimspace calls.
				Valid:       false,
				PaymentYear: 2024, OrdinaryMembershipFee: 1.2, AssociateMembershipFee: 3.4, FriendMembershipFee: 5.6,
				FirstName: " a\t", LastName: " b ", Email: " a@b.com ", FriendInput: "on",
				DonationToSocietyInput: " 7.83\t", DonationToMuseumInput: " 8.9 ", GiftaidInput: "on",
				AssocFirstName: " f ", AssocLastName: " l ", AssocEmail: "  a@l.com  ", AssociateIsFriendInput: "on",

				Friend: true, DonationToSociety: 7.83, DonationToMuseum: 78.9, Giftaid: true, AssociateIsFriend: true,
				UserID: 0, AssocUserID: 0,
				GeneralErrorMessage: "", FirstNameErrorMessage: "", LastNameErrorMessage: "", EmailErrorMessage: "",
				AssocFirstNameErrorMessage: "", AssocLastNameErrorMessage: "", DonationToSocietyErrorMessage: "", DonationToMuseumErrorMessage: "",
			},
			true,

			PaymentFormData{
				true, 2024, 1.2, 3.4, 5.6,
				"a", "b", "a@b.com", "on",
				"7.83", "8.9", "on",
				"f", "l", "a@l.com", "on",
				true, "checked", 7.83, 8.9, true, "checked", true, "checked",
				0, 0, "", "", "", "", "", "", "", "",
			},
			"1.20", "3.40", "5.60", "7.83", "8.90",
		},

		{
			"valid - no friends",
			PaymentFormData{
				false, 2024, 24.0, 6.0, 5,
				"a", "b", "a@b.com", "off",
				"1.5", "2.5", "on",
				"f", "l", "a@l.com", "off",
				false, "", 0.0, 0.0, false, "", false, "",
				0, 0, "", "", "", "", "", "", "", "",
			},
			true,
			PaymentFormData{
				true, 2024, 24.0, 6.0, 5,
				"a", "b", "a@b.com", "off",
				"1.5", "2.5", "on",
				"f", "l", "a@l.com", "off",
				false, "unchecked", 1.5, 2.5, true, "checked", false, "unchecked",
				0, 0, "", "", "", "", "", "", "", "",
			},
			"24.00", "6.00", "5.00", "1.50", "2.50",
		},
		{
			"valid - no associate",
			PaymentFormData{
				false, 2024, 24.0, 6.0, 5,
				"a", "b", "a@b.com", "on",
				"1.5", "2.5", "on",
				"", "", "", "",
				false, "", 0.0, 0.0, false, "", false, "",
				0, 0, "", "", "", "", "", "", "", "",
			},
			true,
			PaymentFormData{
				true, 2024, 24.0, 6.0, 5,
				"a", "b", "a@b.com", "on",
				"1.5", "2.5", "on",
				"", "", "", "",
				true, "checked", 1.5, 2.5, true, "checked", false, "unchecked",
				0, 0, "", "", "", "", "", "", "", "",
			},
			"24.00", "6.00", "5.00", "1.50", "2.50",
		},
		{
			"valid - ordinary member is not a friend",
			PaymentFormData{
				false, 2024, 24.0, 6.0, 5,
				"a", "b", "a@b.com", "off",
				"1.5", "2.5", "on",
				"f", "l", "a@l.com", "on",
				false, "", 0.0, 0.0, false, "checked", true, "checked",
				0, 0, "", "", "", "", "", "", "", "",
			},
			true,
			PaymentFormData{
				true, 2024, 24.0, 6.0, 5,
				"a", "b", "a@b.com", "off",
				"1.5", "2.5", "on",
				"f", "l", "a@l.com", "on",
				false, "unchecked", 1.5, 2.5, true, "checked", true, "checked",
				0, 0, "", "", "", "", "", "", "", "",
			},
			"24.00", "6.00", "5.00", "1.50", "2.50",
		},
		{
			description: "valid - associate is not friend",
			form: PaymentFormData{
				false, 2024, 24.0, 6.0, 5,
				"a", "b", "a@b.com", "on",
				"1.5", "2.5", "on",
				"f", "l", "a@l.com", "",
				false, "", 0.0, 0.0, false, "", true, "checked",
				0, 0, "", "", "", "", "", "", "", "",
			},
			wantValid: true,
			wantForm: PaymentFormData{
				true, 2024, 24.0, 6.0, 5,
				"a", "b", "a@b.com", "on",
				"1.5", "2.5", "on",
				"f", "l", "a@l.com", "",
				true, "checked", 1.5, 2.5, true, "checked", false, "unchecked",
				0, 0, "", "", "", "", "", "", "", "",
			},
			wantOrdinaryMembershipFee:  "24.00",
			wantAssociateMembershipFee: "6.00",
			wantFriendMembershipFee:    "5.00",
			wantDonationToSociety:      "1.50",
			wantDonationToMuseum:       "2.50",
		},
		{
			description: "valid - friend tick box on, others off",
			form: PaymentFormData{
				PaymentYear: 2024,
				FirstName:   "a", LastName: "b", Email: "a@b.com", FriendInput: "on",
				DonationToSocietyInput: "7.8", DonationToMuseumInput: "8.9", GiftaidInput: "off",
				AssocFirstName: "", AssocLastName: "", AssocEmail: "", AssociateIsFriendInput: "",
				Friend: false, DonationToSociety: 1.234, DonationToMuseum: 5.678, Giftaid: false, AssociateIsFriend: false,
				UserID: 0, AssocUserID: 0,
				GeneralErrorMessage: "", FirstNameErrorMessage: "", LastNameErrorMessage: "", EmailErrorMessage: "",
				AssocFirstNameErrorMessage: "", AssocLastNameErrorMessage: "",
			},
			wantValid: true,
			wantForm: PaymentFormData{
				Valid:       true,
				PaymentYear: 2024, OrdinaryMembershipFee: 0, AssociateMembershipFee: 0, FriendMembershipFee: 0,
				FirstName: "a", LastName: "b", Email: "a@b.com", FriendInput: "on",
				DonationToSocietyInput: "7.8", DonationToMuseumInput: "8.9", GiftaidInput: "off",
				AssocFirstName: "", AssocLastName: "", AssocEmail: "", AssociateIsFriendInput: "",
				Friend: true, FriendOutput: "checked", DonationToSociety: 7.8, DonationToMuseum: 8.9,
				Giftaid: false, GiftaidOutput: "unchecked", AssociateIsFriend: false, AssociateIsFriendOutput: "unchecked",
				UserID: 0, AssocUserID: 0,
				GeneralErrorMessage: "", FirstNameErrorMessage: "", LastNameErrorMessage: "", EmailErrorMessage: "",
				AssocFirstNameErrorMessage: "", AssocLastNameErrorMessage: "",
			},
			wantOrdinaryMembershipFee:  "0.00",
			wantAssociateMembershipFee: "0.00",
			wantFriendMembershipFee:    "0.00",
			wantDonationToSociety:      "7.80",
			wantDonationToMuseum:       "8.90",
		},
		{
			description: "valid - assoc friend tick box on, others off",
			form: PaymentFormData{
				PaymentYear: 2024, OrdinaryMembershipFee: 1.234, AssociateMembershipFee: 3.456, FriendMembershipFee: 5.678,
				FirstName: "a", LastName: "b", Email: "a@b.com", FriendInput: "off",
				DonationToSocietyInput: "7.8", DonationToMuseumInput: "8.9", GiftaidInput: "",
				AssocFirstName: "f", AssocLastName: "l", AssocEmail: "", AssociateIsFriendInput: "on",
				Friend: false, DonationToSociety: 0.0, DonationToMuseum: 0.0, Giftaid: false, AssociateIsFriend: false,
				UserID: 0, AssocUserID: 0,
				GeneralErrorMessage: "", FirstNameErrorMessage: "", LastNameErrorMessage: "", EmailErrorMessage: "",
				AssocFirstNameErrorMessage: "", AssocLastNameErrorMessage: "",
			},
			wantValid: true,
			wantForm: PaymentFormData{
				Valid:       true,
				PaymentYear: 2024, OrdinaryMembershipFee: 1.234, AssociateMembershipFee: 3.456, FriendMembershipFee: 5.678,
				FirstName: "a", LastName: "b", Email: "a@b.com", FriendInput: "off",
				DonationToSocietyInput: "7.8", DonationToMuseumInput: "8.9",
				Giftaid: false, GiftaidOutput: "unchecked", AssociateIsFriend: true, AssociateIsFriendOutput: "checked",
				AssocFirstName: "f", AssocLastName: "l", AssocEmail: "", AssociateIsFriendInput: "on",
				Friend: false, FriendOutput: "unchecked", DonationToSociety: 7.8, DonationToMuseum: 8.9,

				UserID: 0, AssocUserID: 0,
				GeneralErrorMessage: "", FirstNameErrorMessage: "", LastNameErrorMessage: "", EmailErrorMessage: "",
				AssocFirstNameErrorMessage: "", AssocLastNameErrorMessage: "",
			},
			wantOrdinaryMembershipFee:  "1.23",
			wantAssociateMembershipFee: "3.46",
			wantFriendMembershipFee:    "5.68",
			wantDonationToSociety:      "7.80",
			wantDonationToMuseum:       "8.90",
		},
		{
			description: "valid - giftaid tick box on, others off",
			form: PaymentFormData{
				PaymentYear: 2024, OrdinaryMembershipFee: 1.2, AssociateMembershipFee: 3.4, FriendMembershipFee: 5.6,
				FirstName: "a", LastName: "b", Email: "a@b.com", FriendInput: "off",
				DonationToSocietyInput: "7.8", DonationToMuseumInput: "8.9", GiftaidInput: "on",
				AssocFirstName: "", AssocLastName: "", AssocEmail: "", AssociateIsFriendInput: "",
				Friend: false, DonationToSociety: 0.0, DonationToMuseum: 0.0, Giftaid: false, AssociateIsFriend: false,
				UserID: 0, AssocUserID: 0,
				GeneralErrorMessage: "", FirstNameErrorMessage: "", LastNameErrorMessage: "", EmailErrorMessage: "",
				AssocFirstNameErrorMessage: "", AssocLastNameErrorMessage: "",
			},
			wantValid: true,
			wantForm: PaymentFormData{
				Valid:       true,
				PaymentYear: 2024, OrdinaryMembershipFee: 1.2, AssociateMembershipFee: 3.4, FriendMembershipFee: 5.6,
				FirstName: "a", LastName: "b", Email: "a@b.com", FriendInput: "off",
				DonationToSocietyInput: "7.8", DonationToMuseumInput: "8.9", GiftaidInput: "on",
				AssocFirstName: "", AssocLastName: "", AssocEmail: "", AssociateIsFriendInput: "",
				Friend: false, FriendOutput: "unchecked", DonationToSociety: 7.8, DonationToMuseum: 8.9,
				Giftaid: true, GiftaidOutput: "checked", AssociateIsFriend: false, AssociateIsFriendOutput: "unchecked",
				UserID: 0, AssocUserID: 0,
				GeneralErrorMessage: "", FirstNameErrorMessage: "", LastNameErrorMessage: "", EmailErrorMessage: "",
				AssocFirstNameErrorMessage: "", AssocLastNameErrorMessage: "",
			},
			wantOrdinaryMembershipFee:  "1.20",
			wantAssociateMembershipFee: "3.40",
			wantFriendMembershipFee:    "5.60",
			wantDonationToSociety:      "7.80",
			wantDonationToMuseum:       "8.90",
		},
		{
			"valid - no giftaid",
			PaymentFormData{
				false, 2024, 24.0, 6.0, 5,
				"a", "b", "a@b.com", "on",
				"1.5", "2.5", "off",
				"f", "l", "a@l.com", "on",
				false, "", 0.0, 0.0, true, "", false, "", // These should all be updated.
				0, 0, "", "", "", "", "", "", "", "",
			},
			true,
			PaymentFormData{
				true, 2024, 24.0, 6.0, 5,
				"a", "b", "a@b.com", "on",
				"1.5", "2.5", "off",
				"f", "l", "a@l.com", "on",
				true, "checked", 1.5, 2.5, false, "unchecked", true, "checked",
				0, 0, "", "", "", "", "", "", "", "",
			},
			"24.00", "6.00", "5.00", "1.50", "2.50",
		},
		{
			"ordinary member first name missing",
			PaymentFormData{
				false, 2024, 24.0, 6.0, 5.0,
				"", "b", "a@b.com", "on",
				"99.989", "11.111", "off",
				"f", "l", "a@l.com", "",
				false, "", 0.0, 0.0, false, "", true, "",
				0, 0, "", "", "", "", "", "", "", "",
			},
			false,
			PaymentFormData{
				false, 2024, 24.0, 6.0, 5.0,
				"", "b", "a@b.com", "on",
				"99.989", "11.111", "off",
				"f", "l", "a@l.com", "",
				true, "checked", 99.989, 11.111, false, "unchecked", false, "unchecked",
				0, 0, "", firstNameErrorMessage, "", "", "", "", "", "",
			},
			"24.00", "6.00", "5.00", "99.99", "11.11",
		},

		{
			"ordinary member last name missing",
			PaymentFormData{
				PaymentYear: 2024, OrdinaryMembershipFee: 1.2, AssociateMembershipFee: 3.4, FriendMembershipFee: 5.6,
				FirstName: " a\t", LastName: "", Email: " a@b.com ", FriendInput: "on",
				DonationToSocietyInput: " 7.8\t", DonationToMuseumInput: " 8.9 ", GiftaidInput: "on",
				AssocFirstName: " f ", AssocLastName: " l ", AssocEmail: "  a@l.com  ", AssociateIsFriendInput: "on",
				Friend: true, DonationToSociety: 0.0, DonationToMuseum: 0.0, Giftaid: true, AssociateIsFriend: true,
				UserID: 0, AssocUserID: 0,
				GeneralErrorMessage: "", FirstNameErrorMessage: "", LastNameErrorMessage: "", EmailErrorMessage: "",
				AssocFirstNameErrorMessage: "", AssocLastNameErrorMessage: "",
			},
			false,
			PaymentFormData{
				PaymentYear: 2024, OrdinaryMembershipFee: 1.2, AssociateMembershipFee: 3.4, FriendMembershipFee: 5.6,
				FirstName: "a", LastName: "", Email: "a@b.com", FriendInput: "on",
				DonationToSocietyInput: "7.8", DonationToMuseumInput: "8.9", GiftaidInput: "on",
				AssocFirstName: "f", AssocLastName: "l", AssocEmail: "a@l.com", AssociateIsFriendInput: "on",
				Friend: true, FriendOutput: "checked", DonationToSociety: 7.8, DonationToMuseum: 8.9,
				Giftaid: true, GiftaidOutput: "checked", AssociateIsFriend: true, AssociateIsFriendOutput: "checked",
				UserID: 0, AssocUserID: 0,
				GeneralErrorMessage: "", FirstNameErrorMessage: "", LastNameErrorMessage: lastNameErrorMessage, EmailErrorMessage: "",
				AssocFirstNameErrorMessage: "", AssocLastNameErrorMessage: "",
			},
			"1.20", "3.40", "5.60", "7.80", "8.90",
		},
		{
			"ordinary member email missing",
			PaymentFormData{
				PaymentYear: 2024, OrdinaryMembershipFee: 1.2, AssociateMembershipFee: 3.4, FriendMembershipFee: 5.6,
				FirstName: " a\t", LastName: "b", Email: "", FriendInput: "on",
				DonationToSocietyInput: " 7.8\t", DonationToMuseumInput: " 8.9 ", GiftaidInput: "on",
				AssocFirstName: " f ", AssocLastName: " l ", AssocEmail: "  a@l.com  ", AssociateIsFriendInput: "on",
				Friend: true, DonationToSociety: 0.0, DonationToMuseum: 0.0, Giftaid: true, AssociateIsFriend: true,
				UserID: 0, AssocUserID: 0,
				GeneralErrorMessage: "", FirstNameErrorMessage: "", LastNameErrorMessage: "", EmailErrorMessage: "",
				AssocFirstNameErrorMessage: "", AssocLastNameErrorMessage: "",
			},
			false,
			PaymentFormData{
				PaymentYear: 2024, OrdinaryMembershipFee: 1.2, AssociateMembershipFee: 3.4, FriendMembershipFee: 5.6,
				FirstName: "a", LastName: "b", Email: "", FriendInput: "on",
				DonationToSocietyInput: "7.8", DonationToMuseumInput: "8.9", GiftaidInput: "on",
				AssocFirstName: "f", AssocLastName: "l", AssocEmail: "a@l.com", AssociateIsFriendInput: "on",
				Friend: true, FriendOutput: "checked", DonationToSociety: 7.8, DonationToMuseum: 8.9,
				Giftaid: true, GiftaidOutput: "checked", AssociateIsFriend: true, AssociateIsFriendOutput: "checked",
				UserID: 0, AssocUserID: 0,
				GeneralErrorMessage: "", FirstNameErrorMessage: "", LastNameErrorMessage: "", EmailErrorMessage: emailErrorMessage,
				AssocFirstNameErrorMessage: "", AssocLastNameErrorMessage: "",
			},
			"1.20", "3.40", "5.60", "7.80", "8.90",
		},
		{
			"associate member first name missing",
			PaymentFormData{
				PaymentYear: 2024, OrdinaryMembershipFee: 1.2, AssociateMembershipFee: 3.4, FriendMembershipFee: 5.6,
				FirstName: " a\t", LastName: "b", Email: "a@b.com", FriendInput: "on",
				DonationToSocietyInput: " 7.8\t", DonationToMuseumInput: " 8.9 ", GiftaidInput: "on",
				AssocFirstName: "", AssocLastName: " l ", AssocEmail: "  a@l.com  ", AssociateIsFriendInput: "on",
				Friend: true, DonationToSociety: 0.0, DonationToMuseum: 0.0, Giftaid: true, AssociateIsFriend: true,
				UserID: 0, AssocUserID: 0,
				GeneralErrorMessage: "", FirstNameErrorMessage: "", LastNameErrorMessage: "", EmailErrorMessage: "",
				AssocFirstNameErrorMessage: "", AssocLastNameErrorMessage: "",
			},
			false,
			PaymentFormData{
				PaymentYear: 2024, OrdinaryMembershipFee: 1.2, AssociateMembershipFee: 3.4, FriendMembershipFee: 5.6,
				FirstName: "a", LastName: "b", Email: "a@b.com", FriendInput: "on",
				DonationToSocietyInput: "7.8", DonationToMuseumInput: "8.9", GiftaidInput: "on",
				AssocFirstName: "", AssocLastName: "l", AssocEmail: "a@l.com", AssociateIsFriendInput: "on",
				Friend: true, FriendOutput: "checked", DonationToSociety: 7.8, DonationToMuseum: 8.9,
				Giftaid: true, GiftaidOutput: "checked", AssociateIsFriend: true, AssociateIsFriendOutput: "checked",
				UserID: 0, AssocUserID: 0,
				GeneralErrorMessage: "", FirstNameErrorMessage: "", LastNameErrorMessage: "", EmailErrorMessage: "",
				AssocFirstNameErrorMessage: assocFirstNameErrorMessage, AssocLastNameErrorMessage: "",
			},
			"1.20", "3.40", "5.60", "7.80", "8.90",
		},
		{
			"associate member last name missing",
			PaymentFormData{
				PaymentYear: 2024, OrdinaryMembershipFee: 1.2, AssociateMembershipFee: 3.4, FriendMembershipFee: 5.6,
				FirstName: "a", LastName: "b", Email: "a@b.com", FriendInput: "on",
				DonationToSocietyInput: " 7.8\t", DonationToMuseumInput: "8.9", GiftaidInput: "on",
				AssocFirstName: "f", AssocLastName: "", AssocEmail: "a@l.com", AssociateIsFriendInput: "on",
				Friend: true, DonationToSociety: 0.0, DonationToMuseum: 0.0, Giftaid: true, AssociateIsFriend: true,
				UserID: 0, AssocUserID: 0,
				GeneralErrorMessage: "", FirstNameErrorMessage: "", LastNameErrorMessage: "", EmailErrorMessage: "",
				AssocFirstNameErrorMessage: "", AssocLastNameErrorMessage: "",
			},
			false,
			PaymentFormData{
				PaymentYear: 2024, OrdinaryMembershipFee: 1.2, AssociateMembershipFee: 3.4, FriendMembershipFee: 5.6,
				FirstName: "a", LastName: "b", Email: "a@b.com", FriendInput: "on",
				DonationToSocietyInput: "7.8", DonationToMuseumInput: "8.9", GiftaidInput: "on",
				AssocFirstName: "f", AssocLastName: "", AssocEmail: "a@l.com", AssociateIsFriendInput: "on",
				Friend: true, FriendOutput: "checked", DonationToSociety: 7.8, DonationToMuseum: 8.9,
				Giftaid: true, GiftaidOutput: "checked", AssociateIsFriend: true, AssociateIsFriendOutput: "checked",
				UserID: 0, AssocUserID: 0,
				GeneralErrorMessage: "", FirstNameErrorMessage: "", LastNameErrorMessage: "", EmailErrorMessage: "",
				AssocFirstNameErrorMessage: "", AssocLastNameErrorMessage: assocLastNameErrorMessage,
			},
			"1.20", "3.40", "5.60", "7.80", "8.90",
		},

		{
			"associate member but no ordinary member",
			PaymentFormData{
				PaymentYear: 2024, OrdinaryMembershipFee: 1.2, AssociateMembershipFee: 3.4, FriendMembershipFee: 5.6,
				FirstName: "", LastName: "", Email: "", FriendInput: "",
				DonationToSocietyInput: " 7.8\t", DonationToMuseumInput: "8.9", GiftaidInput: "",
				AssocFirstName: "f", AssocLastName: "l", AssocEmail: "a@l.com", AssociateIsFriendInput: "on",
				Friend: false, DonationToSociety: 0.0, DonationToMuseum: 0.0, Giftaid: false, AssociateIsFriend: false,
				UserID: 0, AssocUserID: 0,
				GeneralErrorMessage: "", FirstNameErrorMessage: "", LastNameErrorMessage: "", EmailErrorMessage: "",
				AssocFirstNameErrorMessage: "", AssocLastNameErrorMessage: "",
			},
			false,
			PaymentFormData{
				PaymentYear: 2024, OrdinaryMembershipFee: 1.2, AssociateMembershipFee: 3.4, FriendMembershipFee: 5.6,
				FirstName: "", LastName: "", Email: "", FriendInput: "",
				DonationToSocietyInput: "7.8", DonationToMuseumInput: "8.9", GiftaidInput: "",
				AssocFirstName: "f", AssocLastName: "l", AssocEmail: "a@l.com", AssociateIsFriendInput: "on",
				Friend: false, FriendOutput: "unchecked", DonationToSociety: 7.8, DonationToMuseum: 8.9,
				Giftaid: false, GiftaidOutput: "unchecked", AssociateIsFriend: true, AssociateIsFriendOutput: "checked",
				UserID: 0, AssocUserID: 0,
				GeneralErrorMessage: "", FirstNameErrorMessage: firstNameErrorMessage,
				LastNameErrorMessage: lastNameErrorMessage, EmailErrorMessage: emailErrorMessage,
				AssocFirstNameErrorMessage: "", AssocLastNameErrorMessage: "",
			},
			"1.20", "3.40", "5.60", "7.80", "8.90",
		},

		{
			"associate email address but associate member's name missing",
			PaymentFormData{
				PaymentYear: 2024, OrdinaryMembershipFee: 1.2, AssociateMembershipFee: 3.4, FriendMembershipFee: 5.6,
				FirstName: "a", LastName: "b", Email: "a@b.com", FriendInput: "on",
				DonationToSocietyInput: "7.8", DonationToMuseumInput: "8.9", GiftaidInput: "on",
				AssocFirstName: "", AssocLastName: "", AssocEmail: "a@l.com", AssociateIsFriendInput: "on",
				Friend: true, DonationToSociety: 0.0, DonationToMuseum: 0.0, Giftaid: true, AssociateIsFriend: true,
				UserID: 0, AssocUserID: 0,
				GeneralErrorMessage: "", FirstNameErrorMessage: "", LastNameErrorMessage: "", EmailErrorMessage: "",
				AssocFirstNameErrorMessage: "", AssocLastNameErrorMessage: "",
			},
			false,
			PaymentFormData{
				PaymentYear: 2024, OrdinaryMembershipFee: 1.2, AssociateMembershipFee: 3.4, FriendMembershipFee: 5.6,
				FirstName: "a", LastName: "b", Email: "a@b.com", FriendInput: "on",
				DonationToSocietyInput: "7.8", DonationToMuseumInput: "8.9", GiftaidInput: "on",
				AssocFirstName: "", AssocLastName: "", AssocEmail: "a@l.com", AssociateIsFriendInput: "on",
				Friend: true, FriendOutput: "checked", DonationToSociety: 7.8, DonationToMuseum: 8.9,
				Giftaid: true, GiftaidOutput: "checked", AssociateIsFriend: true, AssociateIsFriendOutput: "checked",
				UserID: 0, AssocUserID: 0,
				GeneralErrorMessage: "", FirstNameErrorMessage: "", LastNameErrorMessage: "", EmailErrorMessage: "",
				AssocFirstNameErrorMessage: assocFirstNameErrorMessage, AssocLastNameErrorMessage: assocLastNameErrorMessage,
			},
			"1.20", "3.40", "5.60", "7.80", "8.90",
		},
		{
			"associate friend tick box but associate member's name missing",
			PaymentFormData{
				PaymentYear: 2024, OrdinaryMembershipFee: 1.2, AssociateMembershipFee: 3.4, FriendMembershipFee: 5.6,
				FirstName: "a", LastName: "b", Email: "a@b.com", FriendInput: "on",
				DonationToSocietyInput: "7.8", DonationToMuseumInput: "8.9", GiftaidInput: "on",
				AssocFirstName: "", AssocLastName: "", AssocEmail: "", AssociateIsFriendInput: "on",
				Friend: false, DonationToSociety: 0.0, DonationToMuseum: 0.0, Giftaid: false, AssociateIsFriend: false,
				UserID: 0, AssocUserID: 0,
				GeneralErrorMessage: "", FirstNameErrorMessage: "", LastNameErrorMessage: "", EmailErrorMessage: "",
				AssocFirstNameErrorMessage: "", AssocLastNameErrorMessage: "",
			},
			false,
			PaymentFormData{
				PaymentYear: 2024, OrdinaryMembershipFee: 1.2, AssociateMembershipFee: 3.4, FriendMembershipFee: 5.6,
				FirstName: "a", LastName: "b", Email: "a@b.com", FriendInput: "on",
				DonationToSocietyInput: "7.8", DonationToMuseumInput: "8.9", GiftaidInput: "on",
				AssocFirstName: "", AssocLastName: "", AssocEmail: "", AssociateIsFriendInput: "on",
				Friend: true, FriendOutput: "checked", DonationToSociety: 7.8, DonationToMuseum: 8.9,
				Giftaid: true, GiftaidOutput: "checked", AssociateIsFriend: true, AssociateIsFriendOutput: "checked",
				UserID: 0, AssocUserID: 0,
				GeneralErrorMessage: "", FirstNameErrorMessage: "", LastNameErrorMessage: "", EmailErrorMessage: "",
				AssocFirstNameErrorMessage: assocFirstNameErrorMessage, AssocLastNameErrorMessage: assocLastNameErrorMessage,
			},
			"1.20", "3.40", "5.60", "7.80", "8.90",
		},

		{
			"donation to society invalid number",
			PaymentFormData{
				false, 2024, 24.0, 6.0, 5.0,
				"a", "b", "a@b.com", "on",
				"junk", "4.5", "on",
				"f", "l", "a@l.com", "on",
				false, "", 0.0, 0.0, false, "", false, "",
				0, 0, "", "", "", "", "", "", "", "",
			},
			false,
			PaymentFormData{
				false, 2024, 24.0, 6.0, 5.0,
				"a", "b", "a@b.com", "on",
				"junk", "4.5", "on",
				"f", "l", "a@l.com", "on",
				true, "checked", 0.0, 4.5, true, "checked", true, "checked",
				0, 0, "", "", "", "", invalidNumber, "", "", "",
			},
			"24.00", "6.00", "5.00", "", "4.50",
		},
		{
			"donation to museum invalid number",
			PaymentFormData{
				false, 2024, 24.0, 6.0, 5.0,
				"a", "b", "a@b.com", "off",
				"1.5", "junk", "",
				"f", "l", "a@l.com", "",
				false, "", 0.0, 0.0, false, "", false, "",
				0, 0, "", "", "", "", "", "", "", "",
			},
			false,
			PaymentFormData{
				false, 2024, 24.0, 6.0, 5.0,
				"a", "b", "a@b.com", "off",
				"1.5", "junk", "",
				"f", "l", "a@l.com", "",
				false, "unchecked", 1.5, 0.0, false, "unchecked", false, "unchecked",
				0, 0, "", "", "", "", "", invalidNumber, "", "",
			},
			"24.00", "6.00", "5.00", "1.50", "",
		},
	}

	for _, td := range testData {
		valid := simpleValidate(&td.form)

		// The tick box outputs have been set so the got and want forms are no longer equal.
		// Check the values and then unset them.

		if td.wantForm.FriendOutput != td.form.FriendOutput {
			t.Errorf("%s: want %s got %s", td.description, td.wantForm.FriendOutput, td.form.FriendOutput)
		}

		if td.wantForm.AssociateIsFriendOutput != td.form.AssociateIsFriendOutput {
			t.Errorf("%s: want %s got %s",
				td.description, td.wantForm.AssociateIsFriendOutput, td.form.AssociateIsFriendOutput)
		}

		if td.wantForm.GiftaidOutput != td.form.GiftaidOutput {
			t.Errorf("%s: want %s got %s", td.description, td.wantForm.GiftaidOutput, td.form.GiftaidOutput)
		}

		if td.wantValid != valid {
			t.Errorf("%s want %v got %v", td.description, td.wantValid, valid)
		}

		if td.wantForm != td.form {
			fmt.Printf("%s:\nwant %v\n got %v\n", td.description, td.wantForm, td.form)
		}

		if td.wantOrdinaryMembershipFee != td.form.OrdinaryMembershipFeeForDisplay() {
			fmt.Printf("%s:\nordinary fee - want %v\n got %v\n", td.description, td.wantOrdinaryMembershipFee, td.form.OrdinaryMembershipFeeForDisplay())
		}

		if td.wantAssociateMembershipFee != td.form.AssociateMembershipFeeForDisplay() {
			fmt.Printf("%s:\nassoc fee - want %v\n got %v\n", td.description, td.wantAssociateMembershipFee, td.form.AssociateMembershipFeeForDisplay())
		}

		if td.wantFriendMembershipFee != td.form.FriendMembershipFeeForDisplay() {
			fmt.Printf("%s:\nfriend fee - want %v\n got %v\n", td.description, td.wantFriendMembershipFee, td.form.FriendMembershipFeeForDisplay())
		}

		if td.wantDonationToSociety != td.form.DonationToSocietyForDisplay() {
			fmt.Printf("%s:\ndonation to society - want %v\n got %v\n", td.description, td.wantDonationToSociety, td.form.DonationToSocietyForDisplay())
		}

		if td.wantDonationToMuseum != td.form.DonationToMuseumForDisplay() {
			fmt.Printf("%s:\ndonation to museum - want %v\n got %v\n", td.description, td.wantDonationToMuseum, td.form.DonationToMuseumForDisplay())
		}
	}
}
