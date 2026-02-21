package handler

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"regexp"
	"strings"
	"text/template"
	"time"

	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/checkout/session"

	ps "github.com/goblimey/portablesyscall"

	"github.com/goblimey/go-stripe-payments/code/pkg/config"
	"github.com/goblimey/go-stripe-payments/code/pkg/database"
	"github.com/goblimey/go-stripe-payments/code/pkg/forms"
)

// protocol contains the protocol value for the HTTP requests.  The default is
// https.  The value will be changed to "http" if the serv is running under Windows.
var protocol = "https"

func init() {
	if ps.OSName == "windows" {
		protocol = "http"
	}
}

type Handler struct {
	Conf                   *config.Config     // The incoming config.
	DBConfig               *database.DBConfig // The database config
	DB                     *database.Database // The database connection.
	OrdinaryMembershipFee  float64            // The fee for ordinary membership.
	AssociateMembershipFee float64            // The fee for associate membership (0 if not enabled).
	FriendMembershipFee    float64            // The fee for friend's membership (0 if not enabled).
	PrePaymentErrorHTML    string             // The default error message page before the customer pays.
	PostPaymentErrorHTML   string             // The default error message page after the customer has paid.
	SuccessPageHTML        string             // The page displayed on a successful sale.
	PhoneRegexp            *regexp.Regexp     // The regular expression to valdate a phone number.
	TZ                     *time.Location     // The timezone for this server.
	Logger                 *slog.Logger       // The daily logger.
}

func New(conf *config.Config) *Handler {

	// If something goes badly wrong before we take money then we need to refer
	// the user to somebody who can advise them.
	prePaymentErrorHTML := fmt.Sprintf(prePaymentErrorHTMLPattern, conf.EmailAddressForQuestions, conf.EmailAddressForQuestions)
	// If things go wrong after the customer has paid they should be referred to somebody
	// who can refund their payment, for example the treasurer.
	postPaymentErrorHTML := fmt.Sprintf(postPaymentErrorHTMLPattern, conf.EmailAddressForFailures, conf.EmailAddressForFailures)

	dbConfig := database.DBConfig{
		Type: conf.DBType,
		Host: conf.DBHostname,
		Port: conf.DBPort,
		Name: conf.DBDatabase,
		User: conf.DBUser,
		Pass: conf.DBPassword,
	}

	h := Handler{
		Conf:                   conf,
		DBConfig:               &dbConfig,
		OrdinaryMembershipFee:  conf.OrdinaryMemberFee,
		AssociateMembershipFee: conf.AssocMemberFee,
		FriendMembershipFee:    conf.FriendFee,
		PrePaymentErrorHTML:    prePaymentErrorHTML,
		PostPaymentErrorHTML:   postPaymentErrorHTML,
	}

	// Now that have a logger we can do some setup that, if it fails, forces us
	// to kill the server.  Failure is likely caused by some sort of issue that
	// must be fixed manually, for example via a bug fix and recompile.

	// Regular expression to validate a phone number.  A phone number should start
	// with "+" or "0".  That should be followed by a list of (digit or space).  For
	// example "+44 1234 567890", "01234 567890" or "020 7123 4567".
	const phonePattern = `[+0][0-9 ]+`
	var reError error
	h.PhoneRegexp, reError = regexp.Compile(phonePattern)
	if reError != nil {
		h.logError("%v", reError)
		os.Exit(-1)
	}

	// Set the server's timezone.
	var locationError error
	h.TZ, locationError = time.LoadLocation("Europe/London")
	if locationError != nil {
		h.logError("%v", locationError)
		os.Exit(-1)
	}

	return &h
}

// Home handles the "/" request.  This provides a "proof of life".
func (h *Handler) Home(w http.ResponseWriter, r *http.Request) {
	h.Logger.Info("home")
	w.Write([]byte("home page\n"))
}

// GetPaymentData handles the /subscribe/ request (and, for bacward compatibility,
// the /displayPaymentForm request.)  It validates the incoming payment data form.
// If the data is valid it displays the cost breakdown, otherwise it displays the
// payment data form again with error messages.
func (h *Handler) GetPaymentData(w http.ResponseWriter, r *http.Request) {

	h.Logger.Info("GetPaymentData")

	paymentYear := database.GetMembershipYear(time.Now().In(h.TZ))
	h.DB = database.New(h.DBConfig)
	h.DB.Logger = h.Logger
	connectionError := h.DB.Connect()
	if connectionError != nil {
		fmt.Println(connectionError.Error())
		form := forms.NewSaleForm(h.Conf, paymentYear)
		form.GeneralErrorMessage = fmt.Sprintf("Fatal error - %v", connectionError)
		form.Valid = false
		h.reportError(w, h.PrePaymentErrorHTML, connectionError)
		return
	}

	// Start a transaction, stored in the database object.
	txError := h.DB.BeginTx()

	if txError != nil {
		h.reportError(w, h.PrePaymentErrorHTML, txError)
		return
	}

	// The transaction should be ether committed or rolled back before this function exits.
	// In case it isn't, set up a deferred rollback now.  We choose a rollback rather than
	// a commit because failure to close the transaction already might be caused by some
	// sort of catastrophic error.
	//
	// When the function is returning, the transaction may have already been closed, in which
	// case the second rollback may return an error, but that will be ignored.
	defer h.DB.Rollback()

	defer h.DB.Close()

	// The helper does the work.
	h.paymentDataHelper(w, r, paymentYear)

	// paymentDataHelper doesn't change the database so we can just
	// close the transaction via a rollback.
	h.DB.Rollback()
}

// GetPaymentDataHelper validates the form and prepares the response.
// It's separated out to support unit testing.
func (h *Handler) paymentDataHelper(w http.ResponseWriter, r *http.Request, paymentYear int) {

	h.Logger.Info("paymentDataHelper")
	sf := forms.NewSaleForm(h.Conf, paymentYear)

	sf.OrdinaryMemberFee = h.Conf.OrdinaryMemberFee
	sf.AssocMemberFee = h.Conf.AssocMemberFee
	sf.FriendFee = h.Conf.FriendFee

	sf.Title = r.PostFormValue("title")
	sf.FirstName = r.PostFormValue("first_name")
	sf.LastName = r.PostFormValue("last_name")
	sf.Email = r.PostFormValue("email")
	sf.FriendInput = r.PostFormValue("friend")
	sf.DonationToSocietyInput = r.PostFormValue("donation_to_society")
	sf.DonationToMuseumInput = r.PostFormValue("donation_to_museum")
	sf.GiftaidInput = r.PostFormValue("giftaid")

	sf.AssocTitle = r.PostFormValue("assoc_title")
	sf.AssocFirstName = r.PostFormValue("assoc_first_name")
	sf.AssocLastName = r.PostFormValue("assoc_last_name")
	sf.AssocEmail = r.PostFormValue("assoc_email")
	sf.AssocFriendInput = r.PostFormValue("assoc_friend")

	if len(sf.Title) == 0 &&
		len(sf.FirstName) == 00 &&
		len(sf.LastName) == 0 &&
		len(sf.Email) == 0 &&
		len(sf.DonationToSocietyInput) == 0 &&
		len(sf.DonationToMuseumInput) == 0 &&
		len(sf.AssocFirstName) == 0 &&
		len(sf.AssocLastName) == 0 {

		// On the first call in a sequence, display an empty form with mandatory fields marked.
		h.displayInitialSaleForm(w, paymentYear)
		return
	}

	// Validate the form data.  On the first call the form is empty.
	// The validator sets error messages containing asterisks against
	// the mandatory fields.  On calls with incoming data, it validates
	// that data and sets error messages.

	valid := ValidateSaleForm(sf)

	if !valid {

		// There are errors, display the form again
		// with any supplied fields filled in.
		h.displaySaleForm(w, sf)

		return
	}

	// Build and display the payment confirmation page.

	h.setPayments(sf)

	// Check the template.
	paymentConfirmationPageTemplate, templateError :=
		template.New("PaymentConfirmationPage").Parse(paymentConfirmationPageTemplate)
	if templateError != nil {
		h.logError("paymentDataHelper: %v", templateError)
		w.Write([]byte(h.PrePaymentErrorHTML))
		return
	}

	// Write the response.
	executeError := paymentConfirmationPageTemplate.Execute(w, sf)

	if executeError != nil {
		h.Logger.Info(executeError.Error())
		w.Write([]byte(h.PrePaymentErrorHTML))
		return
	}
}

// setPayments takes the data supplied by the user and sets the payment
// fields in the sale form object - ordinary membership fee etc.
func (h *Handler) setPayments(ms *forms.SaleForm) {

	ms.OrdinaryMemberFee = h.OrdinaryMembershipFee

	if ms.EnableOtherMemberTypes {
		if ms.Friend {
			// The ordinary member is a friend so must pay the friend fee.
			ms.FriendFeeToPay = h.FriendMembershipFee
		}
		if len(ms.AssocFirstName) > 0 {
			// There is an associate member - another fee.
			ms.AssocFeeToPay = h.AssociateMembershipFee

			if ms.AssocFriend {
				// The associate member is a friend, so must pay the friend fee.
				ms.AssocFriendFeeToPay = h.FriendMembershipFee
			}
		}
	}

	h.logMessage("%s %s member %f friend %f assoc member %f assoc friend %f", ms.FirstName, ms.LastName, ms.OrdinaryMemberFee,
		ms.FriendFeeToPay, ms.AssocFeeToPay, ms.AssocFriendFeeToPay)
}

