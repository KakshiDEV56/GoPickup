package partner

import (
	"context"
	"errors"
	"go_pickup/auth"
	"go_pickup/models"
  "time"
	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/crypto/bcrypt"
)
func LoginDeliveryPartner(email, password string) (*models.DeliveryPartner, error) {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    collection := services.GetAgentCollection()

    var partner models.DeliveryPartner
    err := collection.FindOne(ctx, bson.M{"email": email}).Decode(&partner)
    if err != nil {
        return nil, errors.New("delivery partner not found")
    }

    err = bcrypt.CompareHashAndPassword([]byte(partner.Password), []byte(password))
    if err != nil {
        return nil, errors.New("invalid credentials")
    }

    return &partner, nil
}