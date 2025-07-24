package partner

import (
	"context"
	"errors"
	"go_pickup/auth"
	"go_pickup/models"
	"time"
    "go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

func RegisterDeliveryPartner(partner models.DeliveryPartner) error {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    collection := services.GetAgentCollection()
    
    // Check if user already exists
    var existing models.User
    err := collection.FindOne(ctx, bson.M{"email": partner.Email}).Decode(&existing)
    if err == nil {
        return errors.New("user already exists")
    }

    hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(partner.Password), bcrypt.DefaultCost)
    partner.Password = string(hashedPassword)
    partner.ID = primitive.NewObjectID()

    _, err = collection.InsertOne(ctx, partner)
    if err != nil {
        return err
    }
    return nil
}