// Checkout is the handler for the /checkout request.  It validates the
// HTTP parameters and, if valid, creates a MembershipSale record and
// redirects to the Stripe payment website.
func (h *Handler) Checkout(w http.ResponseWriter, r *http.Request) {

	h.Logger.Info("Checkout")

	h.DB = database.New(h.DBConfig)
	h.DB.Logger = h.Logger
	connectionError := h.DB.Connect()
	if connectionError != nil {
		h.reportError(w, h.PrePaymentErrorHTML, connectionError)
		return
	}

	// Start a transaction, stored in the database object.
	txError := h.DB.BeginTx()

	if txError != nil {
		h.reportError(w, h.PrePaymentErrorHTML, txError)
		return
	}

	// The transaction should be ether committed or rolled back before this function exits.
	// In case it isn't, set up a deferred rollback now.  We choose a rollback rather than
	// a commit because failure to close the transaction already is almost certainly caused
	// by some sort of catastrophic error.
	//
	// When the function is returning, the transaction may have already been closed, in which
	// case the second rollback may return an error, but that will be ignored.
	defer h.DB.Rollback()
	defer h.DB.Close()

	paymentYear := database.GetMembershipYear(time.Now().In(h.TZ))

	h.checkoutHelper(w, r, paymentYear)
}

// checkoutHelper creates a MembershipSale record to record progress and prepares the response.
// On success that's a redirect to the Stripe payment system.  The helper is separated out to
// support unit testing.
func (h *Handler) checkoutHelper(w http.ResponseWriter, r *http.Request, paymentYear int) {

	const fn = "checkoutHelper"
	h.Logger.Info(fn)

	// If the incoming data is the result of the expected page flow, everything has
	// been validated but we can't assume that.  Somebody may be trying to pull a
	// fast one, so we validate the form again.  If there is any error, we stop
	// processing.

	sf := forms.NewSaleForm(h.Conf, paymentYear)
	sf.Title = r.PostFormValue("title")
	sf.FirstName = r.PostFormValue("first_name")
	sf.LastName = r.PostFormValue("last_name")
	sf.Email = r.PostFormValue("email")
	sf.FriendInput = r.PostFormValue("friend")
	sf.DonationToSocietyInput = r.PostFormValue("donation_to_society")
	sf.DonationToMuseumInput = r.PostFormValue("donation_to_museum")
	sf.GiftaidInput = r.PostFormValue("giftaid")

	sf.AssocTitle = r.PostFormValue("assoc_title")
	sf.AssocFirstName = r.PostFormValue("assoc_first_name")
	sf.AssocLastName = r.PostFormValue("assoc_last_name")
	sf.AssocEmail = r.PostFormValue("assoc_email")
	sf.AssocFriendInput = r.PostFormValue("assoc_friend")

	if !ValidateSaleForm(sf) {
		// The data should already have been validated so this should never happen.
		h.logError("%s: invalid data", fn)
		h.reportError(w, h.PostPaymentErrorHTML, errors.New("internal error"))
	}

	// The incoming data is valid.  Create and commit the membership_sales record
	// (status pending).
	ms := database.NewMembershipSale(h.Conf)
	ms.MembershipYear = paymentYear
	ms.PaymentService = "Stripe"
	ms.PaymentStatus = "Pending"
	ms.Title = sf.Title
	ms.FirstName = sf.FirstName
	ms.LastName = sf.LastName
	ms.Email = sf.Email
	ms.Friend = sf.Friend
	ms.DonationToSociety = sf.DonationToSociety
	ms.DonationToMuseum = sf.DonationToMuseum
	ms.Giftaid = sf.Giftaid
	ms.AssocTitle = sf.AssocTitle
	ms.AssocFirstName = sf.AssocFirstName
	ms.AssocLastName = sf.AssocLastName
	ms.AssocEmail = sf.AssocEmail
	ms.AssocFriend = sf.AssocFriend
	ms.PaymentStatus = database.PaymentStatusPending
	ms.OrdinaryMemberFeePaid = h.OrdinaryMembershipFee

	h.logMessage("%s: %s %s %s, %s %s %s",
		fn, ms.Title, ms.FirstName, ms.LastName,
		ms.AssocTitle, ms.AssocFirstName, ms.AssocLastName)

	if ms.EnableOtherMemberTypes {
		if ms.Friend {
			// The ordinary member is a friend so must pay the friend fee.
			ms.FriendFeePaid = h.FriendMembershipFee
		}
		if len(ms.AssocFirstName) > 0 {

			// There is an associate member - another fee.
			ms.AssocFeePaid = h.AssociateMembershipFee

			if ms.AssocFriend {
				// The associate member is a friend, so must pay the friend fee.
				ms.AssocFriendFeePaid = h.FriendMembershipFee
			}
		}
	}

	salesID, createError := ms.Create(h.DB)
	if createError != nil {
		h.DB.Rollback()
		h.logError("%s: CreateError - %v", fn, createError)
		h.reportError(w, h.PrePaymentErrorHTML, createError)
	}

	// We have all we need from the database - commit the transaction.
	h.DB.Commit()

	// Prepare to pass control to the Stripe payment page.

	successURL := fmt.Sprintf("%s://%s/success?session_id={CHECKOUT_SESSION_ID}", protocol, r.Host)
	cancelURL := fmt.Sprintf("%s://%s/cancel", protocol, r.Host)

	invoicingEnabled := true

	description := fmt.Sprintf(
		"%s membership year %d", h.Conf.OrganisationName, paymentYear)

	invoiceData := stripe.CheckoutSessionInvoiceCreationInvoiceDataParams{
		Description: &description,
	}

	invoiceCreation := stripe.CheckoutSessionInvoiceCreationParams{
		Enabled:     &invoicingEnabled,
		InvoiceData: &invoiceData,
	}

	// Create a Stripe checkout session containing a client reference ID that gives
	// the ms_id of the membership_sales record.  This allows the application to
	// pick up where it left off when control is returned from Stripe.

	salesIDStr := fmt.Sprintf("%d", salesID)

	priceInPennies := int64(ms.Total()*100 + 0.5)

	params := &stripe.CheckoutSessionParams{
		Mode:            stripe.String(string(stripe.CheckoutSessionModePayment)),
		InvoiceCreation: &invoiceCreation,
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
					Currency: stripe.String("gbp"),
					ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
						Name: stripe.String("Service"),
					},
					UnitAmount: stripe.Int64(priceInPennies),
				},
				Quantity: stripe.Int64(1),
			},
		},
		// This ID will be returned in the session.
		ClientReferenceID: &salesIDStr,
		// Stripe will request this URL if the payment is successful. The
		// {CHECKOUT_SESSION_ID} placeholder will be replaced by the session ID, which
		// allows the handler to retrieve the session.
		SuccessURL: stripe.String(successURL),
		// Stripe will request this URL if the payment if cancelled.
		CancelURL: stripe.String(cancelURL),
	}

	// Create the checkout session.
	s, sessErr := session.New(params)
	if sessErr != nil {

		h.DB.Rollback()
		h.logMessage("error creating Stripe session - %v", sessErr)
		h.reportError(w, h.PrePaymentErrorHTML, sessErr)
	}

	// Redirect to the Stripe system.  On a successful payment, it will
	// redirect to /success and we will continue.
	http.Redirect(w, r, s.URL, http.StatusSeeOther)
}

// CreateCheckoutSession is the handler for the /create-checkout-session
// request.  It prepares the Stripe session and redirects the browser to
// the Stripe payment page.
func (h *Handler) CreateCheckoutSession(w http.ResponseWriter, r *http.Request) {
	successURL :=
		fmt.Sprintf("%s://%s/success?session_id={CHECKOUT_SESSION_ID}",
			protocol, r.Host)
	cancelURL := fmt.Sprintf("%s://%s/cancel", protocol, r.Host)

	invoiceEnabled := true
	year := database.GetMembershipYear(time.Now().In(h.TZ))
	description := fmt.Sprintf("%s membership system %d", h.Conf.OrganisationName, year)
	invoiceData := stripe.CheckoutSessionInvoiceCreationInvoiceDataParams{
		Description: &description,
	}
	// Create a checkout session containing the client reference ID.
	params := &stripe.CheckoutSessionParams{
		Mode: stripe.String(string(stripe.CheckoutSessionModePayment)),
		InvoiceCreation: &stripe.CheckoutSessionInvoiceCreationParams{
			Enabled:     &invoiceEnabled,
			InvoiceData: &invoiceData,
		},
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
					Currency: stripe.String("gbp"),
					ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
						Name: stripe.String("Service"),
					},
					UnitAmount: stripe.Int64(3000),
				},
				Quantity: stripe.Int64(1),
			},
		},
		// This ID will be returned in the session.
		// ClientReferenceID: &userID,
		// Stripe will request this URL if the payment is
		// successful. The {CHECKOUT_SESSION_ID} placeholder will
		// be replaced by the session ID, which allows the handler
		// to retrieve the session.
		SuccessURL: stripe.String(successURL),
		// Stripe will request this URL if the payment if cancelled.
		CancelURL: stripe.String(cancelURL),
	}

	// Create the checkout session.
	s, err := session.New(params)
	if err != nil {
		h.logError("/create-checkout-session: error creating stripe session: %v", err)
	}
	http.Redirect(w, r, s.URL, http.StatusSeeOther)
}

