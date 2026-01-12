// forms contains the structures used to display HTML pages.
package forms

import (
	"fmt"

	"github.com/goblimey/go-stripe-payments/code/pkg/config"
)

// SaleForm holds the data about the sale during initial validation.
type SaleForm struct {

	// Valid is set false during validation if the form data is invalid.
	Valid bool

	// Reference Data.
	OrganisationName       string  // The name of the organisation (for the payment page).
	MembershipYear         int     // The payment year displayed in the title.
	EnableOtherMemberTypes bool    // Enable associate members, friends etc.
	OrdinaryMemberFee      float64 // Ordinary membership fee.
	AssocMemberFee         float64 // Associate membership system.
	FriendFee              float64 // Fee to be a friend of the museum.

	EnableGiftaid bool // Enable giftaid (for UK charities).

	// Data for validation.
	Title                  string `json:"title"`
	FirstName              string `json:"first_name"`
	LastName               string `json:"last_name"`
	Email                  string `json:"email"`
	FriendInput            string `json:"friend"`              // tickbox - "on" or "off"
	DonationToSocietyInput string `json:"donation_to_society"` // number
	DonationToMuseumInput  string `json:"donation_to_museum"`  // number
	GiftaidInput           string `json:"giftaid"`             // tickbox - "on" or "off"
	AssocTitle             string `json:"assoc_title"`
	AssocFirstName         string `json:"assoc_first_name"`
	AssocLastName          string `json:"assoc_last_name"`
	AssocEmail             string `json:"assoc_email"`
	AssocFriendInput       string `json:"assoc_friend"` // tickbox  - "on" or "off"

	//  Values set during validation.
	Friend         bool    // True if the ordinary member's Friend tickbox is valid and true.
	FriendFeeToPay float64 // The friend fee.  (0.0 if not a friend).
	// form.FriendOutput is marked by tyhe compiler as not used.  It's used in the
	// paymentPageTemplateStr but the compiler can't see that.
	FriendOutput      string  // To preset checkbox - "checked" or "unchecked"
	DonationToSociety float64 // Donation to the society.
	DonationToMuseum  float64 // Donation to the museum.
	Giftaid           bool    // True if the giftaid tickbox is valid and true.
	GiftaidOutput     string  // Checkbox setting - "checked" or "unchecked"
	AssocFriend       bool    // True if the associate member's friend tickbox is valid and true.
	AssocFeeToPay     float64 // The associate member fee. (0.0 if no associate.)
	// form.AssocFriendOutput is marked by the compiler as not used.  It's used
	// in the paymentPageTemplateStr but the compiler can't see that.
	AssocFriendOutput   string  // To preset checkbox - "checked" or "unchecked"
	AssocFriendFeeToPay float64 // The fee to be paid for the associate to be a friend (0 if no associate or not a friend).
	UserID              int64   // The ID of the ordinary member in the database (> zero).
	AssocUserID         int64   // The ID of the associate member in the database (zero if no associate).

	// Error messages set if the form data is invalid.
	GeneralErrorMessage           string // Set on a fatal error, eg database connection failure.
	TitleErrorMessage             string
	FirstNameErrorMessage         string
	LastNameErrorMessage          string
	EmailErrorMessage             string
	DonationToSocietyErrorMessage string
	DonationToMuseumErrorMessage  string
	AssocTitleErrorMessage        string
	AssocFirstNameErrorMessage    string
	AssocLastNameErrorMessage     string
}

func NewSaleForm(c *config.Config, membershipYear int) *SaleForm {
	sf := SaleForm{
		OrganisationName:       c.OrganisationName,
		MembershipYear:         membershipYear,
		EnableOtherMemberTypes: c.EnableOtherMemberTypes,
		EnableGiftaid:          c.EnableGiftaid,
		OrdinaryMemberFee:      c.OrdinaryMemberFee,
		AssocMemberFee:         c.AssocMemberFee,
		FriendFee:              c.FriendFee,
	}

	return &sf
}

// MarkMandatoryFields marks the mandatory parameters in a
// payment form by setting error messages containing asterisks.
// This drives the first view of the payment page.
func (sf *SaleForm) MarkMandatoryFields() {
	sf.FirstNameErrorMessage = "*"
	sf.LastNameErrorMessage = "*"
	sf.EmailErrorMessage = "*"
}

