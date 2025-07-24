package email
// using SendGrid's Go Library
// https://github.com/sendgrid/sendgrid-g

import (
	"os"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

func SendEmail(toEmail, subject, plainTextContent string) error {
	from := mail.NewEmail("GoPickup", os.Getenv("SENDER_EMAIL"))
	to := mail.NewEmail("", toEmail)
	message := mail.NewSingleEmail(from, subject, to, plainTextContent, "")
	client := sendgrid.NewSendClient(os.Getenv("SENDGRID_API_KEY"))
   // fmt.Println("SENDGRID_API_KEY:", os.Getenv("SENDGRID_API_KEY"))
//fmt.Println("SENDER_EMAIL:", os.Getenv("SENDER_EMAIL"))
	_, err := client.Send(message)
	if err != nil {
		return err
	}

	// Log status code and body
	//fmt.Printf("SendGrid response code: %d\n", response.StatusCode)
	//fmt.Printf("SendGrid response body: %s\n", response.Body)


	return nil
}