// Success is the handler for the /success request.  On a successful
// payment, the Stripe system issues that request, filling in the
// {CHECKOUT_SESSION_ID} placeholder with the session ID.  The
// handler uses that to retrieve the checkout session, extract the
// client reference and complete the sale.
func (h *Handler) Success(w http.ResponseWriter, r *http.Request) {

	h.logMessage("Success()")

	// We figure out the start and end dates here to support unit testing of the SuccessHelper.
	startTime := time.Now().In(h.TZ)
	// The end date is the end of the calendar year of payment.
	yearEnd := time.Date(
		startTime.Year(), time.December, 31, 23, 59, 59, 999999999, h.TZ,
	)

	h.DB = database.New(h.DBConfig)
	h.DB.Logger = h.Logger
	connectionError := h.DB.Connect()
	if connectionError != nil {
		h.reportError(w, h.PrePaymentErrorHTML, connectionError)
		return
	}

	// Start a transaction.
	txError := h.DB.BeginTx()
	if txError != nil {
		h.reportError(w, h.PrePaymentErrorHTML, txError)
		return
	}

	// The transaction should be ether committed or rolled back before this function exits.
	// In case it isn't, set up a deferred rollback now.  We choose a rollback rather than
	// a commit because failure to close the transaction already is almost certainly caused
	// by some sort of catastrophic error.
	//
	// When the function is returning, the transaction may have already been closed, in which
	// case the second rollback may return an error, but that will be ignored.
	defer h.DB.Rollback()
	defer h.DB.Close()

	// Get the Stripe session.
	sessionID := r.URL.Query().Get("session_id")
	params := stripe.CheckoutSessionParams{}
	stripeSession, sessionGetError := session.Get(sessionID, &params)
	if sessionGetError != nil {
		h.reportError(w, h.PrePaymentErrorHTML, sessionGetError)
		return
	}

	connError := h.connectToDB()
	if connError != nil {
		h.reportError(w, h.PrePaymentErrorHTML, connError)
		return
	}

	now := time.Now().In(h.TZ)
	paymentYear := database.GetMembershipYear(now)

	h.successHelper(w, stripeSession, startTime, yearEnd, now, paymentYear)

	// The helper should send an HTTP response so we shouldn't get to here.
}

// successHelper completes the sale.  It's separated out and the start and end dates are supplied to
// support unit testing.
func (h *Handler) successHelper(w http.ResponseWriter, stripeSession *stripe.CheckoutSession, startDate, endDate, now time.Time, paymentYear int) {

	const fn = "successHelper"

	ms, msError := h.getMembershipSaleOnSuccess(stripeSession, startDate, endDate, now, paymentYear)
	if msError != nil {
		h.reportError(w, h.PostPaymentErrorHTML, msError)
		h.DB.Rollback()
		return
	}

	h.logMessage("%s: payment successful -%s for %s %s %s, %s %s %s",
		fn, ms.TransactionType, ms.Title, ms.FirstName, ms.LastName,
		ms.AssocTitle, ms.AssocFirstName, ms.AssocLastName)

	cmError := h.setMemberDetails(ms, starvitDate, endDate, now, paymentYear)
	if cmError != nil {
		h.reportError(w, h.PostPaymentErrorHTML, cmError)
		h.DB.Rollback()
		return
	}

	// // The user must have paid at least this fee.
	// ms.OrdinaryMemberFeePaid = h.Conf.OrdinaryMemberFee

	// // The sale is complete.
	// ms.PaymentStatus = database.PaymentStatusComplete

	// ms.PaymentID = fmt.Sprintf("%s %s", stripeSession.Customer.ID, stripeSession.Customer.Email)

	// // Membership renewal or new member(s)?
	// var lookupError error
	// ms.UserID, ms.AssocUserID, lookupError = usersExist(ms, h.DB)
	// if lookupError != nil {
	// 	h.DB.Rollback()
	// 	h.reportError(w, h.PrePaymentErrorHTML, lookupError)
	// 	return
	// }

	// var newUser bool
	// if ms.UserID <= 0 {
	// 	newUser = true
	// }

	// if newUser {
	// 	// A new member is registering.  We haven't created the member account(s) yet.
	// 	newUser = true
	// 	ms.TransactionType = database.TransactionTypeNewMember
	// 	var createUserError error
	// 	ms.UserID, ms.AssocUserID, createUserError =
	// 		h.DB.CreateAccounts(ms, startDate, endDate)
	// 	if createUserError != nil {
	// 		// Failed to create one or both of the users.  The user has paid but we
	// 		// can't fulfill the sale so this error is bad.  Report it to the user.
	// 		h.DB.Rollback()
	// 		h.reportError(w, h.PostPaymentErrorHTML, createUserError)
	// 		return
	// 	}
	// } else {
	// 	// An existing member is renewing, possibly with an associate member.
	// 	ms.TransactionType = database.TransactionTypeRenewal

	// 	// For a new member, various fields are set at this point, so set them
	// 	// for the renewing member(s) too.  The most important change is setting
	// 	// the member end date, because that's what marks them as a paid up
	// 	// member, which is what they've just paid for.

	// 	// Set the end date for the ordinary member.
	// 	omError := h.DB.SetMemberEndDate(ms.UserID, ms.MembershipYear)
	// 	if omError != nil {
	// 		h.DB.Rollback()
	// 		h.reportError(w, h.PostPaymentErrorHTML, omError)
	// 		return
	// 	}

	// 	// Set the name fields.
	// 	ttError := h.DB.SetTitle(ms.UserID, ms.Title)
	// 	if ttError != nil {
	// 		h.DB.Rollback()
	// 		h.reportError(w, h.PostPaymentErrorHTML, ttError)
	// 		return
	// 	}

	// 	fnError := h.DB.SetFirstName(ms.UserID, ms.FirstName)
	// 	if fnError != nil {
	// 		h.DB.Rollback()
	// 		h.reportError(w, h.PostPaymentErrorHTML, fnError)
	// 		return
	// 	}

	// 	lnError := h.DB.SetLastName(ms.UserID, ms.LastName)
	// 	if lnError != nil {
	// 		h.DB.Rollback()
	// 		h.reportError(w, h.PostPaymentErrorHTML, lnError)
	// 		return
	// 	}

	// 	if h.Conf.EnableOtherMemberTypes && ms.AssocUserID > 0 {
	// 		// Set the end date for the associate member.
	// 		assocError := h.DB.SetMemberEndDate(ms.AssocUserID, ms.MembershipYear)
	// 		if assocError != nil {
	// 			h.DB.Rollback()
	// 			h.reportError(w, h.PostPaymentErrorHTML, assocError)
	// 			return
	// 		}

	// 		// Set the associate member's name fields.
	// 		assocTTError := h.DB.SetTitle(ms.AssocUserID, ms.AssocTitle)
	// 		if assocTTError != nil {
	// 			h.DB.Rollback()
	// 			h.reportError(w, h.PostPaymentErrorHTML, assocTTError)
	// 			return
	// 		}

	// 		assocFNError := h.DB.SetFirstName(ms.AssocUserID, ms.AssocFirstName)
	// 		if assocFNError != nil {
	// 			h.DB.Rollback()
	// 			h.reportError(w, h.PostPaymentErrorHTML, assocFNError)
	// 			return
	// 		}

	// 		assocLNError := h.DB.SetLastName(ms.AssocUserID, ms.AssocLastName)
	// 		if assocLNError != nil {
	// 			h.DB.Rollback()
	// 			h.reportError(w, h.PostPaymentErrorHTML, assocLNError)
	// 			return
	// 		}
	// 	}
	// }

	// // Set the data protection field for the full-price member.
	// dpError := h.DB.SetDataProtectionField(ms.UserID, true)
	// if dpError != nil {
	// 	h.DB.Rollback()
	// 	h.reportError(w, h.PostPaymentErrorHTML, dpError)
	// 	return
	// }

	// if ms.AssocUserID > 0 {
	// 	assocDPError := h.DB.SetDataProtectionField(ms.AssocUserID, true)
	// 	if assocDPError != nil {
	// 		h.DB.Rollback()
	// 		h.reportError(w, h.PostPaymentErrorHTML, assocDPError)
	// 		return
	// 	}
	// }

	// // Set the data protection field for the full-price member.
	// reError := h.DB.SetReceiveEmailField(ms.UserID, true)
	// if dpError != nil {
	// 	h.DB.Rollback()
	// 	h.reportError(w, h.PostPaymentErrorHTML, reError)
	// 	return
	// }

	// if ms.AssocUserID > 0 {
	// 	assocREError := h.DB.SetReceiveEmailField(ms.AssocUserID, true)
	// 	if assocREError != nil {
	// 		h.DB.Rollback()
	// 		h.reportError(w, h.PostPaymentErrorHTML, assocREError)
	// 		return
	// 	}
	// }

	// We've done the important update.  In case something catastrophic happens later,
	// commit the changes made so far and then open a new transaction.

	updateError1 := ms.Update(h.DB)
	if updateError1 != nil {
		h.logError("%s: user ID %d - failed to update membership sales record %d - %v",
			fn, ms.UserID, ms.ID, updateError1)
	}

	commit1Error := h.DB.Commit()
	if commit1Error != nil {
		h.logError("%s: user ID %d - failed to commit updated membership sales record %d - %v",
			fn, ms.UserID, ms.ID, updateError1)
		h.DB.Rollback()
		h.reportError(w, h.PostPaymentErrorHTML, commit1Error)
		return
	}

	txError := h.DB.BeginTx()
	if txError != nil {
		h.reportError(w, h.PrePaymentErrorHTML, txError)
		return
	}

	h.setAccountingRecordsForMembers(ms, now)

	// Commit the accounting records.
	commit2Error := h.DB.Commit()
	if commit2Error != nil {
		h.logError("%s: user ID %d - %v", fn, ms.UserID, commit2Error)
	}

	txe := h.DB.BeginTx()
	if txe != nil {
		h.logError("%s: %v", fn, txe)
	}
	// There are no more DB writes from now on, so there will be nothing to commit.
	defer h.DB.Rollback()

	user, fue := h.DB.GetUser(ms.UserID)
	if fue != nil {
		h.logError("%s: %v", fn, fue)
	}
	ms.AccountName = user.LoginName

	if ms.AssocUserID > 0 {
		// There is an associate .
		assocUser, faue := h.DB.GetUser(ms.AssocUserID)
		if faue != nil {
			h.logError("%s: %v", fn, faue)
		}
		ms.AssocAccountName = assocUser.LoginName
	}

	if ms.TransactionType == database.TransactionTypeRenewal {
		// A user is renewing.  Get any extra details that they have already set.
		// (for example, in a previous year.  These are used to pre-populate the
		// extra details collection page.
		h.fetchCurrentExtraDetails(ms)
	}

	// Create the selection list of countries.
	var countriesHTML string
	switch {
	case ms.TransactionType == database.TransactionTypeNewMember:
		// For a new member the default country is the UK.
		var ce error
		countriesHTML, ce = h.MakeCountrySelectionListPreSelecting("GBR")
		if ce != nil {
			h.Logger.Error(ce.Error())
			w.Write([]byte(h.PrePaymentErrorHTML))
			return
		}

	case len(ms.CountryCode) > 0:
		// This is an existing member and they specified a country last time.  Put
		// that country at the top of the selection list.
		var ce error
		countriesHTML, ce = h.MakeCountrySelectionListPreSelecting(ms.CountryCode)
		if ce != nil {
			h.Logger.Error(ce.Error())
			w.Write([]byte(h.PrePaymentErrorHTML))
			return
		}

	default:
		// This is an existing member but they didn't specify a country last time.
		// Put the UK at the top of the selection list and pre-select it.
		var ce error
		countriesHTML, ce = h.MakeCountrySelectionListPreSelecting("GBR")
		if ce != nil {
			h.Logger.Error(ce.Error())
			w.Write([]byte(h.PrePaymentErrorHTML))
			return
		}
	}

	// Create the selection list of interests.
	interestHTML := h.makeInterestSelectionHTML(ms)

	// Create the completion page.
	successPageTemplateString := successPageTemplateString1 + countriesHTML +
		successPageTemplateString2 + interestHTML + successPageTemplateString3

	successPageTemplate, parseError :=
		template.New("SuccessPage").Parse(successPageTemplateString)
	if parseError != nil {
		h.Logger.Error(parseError.Error())
		w.Write([]byte(h.PrePaymentErrorHTML))
		return
	}

	// Write the response.
	executeError := successPageTemplate.Execute(w, ms)
	if executeError != nil {
		h.Logger.Error(executeError.Error())
		w.Write([]byte(h.PrePaymentErrorHTML))
		return
	}

	// Success!
}

