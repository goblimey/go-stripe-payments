package handler

// HTML views, separated out for clarity.

// paymentPageTemplateStr defines the initial payment form that collects
// the data for the sale.  Data is taken from a SaleForm object.
const paymentPageTemplateStr = `
<html>
<head>
    <title>Membership Payment System</title>
</head>
	<body style='font-size: 100%'>
		<h2>{{.OrganisationName}}</h2>

		<h3>Membership Year {{.MembershipYear}}</h3>

		<span style="color:red;">{{.GeneralErrorMessage}}</span>
		<p>
			To become a member or renew your membership
			using a credit or debit card,
			please fill in the form below and press the Submit button.
		</p>
	{{if .EnableOtherMemberTypes}}
		<p>
			If you are also paying for an Associate Member
			(a second member at the same address)
			please supply their details too,
			otherwise leave those boxes blank.
		</p>
	{{end}}
		<p>
			Our fees this year are:
		<ul>
			<li>Ordinary member: {{.OrdinaryMemberFeeForDisplay}}</li>
		{{if .EnableOtherMemberTypes}}
			<li>Associate member at the same address: {{.AssocFeeForDisplay}}</li>
			<li>Friend of the Leatherhead museum: {{.FriendFeeForDisplay}}</li>
		{{end}}
		</ul>
		</p>
		<p>
			&nbsp;
		</p>
		<form action="/displayPaymentForm" method="POST">	
			<table style='font-size: 100%'>
				<tr>
					<td style='border: 0'>First Name:</td>
					<td style='border: 0'><input type='text' size='40' name='first_name' value='{{.FirstName}}'></td>
					<td style='border: 0'><span style="color:red;">{{.FirstNameErrorMessage}}</span></td>
				</tr>

				<tr>
					<td style='border: 0'>Last Name:</td>
					<td style='border: 0'><input type='text' size='40' name='last_name' value='{{.LastName}}'></td>
					<td style='border: 0'><span style="color:red;">{{.LastNameErrorMessage}}</span></td>
				</tr>

				<tr>
					<td style='border: 0'>Email Address:</td>
					<td style='border: 0'><input type='text' size='40' name='email' value='{{.Email}}'></td>
					<td style='border: 0'><span style="color:red;">{{.EmailErrorMessage}}</span></td>
				</tr>
			{{if .EnableOtherMemberTypes}}
				<tr>
					<td style='border: 0'>Friend of the Museum:</td>
					<td style='border: 0; '>
						<input style='transform: scale(1.5);' type='checkbox' name='friend' {{.FriendOutput}} />
					</td>
					<td style='border: 0'>&nbsp;</td>
				</tr>
			{{end}}

				<tr>
					<td style='border: 0'>Donation:</td>
					<td style='border: 0'><input type='text' size='40' name='donation_to_society' value='{{.DonationToSocietyForDisplay}}'></td>
					<td style='border: 0;'><span style="color:red;">{{.DonationToSocietyErrorMessage}}</span></td>
				</tr>

			{{if .EnableOtherMemberTypes}}
				<tr>
					<td style='border: 0'>Donation to the Museum:</td>
					<td style='border: 0'><input type='text' size='40' name='donation_to_museum' value='{{.DonationToMuseumForDisplay}}'></td>
					<td style='border: 0'><span style="color:red;">{{.DonationToMuseumErrorMessage}}</span></td>
				</tr>
			{{end}}

			{{if .EnableGiftaid}}
				<tr>
					<td style='border: 0'>Gift Aid:</td>
					<td style='border: 0 '>
						<input style='transform: scale(1.5);' type='checkbox' name='Gif Aid' {{.GiftaidOutput}} />
					</td>
					<td style='border: 0'>&nbsp;</td>
				</tr>
				<tr>
					<td style='border: 0' colspan='3'>&nbsp;</td>
				</tr>
				<tr>
					<td style='border: 0' colspan='3'>
						Tick the Gift Aid box if you are currently a UK tax payer and 
						consent to Gift Aid.
						If you pay less income tax and/or capital gains tax 
						than the amount of Gift Aid paid on all your donations, 
						you are liable to pay the difference to HMRC.
					</td>
				</tr>
			{{end}}
				<tr>
					<td style='border: 0' colspan='3'>&nbsp;</td>
				</tr>
				
				<tr>
					<td style='border: 0' colspan='3'>&nbsp;</td>
				</tr>

			{{if .EnableOtherMemberTypes}}
				<tr>
					<td style='border: 0' colspan='3'>
						If there are two members at your address, 
						fill in the other member's details below:
					</td>
				</tr>
				<tr>
					<td style='border: 0'>Associate First Name:</td>
					<td style='border: 0'><input type='text' size='40' name='assoc_first_name' value='{{.AssocFirstName}}'></td>
					<td style='border: 0'><span style="color:red;">{{.AssocFirstNameErrorMessage}}</span></td>
				</tr>

				<tr>
					<td style='border: 0'>Associate Last Name:</td>
					<td style='border: 0'><input type='text' size='40' name='assoc_last_name' value='{{.AssocLastName}}'></td>
					<td style='border: 0'><span style="color:red;">{{.AssocLastNameErrorMessage}}</span></td>
				</tr>

				<tr>
					<td style='border: 0'>Email Address (optional):</td>
					<td style='border: 0'><input type='text' size='40' name='assoc_email' value='{{.AssocEmail}}'></td>
					<td style='border: 0'>&nbsp;</td>
				</tr>
				<tr>
					<td style='border: 0'>Friend of the Museum:</td>
					<td style='border: 0'>
						<input style='transform: scale(1.5);' type='checkbox' name='assoc_friend' {{.AssocFriendOutput}} />
					</td>
					<td style='border: 0'>&nbsp;</td>
				</tr>

				<tr>
					<td style='border: 0' colspan='3'>&nbsp;</td>
				</tr>
			{{end}}

			</table>
			<input type="submit" value="Submit">
		</form>
	</body>
</html>
`