// Total calculates and returns the total cost of the purchase.  It's used in HTML
// templates so is parameterless and single-valued.  To allow for free membership,
// if the ordinary member fee is zero, the result is always zero.  (In fact, this
// function should not even be called in that case.)  To guard against an attack
// that injects subversive data into the form such as negative numbers, if any
// values are obviously illegal, the result is zero, which never happens with real
// data.
func (sf *SaleForm) Total() float64 {
	switch {
	case sf.OrdinaryMemberFee <= 0:
		return 0
	case sf.FriendFeeToPay < 0:
		return 0
	case sf.AssocFeeToPay < 0:
		return 0
	case sf.AssocFriendFeeToPay < 0:
		return 0
	case sf.DonationToSociety < 0:
		return 0
	case sf.DonationToMuseum < 0:
		return 0
	}

	// A fee is charged for ordinary membership and the incoming data looks
	// legal.  Calculate the total.
	total := sf.OrdinaryMemberFee +
		sf.FriendFeeToPay +
		sf.DonationToSociety +
		sf.DonationToMuseum +
		sf.AssocFeeToPay +
		sf.AssocFriendFeeToPay

	if total < sf.OrdinaryMemberFee {
		// The total is less than the ordinary membership fee , which should
		// never happen.  It implies that the above checks for subversive data
		// are not sufficient and somebody has invented something that has
		// defeated them.
		return 0
	}

	return total
}

func (sf *SaleForm) TotalForDisplay() string {

	total := sf.Total()

	if total == 0 {
		return ""
	}

	return CostForDisplay(total)
}

// OrdinaryMembershipFeeForDisplay gets the ordinary membership fee
// for a display - a number to two decimal places.
func (sf *SaleForm) OrdinaryMemberFeeForDisplay() string {
	if sf.OrdinaryMemberFee == 0 {
		return ""
	}
	return CostForDisplay(sf.OrdinaryMemberFee)
}

// OrdinaryMemberFriendFeeForDisplay gets the ordinary member's
// museum friend fee for display - a number to two decimal places.
// If the member is not a friend, it returns "0.0".
func (sf *SaleForm) OrdinaryMemberFriendFeeForDisplay() string {

	if !sf.Friend {
		return ""
	}

	if sf.FriendFeeToPay == 0 {
		return ""
	}

	return CostForDisplay(sf.FriendFeeToPay)
}

// FriendFeeForDisplay gets the friend fee for display - a number to two decimal places
// or "" for a zero value.
func (sf *SaleForm) FriendFeeForDisplay() string {

	if sf.FriendFee == 0 {
		return ""
	}

	return CostForDisplay(sf.FriendFee)
}

// DonationToMuseumForDisplay gets the donation to museum
// for a display - a number to two decimal places.
func (sf *SaleForm) DonationToMuseumForDisplay() string {
	if sf.DonationToMuseum == 0 {
		return ""
	}
	return CostForDisplay(sf.DonationToMuseum)
}

// DonationToSocietyForDisplay gets the donation to the society
// for a display - a number to two decimal places.
func (sf *SaleForm) DonationToSocietyForDisplay() string {
	if sf.DonationToSociety == 0 {
		return ""
	}
	return CostForDisplay(sf.DonationToSociety)
}

// AssociateMemberFeeForDisplay gets the associate membership fee
// for display - a number to two decimal places.  If the value is
// zero it returns an empty string.
func (sf *SaleForm) AssocFeeForDisplay() string {

	if sf.AssocMemberFee == 0 {
		return ""
	}

	return CostForDisplay(sf.AssocMemberFee)
}

// AssocFriendFeeForDisplay gets the associate member's
// museum friend fee for display - a number to two decimal places.
// If there is no associate or the associate is not a friend, it
// returns "0.0".
func (sf *SaleForm) AssocFriendFeeForDisplay() string {

	if !sf.AssocFriend {
		return ""
	}

	if sf.AssocFeeToPay == 0 {
		return ""
	}

	return CostForDisplay(sf.AssocFriendFeeToPay)
}

// CostForDisplay produces the given number in a form suitable for display as
// a cost - the currency symbol followed by the number to two decimal places.
func CostForDisplay(v float64) string {
	if v == 0 {
		return ""
	}
	return fmt.Sprintf("£%.2f", v)
}