func (h *Handler) getMembershipSaleOnSuccess(stripeSession *stripe.CheckoutSession, startDate, endDate, now time.Time, paymentYear int) (*database.MembershipSale, error) {
	const fn = "getMembershipSaleOnSuccess"

	if stripeSession.PaymentStatus != "paid" {
		e := fmt.Errorf("%s: payment status in stripe session should be paid - %s", fn, stripeSession.PaymentStatus)
		return nil, e
	}

	var saleID int64
	_, saleIDError := fmt.Sscanf(stripeSession.ClientReferenceID, "%d", &saleID)
	if saleIDError != nil {
		e := fmt.Errorf("%s: error converting sales ID %s - %v", fn, stripeSession.ClientReferenceID, saleIDError)
		return nil, e
	}

	// Get the membership sales record.  The ClientReferenceID in the payment
	// session is the ID of the sales record.
	ms, fetchError := h.DB.GetMembershipSale(saleID)
	if fetchError != nil {
		return nil, fetchError
	}

	ms.MembershipYear = paymentYear

	// Add the reference data. (It's used by the HTML pages)
	ms.OrganisationName = h.Conf.OrganisationName
	ms.MembershipYear = paymentYear
	ms.EnableOtherMemberTypes = h.Conf.EnableOtherMemberTypes
	ms.EnableGiftaid = h.Conf.EnableGiftaid
	ms.EmailAddressForQuestions = h.Conf.EmailAddressForQuestions
	ms.EmailAddressForFailures = h.Conf.EmailAddressForFailures

	return ms, nil
}

// setMemberDetails creates the members if necessary and sets the adm_user_data
// fields (title, first name, last name etc).
func (h *Handler) setMemberDetails(ms *database.MembershipSale, startDate, endDate, now time.Time, paymentYear int) error {

	// The sale is complete.
	ms.PaymentStatus = database.PaymentStatusComplete
	ms.OrdinaryMemberFeePaid = h.Conf.OrdinaryMemberFee

	// Check if the users already exist.
	var lookupError error
	ms.UserID, ms.AssocUserID, lookupError = usersExist(ms, h.DB)
	if lookupError != nil {
		return lookupError
	}

	if ms.UserID <= 0 {
		// This is a sale of new membership.
		ms.TransactionType = database.TransactionTypeNewMember
		// Create the member account(s).
		var createUserError error
		ms.UserID, ms.AssocUserID, createUserError =
			h.DB.CreateAccounts(ms, startDate, endDate)
		if createUserError != nil {
			// Failed to create one or both of the users.
			return createUserError
		}
	}

	// Set the end date for the ordinary member.
	omError := h.DB.SetMemberEndDate(ms.UserID, ms.MembershipYear)
	if omError != nil {
		return omError
	}

	// Set the data fields (adm_user_data table) for the full-price member.

	if len(ms.Title) > 0 {
		ttError := h.DB.SetTitle(ms.UserID, ms.Title)
		if ttError != nil {
			return ttError
		}
	}

	if len(ms.FirstName) > 0 {
		fnError := h.DB.SetFirstName(ms.UserID, ms.FirstName)
		if fnError != nil {
			return fnError
		}
	}

	if len(ms.LastName) > 0 {
		lnError := h.DB.SetLastName(ms.UserID, ms.LastName)
		if lnError != nil {
			return lnError
		}
	}

	if len(ms.Email) > 0 {
		emError := h.DB.SetEmail(ms.UserID, ms.Email)
		if emError != nil {
			return emError
		}
	}

	if ms.DonationToSociety > 0 {
		dtsError := h.DB.SetDonationToSociety(ms.UserID, ms.DonationToSociety)
		if dtsError != nil {
			return dtsError
		}
	}

	if ms.DonationToMuseum > 0 {
		dtmError := h.DB.SetDonationToMuseum(ms.UserID, ms.DonationToMuseum)
		if dtmError != nil {
			return dtmError
		}
	}

	// Set the associates's friend field (true or false).
	fError := h.DB.SetFriendField(ms.UserID, ms.Friend)
	if fError != nil {
		return fError
	}
	if ms.Giftaid {
		gError := h.DB.SetGiftaid(ms.UserID, true)
		if gError != nil {
			return gError
		}
	}

	// Set the data protection field for the full-price member.
	dpError := h.DB.SetDataProtectionField(ms.UserID, true)
	if dpError != nil {
		return dpError
	}

	// Set the receive email field for the full-price member.
	reError := h.DB.SetReceiveEmailField(ms.UserID, true)
	if reError != nil {
		return reError
	}

	if h.Conf.EnableOtherMemberTypes && ms.AssocUserID > 0 {
		// Set the end date for the associate member.
		assocError := h.DB.SetMemberEndDate(ms.AssocUserID, ms.MembershipYear)
		if assocError != nil {
			return assocError
		}

		// Set the associate member's name fields.
		if len(ms.AssocTitle) > 0 {
			assocTTError := h.DB.SetTitle(ms.AssocUserID, ms.AssocTitle)
			if assocTTError != nil {
				return assocTTError
			}
		}

		if len(ms.AssocFirstName) > 0 {
			assocFNError := h.DB.SetFirstName(ms.AssocUserID, ms.AssocFirstName)
			if assocFNError != nil {
				return assocFNError
			}
		}

		if len(ms.AssocLastName) > 0 {
			assocLNError := h.DB.SetLastName(ms.AssocUserID, ms.AssocLastName)
			if assocLNError != nil {
				return assocLNError
			}
		}

		if len(ms.AssocEmail) > 0 {
			aemError := h.DB.SetEmail(ms.AssocUserID, ms.AssocEmail)
			if aemError != nil {
				return aemError
			}
		}

		// Set the associates's friend field (true or false).
		afError := h.DB.SetFriendField(ms.AssocUserID, ms.AssocFriend)
		if afError != nil {
			return afError
		}

		// Set the data protection field for the associate member.
		assocDPError := h.DB.SetDataProtectionField(ms.AssocUserID, true)
		if assocDPError != nil {
			return assocDPError
		}

		// Set the receive email field for the full-price member.
		assocREError := h.DB.SetReceiveEmailField(ms.AssocUserID, true)
		if assocREError != nil {
			return assocREError
		}
	}

	//Success!
	return nil
}

