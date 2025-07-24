package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)
type User struct {
	ID       primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name     string             `bson:"name" json:"name"`
	Email    string             `bson:"email" json:"email"`
	Password string             `bson:"password,omitempty" json:"-"`
	ProfileImage string             `bson:"profile_image,omitempty"` // ✅ New field

}
type Parcel struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID         primitive.ObjectID `bson:"user_id,omitempty" json:"user_id"`
	PickupAddress  string             `bson:"pickup_address" json:"pickup_address"`
	DropAddress    string             `bson:"drop_address" json:"drop_address"`
	ReceiverName   string             `bson:"receiver_name" json:"receiver_name"`
	ReceiverPhone  string             `bson:"receiver_phone" json:"receiver_phone"`
	Status         string             `bson:"status" json:"status binding:required"`
	CreatedAt      time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt      time.Time          `bson:"updated_at" json:"updated_at"`
}
type ParcelOTP struct {
    ParcelID    primitive.ObjectID `bson:"parcel_id"`
    OTP         string             `bson:"otp"`
    ExpiresAt   time.Time          `bson:"expires_at"`
}
