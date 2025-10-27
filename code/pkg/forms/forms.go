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

	if sf.FriendFee == 0 {
		return ""
	}

	return fmt.Sprintf("£%.2f", sf.FriendFee)
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
// zero it returns an empty string.
func (sf *SaleForm) AssocFeeForDisplay() string {

	if sf.AssocMemberFee == 0 {
		return ""
	}

	return fmt.Sprintf("£%.2f", sf.AssocMemberFee)
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

// ExtraDetailForm holds data that is not provided at the start of the sale -
// title, address and so on.
type ExtraDetailForm struct {
	Title                 string // optional
	FirstName             string
	FirstNameError        string
	Surname               string
	SurnameError          string
	AddressLine1          string
	AddressLine1Error     string
	AddressLine2          string   // optional
	AddressLine3          string   // optional
	Town                  string   // optional
	County                string   // optional
	Postcode              string   // optional
	CountryCode           string   // Three letter code, for example "GBR" - https://en.wikipedia.org/wiki/ISO_3166-1_alpha-3
	Country               string   // Country (from CountryCode)
	Phone                 string   // starts with '+' or '0'
	Mobile                string   // starts with '+' or '0'
	LocationOfInterest    string   // One of the local parishes
	TopicsOfInterest      []string // Topics from the database
	OtherTopicsOfInterest string   // Topics not in the database, chosen by the member.
}

// SetCountryFromCode takes a three-letter country code such as "GBR" from the
// CountryCode in the form, looks up the country name, sets it in the Country
// field and returns the value.  If the code not in the list, an empty string is
// returned.  For example country code "GBR" sets the country to "United Kingdom".
// The list is from https://en.wikipedia.org/wiki/ISO_3166-1_alpha-3.
func (edf *ExtraDetailForm) SetCountryFromCode() string {
	code := map[string]string{
		"ABW": "Aruba",
		"AFG": "Afghanistan",
		"AGO": "Angola",
		"AIA": "Anguilla",
		"ALA": "Åland Islands",
		"ALB": "Albania",
		"AND": "Andorra",
		"ARE": "United Arab Emirates",
		"ARG": "Argentina",
		"ARM": "Armenia",
		"ASM": "American Samoa",
		"ATA": "Antarctica",
		"ATF": "French Southern Territories",
		"ATG": "Antigua and Barbuda",
		"AUS": "Australia",
		"AUT": "Austria",
		"AZE": "Azerbaijan",
		"BDI": "Burundi",
		"BEL": "Belgium",
		"BEN": "Benin",
		"BES": "Bonaire, Sint Eustatius and Saba",
		"BFA": "Burkina Faso",
		"BGD": "Bangladesh",
		"BGR": "Bulgaria",
		"BHR": "Bahrain",
		"BHS": "Bahamas",
		"BIH": "Bosnia and Herzegovina",
		"BLM": "Saint Barthélemy",
		"BLR": "Belarus",
		"BLZ": "Belize",
		"BMU": "Bermuda",
		"BOL": "Bolivia, Plurinational State of",
		"BRA": "Brazil",
		"BRB": "Barbados",
		"BRN": "Brunei Darussalam",
		"BTN": "Bhutan",
		"BVT": "Bouvet Island",
		"BWA": "Botswana",
		"CAF": "Central African Republic",
		"CAN": "Canada",
		"CCK": "Cocos (Keeling) Islands",
		"CHE": "Switzerland",
		"CHL": "Chile",
		"CHN": "China",
		"CIV": "Côte d'Ivoire",
		"CMR": "Cameroon",
		"COD": "Congo, Democratic Republic of the",
		"COG": "Congo",
		"COK": "Cook Islands",
		"COL": "Colombia",
		"COM": "Comoros",
		"CPV": "Cabo Verde",
		"CRI": "Costa Rica",
		"CUB": "Cuba",
		"CUW": "Curaçao",
		"CXR": "Christmas Island",
		"CYM": "Cayman Islands",
		"CYP": "Cyprus",
		"CZE": "Czechia",
		"DEU": "Germany",
		"DJI": "Djibouti",
		"DMA": "Dominica",
		"DNK": "Denmark",
		"DOM": "Dominican Republic",
		"DZA": "Algeria",
		"ECU": "Ecuador",
		"EGY": "Egypt",
		"ERI": "Eritrea",
		"ESH": "Western Sahara",
		"ESP": "Spain",
		"EST": "Estonia",
		"ETH": "Ethiopia",
		"FIN": "Finland",
		"FJI": "Fiji",
		"FLK": "Falkland Islands (Malvinas)",
		"FRA": "France",
		"FRO": "Faroe Islands",
		"FSM": "Micronesia, Federated States of",
		"GAB": "Gabon",
		"GBR": "United Kingdom of Great Britain and Northern Ireland",
		"GEO": "Georgia",
		"GGY": "Guernsey",
		"GHA": "Ghana",
		"GIB": "Gibraltar",
		"GIN": "Guinea",
		"GLP": "Guadeloupe",
		"GMB": "Gambia",
		"GNB": "Guinea-Bissau",
		"GNQ": "Equatorial Guinea",
		"GRC": "Greece",
		"GRD": "Grenada",
		"GRL": "Greenland",
		"GTM": "Guatemala",
		"GUF": "French Guiana",
		"GUM": "Guam",
		"GUY": "Guyana",
		"HKG": "Hong Kong",
		"HMD": "Heard Island and McDonald Islands",
		"HND": "Honduras",
		"HRV": "Croatia",
		"HTI": "Haiti",
		"HUN": "Hungary",
		"IDN": "Indonesia",
		"IMN": "Isle of Man",
		"IND": "India",
		"IOT": "British Indian Ocean Territory",
		"IRL": "Ireland",
		"IRN": "Iran, Islamic Republic of",
		"IRQ": "Iraq",
		"ISL": "Iceland",
		"ISR": "Israel",
		"ITA": "Italy",
		"JAM": "Jamaica",
		"JEY": "Jersey",
		"JOR": "Jordan",
		"JPN": "Japan",
		"KAZ": "Kazakhstan",
		"KEN": "Kenya",
		"KGZ": "Kyrgyzstan",
		"KHM": "Cambodia",
		"KIR": "Kiribati",
		"KNA": "Saint Kitts and Nevis",
		"KOR": "Korea, Republic of",
		"KWT": "Kuwait",
		"LAO": "Lao People's Democratic Republic",
		"LBN": "Lebanon",
		"LBR": "Liberia",
		"LBY": "Libya",
		"LCA": "Saint Lucia",
		"LIE": "Liechtenstein",
		"LKA": "Sri Lanka",
		"LSO": "Lesotho",
		"LTU": "Lithuania",
		"LUX": "Luxembourg",
		"LVA": "Latvia",
		"MAC": "Macao",
		"MAF": "Saint Martin (French part)",
		"MAR": "Morocco",
		"MCO": "Monaco",
		"MDA": "Moldova, Republic of",
		"MDG": "Madagascar",
		"MDV": "Maldives",
		"MEX": "Mexico",
		"MHL": "Marshall Islands",
		"MKD": "North Macedonia",
		"MLI": "Mali",
		"MLT": "Malta",
		"MMR": "Myanmar",
		"MNE": "Montenegro",
		"MNG": "Mongolia",
		"MNP": "Northern Mariana Islands",
		"MOZ": "Mozambique",
		"MRT": "Mauritania",
		"MSR": "Montserrat",
		"MTQ": "Martinique",
		"MUS": "Mauritius",
		"MWI": "Malawi",
		"MYS": "Malaysia",
		"MYT": "Mayotte",
		"NAM": "Namibia",
		"NCL": "New Caledonia",
		"NER": "Niger",
		"NFK": "Norfolk Island",
		"NGA": "Nigeria",
		"NIC": "Nicaragua",
		"NIU": "Niue",
		"NLD": "Netherlands, Kingdom of the",
		"NOR": "Norway",
		"NPL": "Nepal",
		"NRU": "Nauru",
		"NZL": "New Zealand",
		"OMN": "Oman",
		"PAK": "Pakistan",
		"PAN": "Panama",
		"PCN": "Pitcairn",
		"PER": "Peru",
		"PHL": "Philippines",
		"PLW": "Palau",
		"PNG": "Papua New Guinea",
		"POL": "Poland",
		"PRI": "Puerto Rico",
		"PRK": "Korea, Democratic People's Republic of",
		"PRT": "Portugal",
		"PRY": "Paraguay",
		"PSE": "Palestine, State of",
		"PYF": "French Polynesia",
		"QAT": "Qatar",
		"REU": "Réunion",
		"ROU": "Romania",
		"RUS": "Russian Federation",
		"RWA": "Rwanda",
		"SAU": "Saudi Arabia",
		"SDN": "Sudan",
		"SEN": "Senegal",
		"SGP": "Singapore",
		"SGS": "South Georgia and the South Sandwich Islands",
		"SHN": "Saint Helena, Ascension and Tristan da Cunha",
		"SJM": "Svalbard and Jan Mayen",
		"SLB": "Solomon Islands",
		"SLE": "Sierra Leone",
		"SLV": "El Salvador",
		"SMR": "San Marino",
		"SOM": "Somalia",
		"SPM": "Saint Pierre and Miquelon",
		"SRB": "Serbia",
		"SSD": "South Sudan",
		"STP": "Sao Tome and Principe",
		"SUR": "Suriname",
		"SVK": "Slovakia",
		"SVN": "Slovenia",
		"SWE": "Sweden",
		"SWZ": "Eswatini",
		"SXM": "Sint Maarten (Dutch part)",
		"SYC": "Seychelles",
		"SYR": "Syrian Arab Republic",
		"TCA": "Turks and Caicos Islands",
		"TCD": "Chad",
		"TGO": "Togo",
		"THA": "Thailand",
		"TJK": "Tajikistan",
		"TKL": "Tokelau",
		"TKM": "Turkmenistan",
		"TLS": "Timor-Leste",
		"TON": "Tonga",
		"TTO": "Trinidad and Tobago",
		"TUN": "Tunisia",
		"TUR": "Türkiye",
		"TUV": "Tuvalu",
		"TWN": "Taiwan, Province of China",
		"TZA": "Tanzania, United Republic of",
		"UGA": "Uganda",
		"UKR": "Ukraine",
		"UMI": "United States Minor Outlying Islands",
		"URY": "Uruguay",
		"USA": "United States of America",
		"UZB": "Uzbekistan",
		"VAT": "Holy See",
		"VCT": "Saint Vincent and the Grenadines",
		"VEN": "Venezuela, Bolivarian Republic of",
		"VGB": "Virgin Islands (British)",
		"VIR": "Virgin Islands (U.S.)",
		"VNM": "Viet Nam",
		"VUT": "Vanuatu",
		"WLF": "Wallis and Futuna",
		"WSM": "Samoa",
		"YEM": "Yemen",
		"ZAF": "South Africa",
		"ZMB": "Zambia",
		"ZWE": "Zimbabwe",
	}

	if len(edf.CountryCode) == 0 {
		return ""
	}

	edf.Country = code[edf.CountryCode]

	return edf.Country
}