// setAccountingRecordsForMembers stores some details of the members that are used
// for our accounting.
func (h *Handler) setAccountingRecordsForMembers(ms *database.MembershipSale, paymentDate time.Time) {
	// The members are not very interested in these records.  If we get an error, just log it
	// and continue processing.

	fn := "setAccountingRecordsForMembers"
	dlpError := h.DB.SetDateLastPaid(ms.UserID, paymentDate)
	if dlpError != nil {
		h.logError("%s: user ID %d - %v\n", fn, ms.UserID, dlpError)
	}

	// The ID of the user record of the ordinary member is in the sale record.  If
	// the sale includes an associate member, the ID of their user record is there too.

	// Count members and friends at this address.  Those values will be written later
	// to the ordinary and (if present) the associate.
	membersAtAddress := 1
	var friendsAtAddress int

	if ms.Friend {
		friendsAtAddress++
	}

	if h.Conf.EnableOtherMemberTypes && ms.AssocUserID > 0 {
		membersAtAddress++
		if ms.AssocFriend {
			friendsAtAddress++
		}
	}

	paymentError := h.DB.SetLastPayment(ms.UserID, ms.Total())
	if paymentError != nil {
		em := fmt.Sprintf("error setting last payment for %d - %v",
			ms.UserID, paymentError)
		h.logError("%s: user ID %d - %s", fn, ms.UserID, em)
	}

	// Set the members at address and friends at address in the ordinary member's record.
	setMembersError := h.DB.SetMembersAtAddress(ms.UserID, membersAtAddress)
	if setMembersError != nil {
		h.logError("%s: user ID %d - %v", fn, ms.UserID, setMembersError)
	}

	if h.Conf.EnableGiftaid && ms.Giftaid {
		// Set the giftaid tick box, true or false.
		giftAidError := h.DB.SetGiftaid(ms.UserID, ms.Giftaid)
		if giftAidError != nil {
			h.logError("%s: user ID %d - %v\n", fn, ms.UserID, giftAidError)
		}
	}

	setFriendsError := h.DB.SetFriendsAtAddress(ms.UserID, friendsAtAddress)
	if setFriendsError != nil {
		h.logError("%s: user ID %d - %v", fn, ms.UserID, setFriendsError)
	}

	// If the member is a friend, tick the box.  The user may have been a friend last
	// year and so the record in the DB will be ticked.  The user may not be a friend
	// this year, so always reset the value.
	friendError := h.DB.SetFriendField(
		ms.UserID, ms.Friend)
	if friendError != nil {
		h.logError("%s: user ID %d - %v\n", fn, ms.UserID, friendError)
	}

	// Update the user's donation to society.
	dsError := h.DB.SetDonationToSociety(ms.UserID, ms.DonationToSociety)
	if dsError != nil {
		e := fmt.Errorf("error setting donation to society for %d - %v",
			ms.UserID, dsError)
		h.logError("%s: user ID %d - %v\n", fn, ms.UserID, e)
	}

	// Update the user's donation to museum.
	dmError := h.DB.SetDonationToMuseum(ms.UserID, ms.DonationToMuseum)
	if dmError != nil {
		e := fmt.Errorf("error setting donation to museum for %d - %v",
			ms.UserID, dmError)
		h.logError("%s: user ID %d - %v\n", fn, ms.UserID, e)
	}

	if h.Conf.EnableOtherMemberTypes && ms.AssocUserID > 0 {
		// Associate members are enabled and there is one.  Set the Friend field in the
		// associate member's record.
		setFriendsError := h.DB.SetFriendField(ms.AssocUserID, ms.AssocFriend)
		if setFriendsError != nil {
			e := fmt.Errorf("error setting friend value for %d - %v", ms.AssocUserID, friendError)
			h.logError("%s: user ID %d - %v\n", fn, ms.AssocUserID, e)
		}

		// Set the members at address in the associate member's record.
		setMembersError := h.DB.SetMembersAtAddress(ms.AssocUserID, membersAtAddress)
		if setMembersError != nil {
			h.logError("%s: user ID %d - %v", fn, ms.AssocUserID, setMembersError)
		}

		// Set the members at address in the associate member's record.
		safe := h.DB.SetFriendsAtAddress(ms.AssocUserID, friendsAtAddress)
		if safe != nil {
			h.logError("%s: user ID %d - %v", fn, ms.AssocUserID, setMembersError)
		}
	}
}

// fetchCurrentExtraDetails collects the extra details that the user has already supplied
// (perhaps the last time they paid).
// It's separated out to support unit testing.
func (h *Handler) fetchCurrentExtraDetails(ms *database.MembershipSale) {
	// An existing member is renewing. Fill the extra details (address etc)with
	// the values that the user gave last time. On any error, just set the
	// value to the returned zero value.

	const fn = "fetchCurrentExtraDetails"
	ms.AddressLine1, _ = h.DB.GetAddressLine1(ms.UserID)
	ms.AddressLine2, _ = h.DB.GetAddressLine2(ms.UserID)
	ms.AddressLine3, _ = h.DB.GetAddressLine3(ms.UserID)
	ms.Town, _ = h.DB.GetTown(ms.UserID)
	ms.County, _ = h.DB.GetCounty(ms.UserID)
	ms.Postcode, _ = h.DB.GetPostcode(ms.UserID)
	ms.CountryCode, _ = h.DB.GetCountryCode(ms.UserID)
	ms.Phone, _ = h.DB.GetPhone(ms.UserID)
	ms.Mobile, _ = h.DB.GetMobile(ms.UserID)
	ms.LocationOfInterest, _ = h.DB.GetLocationOfInterest(ms.UserID)
	moi, _ := h.DB.GetMembersOtherInterests(ms.UserID)
	if moi != nil {
		ms.OtherTopicsOfInterest = moi.Interests
	}

	if ms.AssocUserID > 0 {
		// There is an associate member.  They have their own mobile number.
		ms.AssocMobile, _ = h.DB.GetMobile(ms.AssocUserID)
	}

	// Get the interests (if any) that the member selected last time they renewed.
	// These will be pre-selected in the selection list of interests that we about
	// to display.
	interests, ie := h.DB.GetMembersInterests(ms.UserID)
	if ie != nil {
		h.logError("%s: error fetching interests - %v", fn, ie)
	} else {
		if ms.TopicsOfInterest == nil {
			ms.TopicsOfInterest = make(map[int64]interface{})
		}
		for _, interest := range interests {
			ms.TopicsOfInterest[interest.InterestID] = nil
		}
	}
}

// ExtraDetails is the handler for the /extradetails request.  After a successful
// payment, the system displays a page to collect and update extra details such as
// the user's address and phone number.  For a new member the details are initially
// blank.  On a renewal the details are populated with the values from the database.
// Mandatory request parameters: (ordinary member's) account_name, title, first_name,
// last_name.
// Optional request parameters: assoc_account_name, assoc_title, assoc_first_name,
// assoc_last_name.
func (h *Handler) ExtraDetails(w http.ResponseWriter, r *http.Request) {

	fn := "ExtraDetails"
	h.Logger.Info(fn)

	h.DB = database.New(h.DBConfig)
	h.DB.Logger = h.Logger
	connectionError := h.DB.Connect()
	if connectionError != nil {
		h.reportError(w, h.PrePaymentErrorHTML, connectionError)
		return
	}

	// Start a transaction, stored in the database object.
	txError := h.DB.BeginTx()

	if txError != nil {
		h.reportError(w, h.PrePaymentErrorHTML, txError)
		return
	}

	// The transaction should be ether committed or rolled back before this function exits.
	// In case it isn't, set up a deferred rollback now.  We choose a rollback rather than
	// a commit because failure to close the transaction already is almost certainly caused
	// by some sort of catastrophic error.
	//
	// When the function is returning, the transaction may have already been closed, in which
	// case the second rollback may return an error, but that will be ignored.
	defer h.DB.Rollback()
	defer h.DB.Close()

	// We figure out the start and end dates here to support unit testing of the ExtraDetailsHelper.

	// The end date is the end of the calendar year.
	now := time.Now().In(h.TZ)
	paymentYear := database.GetMembershipYear(now)

	err := h.ExtraDetailsHelper(w, r, paymentYear, now)

	if err != nil {
		// Roll back the transaction (which should be done by the defer anyway).
		ce := h.DB.Rollback()
		if ce != nil {
			h.logError("%s: %v", fn, ce)
			w.Write([]byte(h.PrePaymentErrorHTML))
			return
		}
	} else {
		// Commit the changes.
		ce := h.DB.Commit()
		if ce != nil {
			h.logError("%s: %v", fn, ce)
			w.Write([]byte(h.PrePaymentErrorHTML))
			return
		}
	}
}

