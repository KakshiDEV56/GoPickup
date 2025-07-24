package models

import (
	"time"
    "go.mongodb.org/mongo-driver/bson/primitive"
)

type DeliveryPartner struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name      string             `bson:"name" json:"name"`
	Email     string             `bson:"email" json:"email"`
	License   string             `bson:"license" json:"license"`
	Password  string             `bson:"password,omitempty" json:"password"`
	Phone     string             `bson:"phone" json:"phone"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at" json:"updated_at"`
	ProfileImage string             `bson:"profile_image,omitempty"` // ✅ New field

}
// i need to enter the id of the parcel created that the user has implemented
type LocationUpdate struct {
	DeliveryPartner
	ParcelID        primitive.ObjectID `bson:"parcel_id" json:"parcel_id"`
	Latitude float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Timestamp time.Time `json:"timestamp"`
	Status string `josn:"status,omitempty"`
}
