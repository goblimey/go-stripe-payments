// forms contains the structures used to display HTML pages.
package forms

import (
	"fmt"
	"time"

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
	FirstName              string
	LastName               string
	Email                  string
	FriendInput            string  // tickbox - "on" or "off"
	FriendFeeToPay         float64 // The friend fee.  (0 if not a friend).
	DonationToSocietyInput string  // number
	DonationToMuseumInput  string  // number
	GiftaidInput           string  // tickbox - "on" or "off"
	AssocFirstName         string
	AssocLastName          string
	AssocEmail             string
	AssocFriendInput       string // tickbox  - "on" or "off"

	//  Values set during validation.
	Friend bool // True if the ordinary member's Friend tickbox is valid and true.
	// form.FriendOutput is marked as not used.  It's used in the
	// paymentPageTemplateStr but the compiler can't see that.
	FriendOutput      string  // To preset checkbox - "checked" or "unchecked"
	DonationToSociety float64 // Donation to the society.
	DonationToMuseum  float64 // Donation to the museum.
	Giftaid           bool    // True if the giftaid tickbox is valid and true.
	GiftaidOutput     string  // Checkbox setting - "checked" or "unchecked"
	AssocFriend       bool    // True if the associate member's friend tickbox is valid and true.
	AssocFeeToPay     float64 // The associate member fee. (0 if no associate.)
	// form.AssocFriendOutput is marked by the compiler as not used.  It's used
	// in the paymentPageTemplateStr but the compiler can't see that.
	AssocFriendOutput   string  // To preset checkbox - "checked" or "unchecked"
	AssocFriendFeeToPay float64 // The fee to be paid for the associate to be a friend (0 if no associate or not a friend).
	UserID              int64   // The ID of the ordinary member in the database (> zero).
	AssocUserID         int64   // The ID of the associate member in the database (zero if no associate).

	// Error messages set if the form data is invalid.
	GeneralErrorMessage           string // Set on a fatal error, eg database connection failure.
	FirstNameErrorMessage         string
	LastNameErrorMessage          string
	EmailErrorMessage             string
	DonationToSocietyErrorMessage string
	DonationToMuseumErrorMessage  string
	AssocFirstNameErrorMessage    string
	AssocLastNameErrorMessage     string
}

func NewSaleForm(c *config.Config, membershipYear int, now time.Time) *SaleForm {
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
		// never happen.  It implies that the above checks for subverive data
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

	return fmt.Sprintf("£%.2f", total)
}

// OrdinaryMembershipFeeForDisplay gets the ordinary membership fee
// for a display - a number to two decimal places.
func (sf *SaleForm) OrdinaryMemberFeeForDisplay() string {
	if sf.OrdinaryMemberFee == 0 {
		return ""
	}
	return fmt.Sprintf("£%.2f", sf.OrdinaryMemberFee)
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

	return fmt.Sprintf("£%.2f", sf.FriendFeeToPay)
}

// FriendFeeForDisplay gets the friend fee for display - a number to two decimal places
// or "" for a zero value.
func (sf *SaleForm) FriendFeeForDisplay() string {

	if sf.FriendFeeToPay == 0 {
		return ""
	}

	return fmt.Sprintf("£%.2f", sf.FriendFeeToPay)
}

// DonationToMuseumForDisplay gets the donation to museum
// for a display - a number to two decimal places.
func (sf *SaleForm) DonationToMuseumForDisplay() string {
	if sf.DonationToMuseum == 0 {
		return ""
	}
	return fmt.Sprintf("£%.2f", sf.DonationToMuseum)
}

// DonationToSocietyForDisplay gets the donation to the society
// for a display - a number to two decimal places.
func (sf *SaleForm) DonationToSocietyForDisplay() string {
	if sf.DonationToSociety == 0 {
		return ""
	}
	return fmt.Sprintf("£%.2f", sf.DonationToSociety)
}

// AssociateMemberFeeForDisplay gets the associate membership fee
// for display - a number to two decimal places.  If the value is
// zero or there is no associate, it returns "".
func (sf *SaleForm) AssocFeeForDisplay() string {

	if sf.AssocFeeToPay == 0 {
		return ""
	}

	return fmt.Sprintf("£%.2f", sf.AssocFeeToPay)
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

	return fmt.Sprintf("£%.2f", sf.AssocFriendFeeToPay)
}