// ExtraDetailsHelper is a helper for ExtraDetails.  It collects the user's
// extra details (address etc) validatres them and, if valid. commits them
// to the database.
//
// To support unit testing, the helper leaves the caller to commit or roll back
// the transaction, depending on the error return.
func (h *Handler) ExtraDetailsHelper(w http.ResponseWriter, r *http.Request, paymentYear int, now time.Time) error {

	const fn = "ExtraDetailsHelper"

	// The request contains the paying user's account name, the associate member's
	// account if there is one and the extra details.  First check the account names
	// by fetching the data for the accounts.

	// Get the record for the ordinary member.
	accountName := strings.TrimSpace(r.PostFormValue("account_name"))
	if len(accountName) <= 0 {
		// The account name should be carried through the page flow automatically
		// so this is fatal.
		h.logError("%s: account_name not given", fn)
		w.Write([]byte(h.PostPaymentErrorHTML))
		return errors.New("account_name not given")
	}

	user, ue := h.DB.GetUserByLoginName(accountName)
	if ue != nil {
		h.logError("%s: ordinary user account %s does not exist", fn, accountName)
		w.Write([]byte(h.PostPaymentErrorHTML))
		return ue
	}

	ms := database.NewMembershipSale(h.Conf)
	ms.UserID = user.ID
	ms.AccountName = accountName
	ms.MembershipYear = paymentYear

	// Carried via hidden variables in the submitting page.
	ms.Title = strings.TrimSpace(r.PostFormValue("title"))
	ms.FirstName = strings.TrimSpace(r.PostFormValue("first_name"))
	ms.LastName = strings.TrimSpace(r.PostFormValue("last_name"))

	// The associate member is optional.  If one is specified, get their user record. If it turns
	// out to be junk, stop the flow - somebody may be trying to pull a fast one.
	var assocUser *database.User
	assocAccountName := strings.TrimSpace(r.PostFormValue("assoc_account_name"))
	if len(assocAccountName) > 0 {
		var err error
		assocUser, err = h.DB.GetUserByLoginName(assocAccountName)
		if err != nil {
			// The associate's account name is optional but if it's given, it should be
			// carried through the page flow automatically, so this is fatal.
			h.logError("%s: associate user account %s does not exist",
				fn, assocAccountName)
			w.Write([]byte(h.PostPaymentErrorHTML))
			return err
		}

		ms.AssocUserID = assocUser.ID
		ms.AssocAccountName = assocAccountName

		// These are carried via hidden variables in the submitting page.
		ms.AssocTitle = strings.TrimSpace(r.PostFormValue("assoc_title"))
		ms.AssocFirstName = strings.TrimSpace(r.PostFormValue("assoc_first_name"))
		ms.AssocLastName = strings.TrimSpace(r.PostFormValue("assoc_last_name"))
	}

	// Get the incoming form data.

	// Don't assume that the user has filled in the address boxes in order starting at
	// address line 1.  Go through them one by one and store the non-empty lines.
	addrLine := make([]string, 0, 3)
	a := strings.TrimSpace(r.PostFormValue("address_line_1"))
	if len(a) > 0 {
		addrLine = append(addrLine, a)
	}
	a = strings.TrimSpace(r.PostFormValue("address_line_2"))
	if len(a) > 0 {
		addrLine = append(addrLine, a)
	}
	a = strings.TrimSpace(r.PostFormValue("address_line_3"))
	if len(a) > 0 {
		addrLine = append(addrLine, a)
	}

	// Read the address lines back and store them.
	for i, l := range addrLine {
		switch i {
		case 0:
			ms.AddressLine1 = l
		case 1:
			ms.AddressLine2 = l
		case 2:
			ms.AddressLine3 = l
		}
	}

	ms.Town = strings.TrimSpace(r.PostFormValue("town"))
	ms.County = strings.TrimSpace(r.PostFormValue("county"))
	ms.Postcode = strings.TrimSpace(r.PostFormValue("postcode"))
	// CountryCode is three letters, eg "GBR"
	ms.CountryCode = strings.TrimSpace(r.PostFormValue("country_code"))

	// Country code 0 is the heading "Select your country" - ignore.
	if len(ms.CountryCode) > 0 && ms.CountryCode != "0" {
		ct, ce := h.DB.GetCountryByCode(ms.CountryCode)
		if ce != nil {
			h.logError("%s: error getting country code - %v", fn, ce)
			w.Write([]byte(h.PostPaymentErrorHTML))
			return ce
		}
		ms.Country = ct.Name
	}

	ms.LocationOfInterest = strings.TrimSpace(r.PostFormValue("location_of_interest"))
	ms.Phone = strings.TrimSpace(r.PostFormValue("phone"))
	ms.Mobile = strings.TrimSpace(r.PostFormValue("mobile"))
	if assocUser != nil {
		ms.AssocMobile = strings.TrimSpace(r.PostFormValue("assoc_mobile"))
	}
	ms.OtherTopicsOfInterest = strings.TrimSpace(r.PostFormValue("other_topics_of_interest"))

	// This is displayed on the success page.  In test, the result will depend on when the
	// test is run, so don't check it!
	ms.MembershipYear = database.GetMembershipYear(now)

	// The interest selection list is multi-value so there may be many "interest" request
	// parameters.  They should be string versions of IDs from the adm_interests table
	// (id, name).  The supplied interest values should be names from that table.  On valid
	// input we create a row in adm_members_interests (id, userID, interestID) for each
	// interest ID given.  On invalid input we display the page again with the chosen interest
	// values in the selection list marked as selected so that the user doesn't have to choose
	// them again.
	//
	// If the user follows the expeted page flow, the interest IDs should be valid.  If not,
	// something is badly wrong or somebody might be trying to pull a fast one.  Stop the flow.
	iList, ie := h.DB.GetInterests()
	if ie != nil {
		h.logError("%s: error fetching interests - %v", fn, ie)
		w.Write([]byte(h.PostPaymentErrorHTML))
		return ie
	}

	// Get the list of given interests.  This is used on invalid input to annotate the
	// selection list.  We get the parameter values as strings but we check that they are
	// valid int64 values to guard against funny business by the user.
	ms.TopicsOfInterest = make(map[int64]interface{}, 0)
	for _, idStr := range r.PostForm["interest"] {
		var id int64
		n, err := fmt.Sscanf(idStr, "%d", &id)
		if err != nil {
			h.logError("%s: interest ID %s should be an integer", fn, idStr)
			w.Write([]byte(h.PostPaymentErrorHTML))
			return err
		}
		if n == 0 {
			h.logError("%s: failed to convert %s to an integer interest ID", fn, idStr)
			w.Write([]byte(h.PostPaymentErrorHTML))
			return errors.New("")
		}
		ms.TopicsOfInterest[id] = nil
	}

	// This is used on valid input to create the adm_members_interest objects in the database.
	interestList := make(map[int64]database.Interest)
	for _, interest := range iList {
		interestList[interest.ID] = interest
	}

	valid := h.validateExtraDetails(ms)

	if !valid {

		// The data are not valid.  Display the extra details form agaih with error messages.
		// Create the selection list of countries.  By default, the first item in the list
		// is "GBR".  If an existing user is renewing and they specified a country last
		// time, put that at the top of the list and mark it as selected.

		var countriesHTML string

		// An existing user is renewing.  If they specified a country last time,
		// put that at the top of the countries selection list, otherwise put
		// the UK at the top.
		ms.CountryCode, _ = h.DB.GetCountryCode(ms.UserID)
		if len(ms.CountryCode) > 0 {
			// The user has specified a country code.  (Presumably this is a
			// membership renewal and they specified a country last time they paid.)
			// Create a selecton list of countries with that country at the top,
			// pre-selected.
			var ce error
			countriesHTML, ce = h.MakeCountrySelectionListPreSelecting(ms.CountryCode)
			if ce != nil {
				h.logError("%v", ce)
				w.Write([]byte(h.PrePaymentErrorHTML))
				return ce
			}
		} else {
			// The user has not specified a country.  This could be a new
			// membership or a renewal but they have never supplied a country
			// during previous payments.  Make a list of countries with the
			// UK appearing first.
			var ce error
			countriesHTML, ce = h.MakeCountrySelectionListFavouring("GBR")
			if ce != nil {
				h.logError("%v", ce)
				w.Write([]byte(h.PrePaymentErrorHTML))
				return ce
			}
		}

		// Get the HTML to collect the member's interests.
		interestHTML := h.makeInterestSelectionHTML(ms)
		// Create the response page.
		extraDetailsPageTemplateString := successPageTemplateString1 + countriesHTML +
			successPageTemplateString2 + interestHTML + successPageTemplateString3

		// Check and create the template.
		extraDetailsPageTemplate, parseError :=
			template.New("ExtraDetailsPage").Parse(extraDetailsPageTemplateString)
		if parseError != nil {
			h.Logger.Error(parseError.Error())
			w.Write([]byte(h.PrePaymentErrorHTML))
			return parseError
		}

		// Write the response.
		executeError := extraDetailsPageTemplate.Execute(w, ms)
		if executeError != nil {
			h.Logger.Error(executeError.Error())
			w.Write([]byte(h.PrePaymentErrorHTML))
			return executeError
		}

		return executeError
	}

	// The extra details are valid.  Store them in the ordinary user's records.

	saveUserError := h.DB.SaveExtraDetails(ms)
	if saveUserError != nil {
		h.logError("%v", saveUserError)
		w.Write([]byte(h.PrePaymentErrorHTML))
		return saveUserError
	}

	if assocUser != nil {
		// Store the same details (address etc) in the associate user's records.
		msAssocUser := database.NewMembershipSale(h.Conf)
		msAssocUser.UserID = assocUser.ID
		msAssocUser.Phone = ms.Phone
		msAssocUser.AddressLine1 = ms.AddressLine1
		msAssocUser.AddressLine2 = ms.AddressLine2
		msAssocUser.AddressLine3 = ms.AddressLine3
		msAssocUser.Town = ms.Town
		msAssocUser.Postcode = ms.Postcode
		msAssocUser.County = ms.County
		msAssocUser.CountryCode = ms.CountryCode

		saveAssocUserError := h.DB.SaveExtraDetails(msAssocUser)
		if saveAssocUserError != nil {
			h.logError("%v", saveAssocUserError)
			w.Write([]byte(h.PrePaymentErrorHTML))
			return saveAssocUserError
		}
	}

	// Create the completion page.
	completionPageTemplate, parseError :=
		template.New("CompletionPage").Parse(completionPageTemplateString)
	if parseError != nil {
		h.logError("%v", parseError)
		w.Write([]byte(h.PrePaymentErrorHTML))
		return parseError
	}

	// Write the response.
	executeError := completionPageTemplate.Execute(w, ms)
	if executeError != nil {
		h.logError("%v", executeError)
		w.Write([]byte(h.PrePaymentErrorHTML))
		return executeError
	}

	// Success!  The caller should commit the changes.
	return nil
}