// paymentConfirmationPageTemplate defines the payment confirmation
// page.  Data is taken from a SaleForm object.
const paymentConfirmationPageTemplate = `
<html>
    <head><title>payment confirmation</title></head>
	<body style='font-size: 100%'>
		<h2>{{.OrganisationName}}</h2>
		<h3>Membership Fee for {{.MembershipYear}}</h3>
		<p>
			If you are happy with the total,
			please press the submit button.
			You will be transferred to the Stripe payment system
			to make the payment.
		</p>
		<form action="/checkout" method="POST">

			<table>
				<tr>
					<td style='border: 0'>ordinary membership</td>
					<td style='border: 0'>{{.OrdinaryMemberFeeForDisplay}}</td>
				</tr>
			{{if .Friend}}
				<tr>
					<td style='border: 0'>friend of the museum</td>
					<td style='border: 0'>{{.FriendFeeForDisplay}}</td>
				</tr>
			{{end}}
			{{if ne .AssocFirstName ""}}
				<tr>
					<td style='border: 0'>associate member</td>
					<td style='border: 0'>{{.AssocFeeForDisplay}}</td>
				</tr>
			{{if .AssocFriend}}
				<tr>
					<td style='border: 0'>associate is a friend of the museum</td>
					<td style='border: 0'>{{.FriendFeeForDisplay}}</td>
				</tr>
			{{end}}
			{{end}}

			{{if gt .DonationToSociety 0.0}}
			<tr>
				<td style='border: 0'>donation to the Society</td>
				<td style='border: 0'>{{.DonationToSocietyForDisplay}}</td>
			</tr>
			{{end}}

			{{if gt .DonationToMuseum 0.0}}
			<tr>
				<td style='border: 0'>donation to the museum</td>
				<td style='border: 0'>{{.DonationToMuseumForDisplay}}</td>
			</tr>
			{{end}}

			<tr>
				<td style='border: 0'><b>Total</b></td>
				<td style='border: 0'>{{.TotalForDisplay}}</td>
			</tr>

			</table>
			<input type='hidden' name='first_name' value={{.FirstName}}>
			<input type='hidden' name='last_name' value={{.LastName}}>
			<input type='hidden' name='email' value={{.Email}}>
		{{if .Friend}}
			<input type='hidden' name='friend' value='on'>
		{{end}}
		{{if ne .AssocFirstName ""}}
			<input type='hidden' name='assoc_first_name' value={{.AssocFirstName}}>
			<input type='hidden' name='assoc_last_name' value={{.AssocLastName}}>
			<input type='hidden' name='assoc_email' value={{.AssocEmail}}>
		{{if .AssocFriend}}
			<input type='hidden' name='assoc_friend' value='on'>
		{{end}}
		{{end}}
		{{if .Giftaid}}
			<input type='hidden' name='giftaid' value='on'>
		{{end}}
		{{if gt .DonationToSociety 0.0}}
			<input type='hidden' name='donation_to_society' value='{{.DonationToSociety}}'>
		{{end}}
		{{if gt .DonationToMuseum 0.0}}
			<input type='hidden' name='donation_to_museum' value='{{.DonationToMuseum}}'>
		{{end}}
			<input type="submit" value="Submit">
		</form>
	</body>
`

