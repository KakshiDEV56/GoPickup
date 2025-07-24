package twilio

import (
	"github.com/twilio/twilio-go"
	openapi "github.com/twilio/twilio-go/rest/api/v2010"
	"log"
	"os"
)

var client *twilio.RestClient

func Init() {
	client = twilio.NewRestClientWithParams(twilio.ClientParams{
		Username: os.Getenv("TWILIO_ACCOUNT_SID"),
		Password: os.Getenv("TWILIO_AUTH_TOKEN"),
	})
}

func SendSMS(to string, message string) {
	params := &openapi.CreateMessageParams{}
	params.SetTo(to)                      // e.g., "+919876543210"
	params.SetFrom(os.Getenv("TWILIO_PHONE_NUMBER"))
	params.SetBody(message)

	_, err := client.Api.CreateMessage(params)
	if err != nil {
		log.Printf("Twilio SMS error: %v\n", err)
	} else {
		log.Println("Twilio SMS sent successfully!")
	}
}