// validateExtraDetails is called by the /extradetails handler to validate extra details
// (postal address etc) after a successful sale.
func (h *Handler) validateExtraDetails(msUser *database.MembershipSale) bool {

	// valid starts true and is set false on any validation error.
	valid := true

	if len(msUser.Phone) > 0 {
		msUser.PhoneError = h.ValidatePhoneNumber(msUser.Phone)
		if len(msUser.PhoneError) != 0 {
			// Validation returned an error.
			valid = false
		}
	}

	if len(msUser.Mobile) > 0 {
		msUser.AssocMobileError = h.ValidatePhoneNumber(msUser.Mobile)
		if len(msUser.MobileError) != 0 {
			// Validation returned an error.
			valid = false
		}
	}

	if len(msUser.AssocMobile) > 0 {
		msUser.AssocMobileError = h.ValidatePhoneNumber(msUser.AssocMobile)
		if len(msUser.AssocMobileError) != 0 {
			// Validation returned an error.
			valid = false
		}
	}

	return valid
}

// Completion is the handler for the /completion request.  It displays the final
// page in the page flow.
// Mandatory request parameters: organisation_name, membership_year ("YYYY" format)
func (h *Handler) Completion(w http.ResponseWriter, r *http.Request) {

	fn := "Completion"
	h.Logger.Info(fn)

	membershipYear := database.GetMembershipYear(time.Now().In(h.TZ))

	h.completionHelper(w, r, membershipYear)

	// Success!
	return
}

// completionHelper is a helper for the Completion handler.  It's separated out
// to support unit testing.
func (h *Handler) completionHelper(w http.ResponseWriter, r *http.Request, membershipYear int) {
	fn := "completionHelper"
	h.Logger.Info(fn)

	ms := database.MembershipSale{
		OrganisationName: strings.TrimSpace(r.PostFormValue("organisation_name")),
		MembershipYear:   membershipYear,
	}

	// Create the completion page.
	completionPageTemplate, parseError :=
		template.New("CompletionPage").Parse(completionPageTemplateString)
	if parseError != nil {
		h.logError("%v", parseError)
		w.Write([]byte(h.PrePaymentErrorHTML))
		return
	}

	// Write the response.
	executeError := completionPageTemplate.Execute(w, ms)
	if executeError != nil {
		h.logError("%v", executeError)
		w.Write([]byte(h.PrePaymentErrorHTML))
		return
	}

	// Success!
	return
}

// Cancel is the handler for the /cancel request.  Stripe makes that
// request when the payment is cancelled.
func (h *Handler) Cancel(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(cancelHTML))
}

func (h *Handler) connectToDB() error {
	h.DB = database.New(h.DBConfig)
	h.DB.Logger = h.Logger
	connectError := h.DB.Connect()
	if connectError != nil {
		return connectError
	}

	txError := h.DB.BeginTx()
	if txError != nil {
		return txError
	}

	return nil
}

// makeInterestSectionHTML creates and returns the HTML to collect a member's interests.
// If the adm_interests table doesn't exist or is empty it returns an empty string.
func (h *Handler) makeInterestSelectionHTML(ms *database.MembershipSale) string {
	// The user chooses interests from a multi-choice selection list in the success
	// page and the extra details page.  The interests are taken from the adm_interests
	// table.  The selection list is keyed on the ID of the row in that table.  The
	// given selectedInterests map contain interests that the user selected the last
	// time the selection list was displayed (maybe last year).  They are annotated so
	// that they are selected by default when the list is displayed again.  For example,
	// "Local History" and "Family History" were selected last time so they are marked as
	// "selected" when the page is displayed now:
	//
	// <select name="interest" id="interests" multiple size='5'>
	// 	<option value="1" selected>Local History</option>
	// 	<option value="2">Ancient Buildings</option>
	// 	<option value="3" selected>Family History</option>
	// 	<option value="4">Historic Maps</option>
	// </select>
	//
	// The "other topics of interest" section is just a text box that collects interests
	// that aren't in the ready-made selection list.

	fn := "makeInterestSectionHTML"

	interests, ie := h.DB.GetInterests()
	if ie != nil {
		h.logError("%s:error getting interest list - %v", fn, ie)
		return ""
	}

	if len(interests) <= 0 {
		h.logMessage("%s: adm_interests is empty", fn)
		return ""
	}

	const leader = `
<tr>
    <td>
		Choose any or all of the subjects that you are
		interested in
	</td>
    <td>
	    <select name='interest' id='interests' size='5' multiple>
`
	const trailer = `
		</select>
    </td>
    <td></td>
</tr>
<tr>
	<td style='border: 0'><b>Other Topics of Interest</b></td>
	<td style='border: 0'><input type='text' size='40' 
		name='other_topics_of_interest' value='{{.OtherTopicsOfInterest}}'
	</td>
	<td style="color:red;">{{.OtherTopicsOfInterestError}}</td>
</tr>
`
	var body string
	// Create a selection list of interests, pre-selecting any that the member
	// selected last time.  (That may be when they paid last year.)
	for _, interest := range interests {
		_, ok := ms.TopicsOfInterest[interest.ID]
		if ok {
			// The user selected this item last time the page was diplayed.
			// Mark it as selected.
			body += fmt.Sprintf("            <option selected value='%d'>%s</option>\n",
				interest.ID, interest.Name)
		} else {
			body += fmt.Sprintf("            <option value='%d'>%s</option>\n",
				interest.ID, interest.Name)
		}
	}

	return leader + body + trailer
}

func (hdlr *Handler) reportError(w http.ResponseWriter, errorHTML string, err error) {

	hdlr.Logger.Error(err.Error())
	w.Write([]byte(errorHTML))
}

// Fatal logs a fatal error to the structured log and exits.
func (hdlr *Handler) Fatal(err error) {
	hdlr.Logger.Error(err.Error())
	os.Exit(-1)
}

// DisplayPaymentForm displays the given payment form
func (h *Handler) displaySaleForm(w io.Writer, form *forms.SaleForm) {

	paymentPageTemplate, tpError := template.New("paymentFormTemplate").
		Parse(paymentPageTemplateStr)
	if tpError != nil {
		h.logError("%v", tpError)
		w.Write([]byte(h.PrePaymentErrorHTML))
		return

	}

	executeError := paymentPageTemplate.Execute(w, form)
	if executeError != nil {
		h.logError("%v", executeError)
		w.Write([]byte(h.PrePaymentErrorHTML))
		return
	}
}

// DisplayInitialSaleForm displays an empty payment form
// with the mandatory parameters marked with asterisks.
func (h *Handler) displayInitialSaleForm(w io.Writer, paymentYear int) {

	form := forms.NewSaleForm(h.Conf, paymentYear)
	form.MarkMandatoryFields()

	paymentPageTemplate, tpError := template.New("paymentFormTemplate").
		Parse(paymentPageTemplateStr)
	if tpError != nil {
		h.logError("%v", tpError)
		w.Write([]byte(h.PrePaymentErrorHTML))
		return

	}

	executeError := paymentPageTemplate.Execute(w, form)

	if executeError != nil {
		h.logError("%v", executeError)
		w.Write([]byte(h.PrePaymentErrorHTML))
		return
	}
}

func (h *Handler) logMessage(pattern string, a ...any) {
	str := fmt.Sprintf(pattern, a...)
	h.Logger.Info(str)
}

func (h *Handler) logError(pattern string, a ...any) {
	str := fmt.Sprintf(pattern, a...)
	h.Logger.Error(str)
}

// MakeCountrySelectionListPreSelecting creates a country selection list with the country
// given by the given code (eg "GBR" for the UK) appearing first in the list and
// pre-selected.
func (h *Handler) MakeCountrySelectionListPreSelecting(codeOfPreSelectedCountry string) (string, error) {

	countries, cse := h.DB.GetCountries()
	if cse != nil {
		return "", cse
	}

	const leader = `
<tr>
	<td style='border: 0'><b>Country</b></td>
	<td style='border: 0'>
		<select name='country_code' id='countries' size='5'>
		<option value='0'>Select your country</option>
`
	const trailer = `
		</select>
	</td>
	<td style="color:red;">{{.CountryError}}</td>
</tr>
`

	var first string
	var theRest string
	for _, country := range countries {
		if country.Code == codeOfPreSelectedCountry {
			// We want this country to be first in the list and pre-selected.
			first = fmt.Sprintf("    <option value='%s' selected>%s</option>\n",
				country.Code, country.Name)
		} else {
			// This should not be the first country in the list and it should
			// not be selected.
			theRest += fmt.Sprintf("    <option value='%s'>%s</option>\n",
				country.Code, country.Name)
		}
	}

	return leader + first + theRest + trailer, nil
}