// successPageTemplateStr defines the page shown on a successful sale.
// Data is taken from a MembershipSale object.
const successPageTemplateStr = `
<html>
	<head><title>Payment Successful</title></head>
    <body style='font-size: 100%'>
	<h2>{{.OrganisationName}}</h2>
        <p>
			Thank you for your payment.
			You are now a member until the end of {{.MembershipYear}}.
		</p>
		<p>
			<table>
				<tr>
					<td style='border: 0'>ordinary membership</td>
					<td style='border: 0'>{{.OrdinaryMemberFeeForDisplay}}</td>
				</tr>
			{{if .Friend}}
				<tr>
					<td style='border: 0'>friend of the museum</td>
					<td style='border: 0'>{{.FriendFeeForDisplay}}</td>
				</tr>
			{{end}}
			{{if ne .AssocFirstName ""}}
				<tr>
					<td style='border: 0'>associate member</td>
					<td style='border: 0'>{{.AssocFeeForDisplay}}</td>
				</tr>
			{{if .AssocFriend}}
				<tr>
					<td style='border: 0'>associate is friend of the museum</td>
					<td style='border: 0'>{{.AssocFriendFeeForDisplay}}</td>
				</tr>
			{{end}}
			{{end}}

			{{if gt .DonationToMuseum 0.0}}
			<tr>
				<td style='border: 0'>donation to the Society</td>
				<td style='border: 0'>{{.DonationToSocietyForDisplay}}</td>
			</tr>
			{{end}}

			{{if gt .DonationToMuseum 0.0}}
			<tr>
				<td style='border: 0'>donation to the museum</td>
				<td style='border: 0'>{{.DonationToMuseumForDisplay}}</td>
			</tr>
			{{end}}

			<tr>
				<td style='border: 0'><b>Total</b></td>
				<td style='border: 0'>{{.TotalForDisplay}}</td>
			</tr>

			</table>
		</p>
		<p>
		    If you have any questions, please email
		    <a href="mailto:{{.EmailAddressForQuestions}}">
			    {{.EmailAddressForQuestions}}
			</a>.
		</p>
    </body>
</html>
`

// cancelHTML defines the cancel page, called when the payment is cancelled
// on the Stripe system.  Not sure under what circumstances this happens or
// how to provoke it in the test environment.
const cancelHTML = `
<html>
  <head><title>cancelled</title></head>
  <body style='font-size: 100%'>
    <h1>Payment cancelled</h1>
  </body>
</html>
`

// prePaymentErrorHTMLPattern defines the default error message page before the
// customer has paid (not as bad as a failure after).  It contains configurable
// text but it may be used when all sorts of things are going wrong, including
// the template system itsef, so we make the rendering of the page as simple as
// possible and not done via the template system.  We just expand the pattern
// using Sprintf (which means that "100%" must be presented as "100%%").
//
// None of the configurable data is submitted by the user so there is no chance
// of an injection attack.
const prePaymentErrorHTMLPattern = `
<html>
    <head><title>error</title></head>
    <body style='font-size: 100%%'>
		<p>
			Internal error.  Please try again.
			If the error persists, please email
			<a href="mailto:%s">
				%s
			</a>
		</p>
    </body>
</html>
`

// postPaymentErrorHTMLPattern defines the default error message page after the
// customer has paid.  It contains configurable text but it may be used when
// all sorts of things are going wrong, including the template system itsef, so
// we make the rendering of the page as simple as possible and not done via the
// template system.  We just expand the pattern using Sprintf (which means that
// "100%" must be presented as "100%%").  None of the configurable data is
// submitted by the user so there is no chance of an injection attack.
//
// The customer is now due a refund so the given email address should be somebody
// who can arrange that, eg the treasurer.
const postPaymentErrorHTMLPattern = `
<html>
    <head><title>error</title></head>
    <body style='font-size: 100%%'>
		<p>
			Something went wrong after you had paid.
			Please
			<a href="mailto:%s">
				%s
			</a>
		</p>
    </body>
</html>
`