// MakeCountrySelectionListFavouring creates a country selection list with the country
// given by the favoured code (eg "GBR" for the UK) appearing first in the list.
func (h *Handler) MakeCountrySelectionListFavouring(firstCode string) (string, error) {

	favouredCountry, ce := h.DB.GetCountryByCode(firstCode)
	if ce != nil {
		return "", ce
	}
	leaders := []database.Country{*favouredCountry}
	countries, cse := h.DB.GetCountries()
	if cse != nil {
		return "", cse
	}

	html := MakeCountrySelectionList(leaders, countries)

	return html, nil
}

// MakeCountrySelectionList creates an HTML selection list of countries keyed on
// three-letter country codes.  For example, "GBR" is the code for "United
// Kingdom of Great Britain and Northern Ireland".  The leader list contains
// countries that should be displayed first.  The countries list contains the
// complete list
func MakeCountrySelectionList(leaders, countries []database.Country) string {
	// lookUp is a look up table containing the codes of the leaders.
	lookUp := make(map[string]interface{}, len(leaders))
	for _, c := range leaders {
		lookUp[c.Code] = nil
	}
	listHTML := "<select name='country_code' id='countries' size='5'>\n"
	// Add the leader countries to the selection list first.
	for _, country := range leaders {
		listHTML += fmt.Sprintf("    <option value='%s'>%s</option>\n",
			country.Code, country.Name)
	}

	// Now add the rest, missing out any already added.
	for _, country := range countries {
		_, alreadyAdded := lookUp[country.Code]
		if !alreadyAdded {
			listHTML += fmt.Sprintf("    <option value='%s'>%s</option>\n",
				country.Code, country.Name)
		}
	}
	listHTML += `</select>`

	return listHTML
}

// Validation error messages - factored out to support unit testing.
const firstNameErrorMessage = "You must fill in the first name"
const lastNameErrorMessage = "You must fill in the last name"
const emailErrorMessage = "You must fill in the email address"
const assocFirstNameErrorMessage = "If you fill in anything in this section, you must fill in the first name"
const assocLastNameErrorMessage = "If you fill in anything in this section, you must fill in the last name"
const invalidNumber = "must be a number"
const negativeNumber = "must be a 0 or greater"

// ValidateSaleForm takes the form parameters as arguments.  It returns true
// and all empty strings if the form is valid, false and the error messages set
// if it's invalid.
func ValidateSaleForm(sf *forms.SaleForm) bool {

	// form.Valid is set false if any of the form data is invalid.
	sf.Valid = true

	// The "ToPay" fields should not be set before validation - they are set after.
	sf.FriendFeeToPay = 0.0
	sf.AssocFeeToPay = 0.0
	sf.AssocFriendFeeToPay = 0.0

	if len(sf.Title) == 0 &&
		len(sf.FirstName) == 0 &&
		len(sf.LastName) == 0 &&
		len(sf.Email) == 0 &&
		len(sf.FriendInput) == 0 &&
		len(sf.AssocTitle) == 0 &&
		len(sf.AssocFirstName) == 0 &&
		len(sf.AssocLastName) == 0 &&
		len(sf.AssocEmail) == 0 &&
		len(sf.AssocFriendInput) == 0 &&
		len(sf.DonationToSocietyInput) == 0 &&
		len(sf.DonationToMuseumInput) == 0 &&
		len(sf.GiftaidInput) == 0 {

		// On the first call the form is empty.  Mark the mandatory fields.
		// Return false and the handler will display the form
		// again, with the marks.
		sf.MarkMandatoryFields()
		sf.Valid = false
		return sf.Valid
	}

	sf.Title = strings.TrimSpace(sf.Title)
	sf.FirstName = strings.TrimSpace(sf.FirstName)
	sf.LastName = strings.TrimSpace(sf.LastName)
	sf.Email = strings.TrimSpace(sf.Email)
	sf.FriendInput = strings.TrimSpace(sf.FriendInput)
	sf.DonationToSocietyInput = strings.TrimSpace(sf.DonationToSocietyInput)
	sf.DonationToMuseumInput = strings.TrimSpace(sf.DonationToMuseumInput)
	sf.GiftaidInput = strings.TrimSpace(sf.GiftaidInput)

	sf.AssocTitle = strings.TrimSpace(sf.AssocTitle)
	sf.AssocFirstName = strings.TrimSpace(sf.AssocFirstName)
	sf.AssocLastName = strings.TrimSpace(sf.AssocLastName)
	sf.AssocEmail = strings.TrimSpace(sf.AssocEmail)
	sf.AssocFriendInput = strings.TrimSpace(sf.AssocFriendInput)

	sf.Friend, sf.FriendInput, sf.FriendOutput = getTickBox(sf.FriendInput)
	sf.AssocFriend, sf.AssocFriendInput, sf.AssocFriendOutput = getTickBox(sf.AssocFriendInput)
	sf.Giftaid, sf.GiftaidInput, sf.GiftaidOutput = getTickBox(sf.GiftaidInput)

	if len(sf.FirstName) == 0 {
		sf.FirstNameErrorMessage = firstNameErrorMessage
		sf.Valid = false
	}

	if len(sf.LastName) == 0 {
		sf.LastNameErrorMessage = lastNameErrorMessage
		sf.Valid = false
	}

	if len(sf.Email) == 0 {
		sf.EmailErrorMessage = emailErrorMessage
		sf.Valid = false
	}

	// The associate fields are optional but if you fill in any of them, you must
	// fill in the first and last name.  Filling in the AssociateFriendInput means
	// ticking it, which sets it to "on".
	if len(sf.AssocFirstName) == 0 || len(sf.AssocLastName) == 0 {
		if len(sf.AssocTitle) > 0 ||
			len(sf.AssocEmail) > 0 ||
			sf.AssocFriendInput == "on" {

			sf.Valid = false

			if len(sf.AssocFirstName) == 0 {
				sf.AssocFirstNameErrorMessage = assocFirstNameErrorMessage
			}
			if len(sf.AssocLastName) == 0 {
				sf.AssocLastNameErrorMessage = assocLastNameErrorMessage
			}
		}
	}

	// The mandatory parameters are all present.  Now check the contents of number fields.

	// If donation values are submitted, they must be numbers and not
	// negative.

	if len(sf.DonationToSocietyInput) > 0 {

		// The donation must be a number, zero or greater.
		errorMessage, dts := checkNonNegativeNumber(sf.DonationToSocietyInput)
		if len(errorMessage) > 0 || dts < 0.0 {
			sf.DonationToSocietyErrorMessage = errorMessage
			sf.Valid = false
		} else {
			sf.DonationToSociety = dts
		}
	}

	if len(sf.DonationToMuseumInput) > 0 {
		errorMessage, dtm := checkNonNegativeNumber(sf.DonationToMuseumInput)
		if len(errorMessage) > 0 || dtm < 0.0 {
			sf.DonationToMuseumErrorMessage = errorMessage
			sf.Valid = false
		} else {
			sf.DonationToMuseum = dtm
		}
	}

	return sf.Valid
}

// Error message for an illegalphone number.  Factored out for unit testing.
const IllegalPhoneNumber = "phone number must start with '+' or '0' and then must be all digits or spaces"

// ValidatePhoneNumber checks a phone number.  It should start with "+"
// or "0" and this should be followed by digits or spaces. For example
// "+44 1234 567890" or "01234 567890" or "020 7123 4567".  If the number
// is invalid, an error message is returned.
func (h *Handler) ValidatePhoneNumber(phoneNumber string) string {
	if !h.PhoneRegexp.Match([]byte(phoneNumber)) {
		return IllegalPhoneNumber
	}
	return ""
}

// checkNonNegativeNumber checks a donation value - must be a valid float
// and not negative.  Returns an empty error message and the donation
// as a float64 OR an error message and 0.0.
func checkNonNegativeNumber(str string) (string, float64) {
	var v float64
	if len(str) > 0 {

		scannedItems, scanError := fmt.Sscanf(str, "%f", &v)

		if scanError != nil {
			return invalidNumber, 0.0
		}

		if scannedItems < 1 {
			return invalidNumber, 0.0
		}

		// The number must not be negative!
		if v < 0 {
			return negativeNumber, 0.0
		}
	}

	// Success!
	return "", v
}

// usersExist checks if the users in the payment form are already in the system -
// meaning that this is a membership renewal. It looks up the ordinary user in
// the database with the name and/or email address given in the form to check that
// they exists.  If the details of an associate member are given, it checks that
// user too.  It returns the user IDs if both user(s) exist - (42,43,nil) or just
// the ID of the ordinary user if there is no associate - (42,0, nil).  If there
// is no match, it returns (0,0,nil).  If there is an error it returns (0,0,error).
func usersExist(ms *database.MembershipSale, db *database.Database) (int64, int64, error) {

	userID, userIDError := db.GetUserIDofMember(ms.FirstName, ms.LastName, ms.Email)
	if userIDError != nil {
		return 0, 0, userIDError
	}

	var assocUserID int64

	if len(ms.AssocFirstName) > 0 {
		var assocUserIDError error
		assocUserID, assocUserIDError =
			db.GetUserIDofMember(ms.AssocFirstName, ms.AssocLastName, ms.AssocEmail)
		if assocUserIDError != nil {
			return 0, 0, assocUserIDError
		}
	}
	return userID, assocUserID, nil
}

// getTickBox returns true and "checked" if the tickbox is ticked ("on"),
// false and "unchecked" otherwise.
func getTickBox(value string) (bool, string, string) {
	switch {
	case len(value) == 0:
		return false, "off", "unchecked"
	case value == "on":
		return true, "on", "checked"
	default:
		return false, "off", "unchecked"
	}
}
